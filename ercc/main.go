/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"log"
	"os"

	"github.com/hyperledger-labs/fabric-private-chaincode/ercc/attestation"
	"github.com/hyperledger-labs/fabric-private-chaincode/ercc/registry"
	"github.com/hyperledger-labs/fabric-private-chaincode/internal/utils"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type serverConfig struct {
	CCID    string
	Address string
}

func main() {

	c := &registry.Contract{}
	c.Verifier = attestation.NewVerifier()
	c.IEvaluator = &utils.IdentityEvaluator{}

	ercc, err := contractapi.NewChaincode(c)
	if err != nil {
		log.Panicf("error create enclave registry chaincode: %s", err)
	}

	// start chaincode as a service
	config := serverConfig{
		CCID:    os.Getenv("CHAINCODE_ID"),
		Address: os.Getenv("CHAINCODE_SERVER_ADDRESS"),
	}

	server := &shim.ChaincodeServer{
		CCID:    config.CCID,
		Address: config.Address,
		CC:      ercc,
		TLSProps: shim.TLSProperties{
			Disabled: true,
		},
	}

	log.Printf("starting enclave registry (%s)\n", config.CCID)

	if err := server.Start(); err != nil {
		log.Panicf("error starting enclave registry chaincode: %s", err)
	}

	// alternatively we can start the chaincode in the normal way and let it connect the its peer
	// TODO integrate the code below
	// some switch is needed (e.g., via --arg or ENV_VAR) to start the chaincode in one or the other mode
	// note that this code will also be shared with ecc

	//ercc, err := contractapi.NewChaincode(&registry.Contract{})
	//if err != nil {
	//	log.Panicf("Error creating registry chaincode: %v", err)
	//}
	//
	//if err := ercc.Start(); err != nil {
	//	log.Panicf("Error starting registry chaincode: %v", err)
	//}

}
