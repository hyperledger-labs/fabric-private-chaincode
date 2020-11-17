/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

const MrEnclaveStateKey = "MRENCLAVE"

// Response contains the response data and signature produced by the enclave
// TODO remove once ecc uses new ChaincodeResponseMessage
type Response struct {
	ResponseData []byte `json:"ResponseData"`
	Signature    []byte `json:"Signature"`
	PublicKey    []byte `json:"PublicKey"`
}

func UnmarshalResponse(respBytes []byte) (*Response, error) {
	response := &Response{}
	err := json.Unmarshal(respBytes, response)
	if err != nil {
		return nil, fmt.Errorf("unmarshaling FPC response err: %s", err)
	}
	return response, nil
}

// TODO replace this with a proto? TBD
type AttestationParams struct {
	Params []string `json:"params"`
}

const sep = "."

func Read(file string) []byte {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}
	if data == nil {
		panic(fmt.Errorf("File is empty"))
	}
	return data
}

func IsFPCCompositeKey(comp string) bool {
	return strings.HasPrefix(comp, sep) && strings.HasSuffix(comp, sep)
}

func TransformToFPCKey(comp string) string {
	return strings.Replace(comp, "\x00", sep, -1)
}

func SplitFPCCompositeKey(comp_str string) []string {
	// check it has sep in front and end
	if !IsFPCCompositeKey(comp_str) {
		panic("comp_key has wrong format")
	}
	comp := strings.Split(comp_str, sep)
	return comp[1 : len(comp)-1]
}

func ValidateEndpoint(endpoint string) error {
	colon := strings.LastIndexByte(endpoint, ':')
	if colon == -1 {
		return fmt.Errorf("invalid format")
	}

	_, err := strconv.Atoi(endpoint[colon+1:])
	if err != nil {
		return errors.Wrap(err, "invalid port")
	}

	return nil
}
