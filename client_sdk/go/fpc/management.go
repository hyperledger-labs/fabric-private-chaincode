/*
Copyright IBM Corp. All Rights Reserved.
Copyright 2020 Intel Corporation

SPDX-License-Identifier: Apache-2.0
*/

package fpc

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/hyperledger-labs/fabric-private-chaincode/client_sdk/go/fpc/attestation"
	"github.com/hyperledger-labs/fabric-private-chaincode/internal/protos"
	pbatt "github.com/hyperledger-labs/fabric-private-chaincode/internal/protos/attestation"
	"github.com/hyperledger-labs/fabric-private-chaincode/internal/utils"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
	"github.com/hyperledger/fabric/protoutil"
)

// ManagementAPI provides FPC specific chaincode management functionality.
// ManagementAPI objects should be created using the GetManagementAPI() factory method.
// For an example of its use, see https://github.com/hyperledger-labs/fabric-private-chaincode/blob/master/client_sdk/go/test/main.go
type ManagementAPI interface {
	// InitEnclave initializes and registers an enclave for a particular chaincode.
	//  Parameters:
	//  peerEndpoint is the endpoint on which the enclave should be instantiated.
	//  attestationParams are parameters used during attestation of the instantiated enclave.
	InitEnclave(peerEndpoint string, attestationParams ...string) error
}

// GetManagementAPI is the factory method for ManagementAPI objects.
//  Parameters:
//  network is an initialized Fabric network object
//  chaincodeID is the ID of the target chaincode
//
//  Returns:
//  The ManagementAPI object
func GetManagementAPI(network *gateway.Network, chaincodeID string) ManagementAPI {
	contract := network.GetContract(chaincodeID)
	ercc := network.GetContract("ercc")
	return &managementState{contract: contract, ercc: ercc}
}

type managementState struct {
	contract *gateway.Contract
	ercc     *gateway.Contract
}

func (c *managementState) InitEnclave(peerEndpoint string, attestationParams ...string) error {
	txn, err := c.contract.CreateTransaction(
		"__initEnclave",
		gateway.WithEndorsingPeers(peerEndpoint),
	)
	if err != nil {
		return err
	}

	if err := utils.ValidateEndpoint(peerEndpoint); err != nil {
		return err
	}

	// TODO get real attestation params from somewhere (initially probably best via `SGX_CREDENTIALS_PATH` env variable like `fabric/bin/peer.sh` ...)
	type Params struct {
		AttestationType string `json:"attestation_type"`
	}

	serializedJSONParams, err := json.Marshal(Params{
		AttestationType: "simulated",
	})
	if err != nil {
		shim.Error(err.Error())
	}

	initMsg := &protos.InitEnclaveMessage{
		PeerEndpoint: peerEndpoint,
		AttestationParams: protoutil.MarshalOrPanic(&pbatt.AttestationParameters{
			// TODO this base64 encoding is nasty but needed because in `cc_data.cpp` we do `attestation_parameters_s = base64_decode(b64_ap_s);`
			// this is probably because in peer cli a json string would have caused trouble.  That part could be addressed by adding another command to `utils/fabric/peer-cli-assist`.
			Parameters: []byte(base64.StdEncoding.EncodeToString(serializedJSONParams)),
		}),
	}

	logger.Debugf("calling __initEnclave")
	credentialsBytes, err := txn.Evaluate(utils.MarshallProto(initMsg))
	if err != nil {
		return fmt.Errorf("evaluation error: %s", err)
	}

	var convertedCredentials string
	convertedCredentials, err = ConvertCredentials(string(credentialsBytes))
	if err != nil {
		return fmt.Errorf("evaluation error: %s", err)
	}

	logger.Debugf("calling registerEnclave")
	_, err = c.ercc.SubmitTransaction("registerEnclave", convertedCredentials)
	if err != nil {
		return err
	}

	return nil
}

// perform attestation evidence transformation
func ConvertCredentials(credentialsOnlyAttestation string) (credentialsWithEvidence string, err error) {
	logger.Debugf("Received Credential: '%s'", credentialsOnlyAttestation)
	credentials, err := utils.UnmarshalCredentials(credentialsOnlyAttestation)
	if err != nil {
		return "", fmt.Errorf("cannot decode credentials: %s", err)
	}

	credentials, err = attestation.ToEvidence(credentials)
	if err != nil {
		return "", err
	}
	credentialsOnlyAttestation = utils.MarshallProto(credentials)
	logger.Debugf("Converted to Credential: '%s'", credentialsOnlyAttestation)
	return credentialsOnlyAttestation, nil
}
