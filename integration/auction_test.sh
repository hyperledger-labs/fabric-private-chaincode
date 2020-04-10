#!/bin/bash

# Copyright 2020 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

SCRIPTDIR="$(dirname $(readlink --canonicalize ${BASH_SOURCE}))"
FPC_TOP_DIR="${SCRIPTDIR}/.."
FABRIC_SCRIPTDIR="${FPC_TOP_DIR}/fabric/bin/"

: ${FABRIC_CFG_PATH:="${SCRIPTDIR}/config"}

. ${FABRIC_SCRIPTDIR}/lib/common_utils.sh
. ${FABRIC_SCRIPTDIR}/lib/common_ledger.sh

CC_ID=auction_test

#this is the path that will be used for the docker build of the chaincode enclave
ENCLAVE_SO_PATH=examples/auction/_build/lib/

CC_VERS=0
num_rounds=3
num_clients=10
FAILURES=0

auction_test() {
    # install, init, and register (auction) chaincode
    try ${PEER_CMD} chaincode install -l fpc-c -n ${CC_ID} -v ${CC_VERS} -p ${ENCLAVE_SO_PATH}
    sleep 3

    try ${PEER_CMD} chaincode instantiate -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -v ${CC_VERS} -c '{"Args":[]}' -V ecc-vscc
    sleep 3

    # Scenario 1
    becho ">>>> Close and evaluate non existing auction. Response should be AUCTION_NOT_EXISTING"
    try_r ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Function":"init", "Args": ["MyAuctionHouse"]}' --waitForEvent
    check_result "OK"
    try_r ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Function":"close", "Args": ["MyAuction"]}' --waitForEvent
    check_result "AUCTION_NOT_EXISTING"
    try_r ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Function":"eval", "Args": ["MyAuction0"]}' # Don't do --waitForEvent, so potentially there is some parallelism here ..
    check_result "AUCTION_NOT_EXISTING"

    # Scenario 2
    becho ">>>> Create an auction. Response should be OK"
    try_r ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Function":"init", "Args": ["MyAuctionHouse"]}' --waitForEvent
    check_result "OK"
    try_r ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Function":"create", "Args": ["MyAuction1"]}' --waitForEvent
    check_result "OK"
    becho ">>>> Create two equivalent bids. Response should be OK"
    try_r ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Function":"submit", "Args": ["MyAuction1", "JohnnyCash0", "2"]}' # Don't do --waitForEvent, so potentially there is some parallelism here ..
    check_result "OK"
    try_r ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Function":"submit", "Args": ["MyAuction1", "JohnnyCash1", "2"]}' # Don't do --waitForEvent, so potentially there is some parallelism here ..
    check_result "OK"
    becho ">>>> Close auction. Response should be OK"
    try_r ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Function":"close", "Args": ["MyAuction1"]}' --waitForEvent
    check_result "OK"
    becho ">>>> Submit a bid on a closed auction. Response should be AUCTION_ALREADY_CLOSED"
    try_r ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Function":"submit", "Args": ["MyAuction1", "JohnnyCash2", "2"]}' # Don't do --waitForEvent, so potentially there is some parallelism here ..
    check_result "AUCTION_ALREADY_CLOSED";
    becho ">>>> Evaluate auction. Response should be DRAW"
    try_r ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Function":"eval", "Args": ["MyAuction1"]}' # Don't do --waitForEvent, so potentially there is some parallelism here ..
    check_result "DRAW"

    # Scenario 3
    becho ">>>> Create an auction. Response should be OK"
    try_r ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Function":"init", "Args": ["MyAuctionHouse"]}' --waitForEvent
    check_result "OK"
    try_r ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Function":"create", "Args": ["MyAuction2"]}' --waitForEvent
    check_result "OK"
    for (( i=0; i<=$num_rounds; i++ ))
    do
        becho ">>>> Submit unique bid. Response should be OK"
        b="$(($i%$num_clients))"
        try_r ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Function":"submit", "Args": ["MyAuction2", "JohnnyCash'$b'", "'$b'"]}' # Don't do --waitForEvent, so potentially there is some parallelism here ..
        check_result "OK"
    done
    becho ">>>> Close auction. Response should be OK"
    try_r ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Function":"close", "Args": ["MyAuction2"]}' --waitForEvent
    check_result "OK"
    becho ">>>> Evaluate auction. Auction Result should be printed out"
    try_r ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Function":"eval", "Args": ["MyAuction2"]}' # Don't do --waitForEvent, so potentially there is some parallelism here ..
    check_result '{"bidder":"JohnnyCash3","value":3}'

    # Scenario 4
    becho ">>>> Create a new auction. Response should be OK"
    try_r ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Function":"init", "Args": ["MyAuctionHouse"]}' --waitForEvent
    check_result "OK"
    try_r ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Function":"create", "Args": ["MyAuction3"]}' --waitForEvent
    check_result "OK"
    becho  ">>>> Create a duplicate auction. Response should be AUCTION_ALREADY_EXISTING"
    try_r ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Function":"create", "Args": ["MyAuction3"]}' --waitForEvent
    check_result "AUCTION_ALREADY_EXISTING"
    becho ">>>> Close auction and evaluate. Response should be OK"
    try_r ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Function":"close", "Args": ["MyAuction3"]}' --waitForEvent
    check_result "OK"
    becho ">>>> Close an already closed auction. Response should be AUCTION_ALREADY_CLOSED"
    try_r ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Function":"close", "Args": ["MyAuction3"]}' --waitForEvent
    check_result "AUCTION_ALREADY_CLOSED"
    becho ">>>> Evaluate auction. Response should be NO_BIDS"
    try_r ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Function":"eval", "Args": ["MyAuction3"]}' # Don't do --waitForEvent, so potentially there is some parallelism here ..
    check_result "NO_BIDS"

    # Code below is used to test bug in issue #42
    becho ">>>> Create a new auction. Response should be OK"
    try_r ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Function":"init", "Args": ["MyAuctionHouse"]}' --waitForEvent
    check_result "OK"
    try_r ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Function":"create", "Args": ["MyAuction4"]}' --waitForEvent
    check_result "OK"
}

# 1. prepare
para
say "Preparing Auction Test ..."
# - clean up relevant docker images
docker_clean ${ERCC_ID}
docker_clean ${CC_ID}

trap ledger_shutdown EXIT

para
say "Run auction test"

say "- setup ledger"
ledger_init

say "- auction test"
auction_test

say "- shutdown ledger"
ledger_shutdown

para
if [[ "$FAILURES" == 0 ]]; then
    yell "Auction test PASSED"
else
    yell "Auction test had ${FAILURES} failures"
    exit 1
fi
exit 0
