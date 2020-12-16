#!/bin/bash

# Copyright 2020 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

SCRIPTDIR="$(dirname $(readlink --canonicalize ${BASH_SOURCE}))"
FPC_PATH="${SCRIPTDIR}/.."
FABRIC_SCRIPTDIR="${FPC_PATH}/fabric/bin/"

: ${FABRIC_CFG_PATH:="${SCRIPTDIR}/config"}

. ${FABRIC_SCRIPTDIR}/lib/common_utils.sh
. ${FABRIC_SCRIPTDIR}/lib/common_ledger.sh

CC_ID=echo_test
CC_PATH=${FPC_PATH}/examples/echo/_build/lib/
CC_LANG=fpc-c
CC_VER="$(cat ${CC_PATH}/mrenclave)"
CC_SEQ="1"
CC_EP="OR('SampleOrg.member')" # note that we use .member as NodeOUs is disabled with the crypto material used in the integration tests.

num_rounds=10
FAILURES=0

echo_test() {
    PKG=/tmp/${CC_ID}.tar.gz

    try ${PEER_CMD} lifecycle chaincode package --lang ${CC_LANG} --label ${CC_ID} --path ${CC_PATH} ${PKG}
    try ${PEER_CMD} lifecycle chaincode install ${PKG}

    PKG_ID=$(${PEER_CMD} lifecycle chaincode queryinstalled | awk "/Package ID: ${CC_ID}/{print}" | sed -n 's/^Package ID: //; s/, Label:.*$//;p')

    # first call negated as it fails due to specification of validation plugin
    try_fail ${PEER_CMD} lifecycle chaincode approveformyorg -o ${ORDERER_ADDR} -C ${CHAN_ID} --package-id ${PKG_ID} --name ${CC_ID} --version ${CC_VER} --sequence ${CC_SEQ} --signature-policy ${CC_EP} -E mock-escc
    try ${PEER_CMD} lifecycle chaincode approveformyorg -o ${ORDERER_ADDR} -C ${CHAN_ID} --package-id ${PKG_ID} --name ${CC_ID} --version ${CC_VER} --sequence ${CC_SEQ} --signature-policy ${CC_EP}

    # first call negated as it fails due to specification of validation plugin
    try_fail ${PEER_CMD} lifecycle chaincode checkcommitreadiness -C ${CHAN_ID} --name ${CC_ID} --version ${CC_VER} --sequence ${CC_SEQ} --signature-policy ${CC_EP} -E mock-escc
    try ${PEER_CMD} lifecycle chaincode checkcommitreadiness -C ${CHAN_ID} --name ${CC_ID} --version ${CC_VER} --sequence ${CC_SEQ} --signature-policy ${CC_EP}

    # first call negated as it fails due to specification of validation plugin
    try_fail ${PEER_CMD} lifecycle chaincode commit -o ${ORDERER_ADDR} -C ${CHAN_ID} --name ${CC_ID} --version ${CC_VER} --sequence ${CC_SEQ} --signature-policy ${CC_EP} -E mock-escc
    try ${PEER_CMD} lifecycle chaincode commit -o ${ORDERER_ADDR} -C ${CHAN_ID} --name ${CC_ID} --version ${CC_VER} --sequence ${CC_SEQ} --signature-policy ${CC_EP}

    # first call negated as it fails
    try_fail ${PEER_CMD} lifecycle chaincode initEnclave -o ${ORDERER_ADDR} --peerAddresses "localhost:7051" --name wrong-cc-id
    try ${PEER_CMD} lifecycle chaincode initEnclave -o ${ORDERER_ADDR} --peerAddresses "localhost:7051" --name ${CC_ID}

    # negated test: double registration must fail -- more precisely, inner init succeeds, but ercc registration fails
    try_fail ${PEER_CMD} lifecycle chaincode initEnclave -o ${ORDERER_ADDR} --peerAddresses "localhost:7051" --name ${CC_ID}

    try ${PEER_CMD} lifecycle chaincode querycommitted -C ${CHAN_ID}

    say "do echos"
    for (( i=1; i<=$num_rounds; i++ ))
    do
        # echos
        try_r ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Args": ["echo-'$i'"]}' --waitForEvent
        check_result "echo-$i"
     done
}

# 1. prepare
para
say "Preparing Echo Test ..."
# - clean up relevant docker images
docker_clean ${ERCC_ID}

trap ledger_shutdown EXIT


para
say "Run echo test"

say "- setup ledger"
ledger_init

say "- echo test"
echo_test

say "- shutdown ledger"
ledger_shutdown

para
if [[ "$FAILURES" == 0 ]]; then
    yell "Echo test PASSED"
else
    yell "Echo test had ${FAILURES} failures"
    exit 1
fi
exit 0

