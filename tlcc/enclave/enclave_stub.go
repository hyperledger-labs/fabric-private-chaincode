/*
* Copyright IBM Corp. 2018 All Rights Reserved.
*
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*
*     http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
 */

package enclave

import (
	"unsafe"

	"github.com/hyperledger-labs/fabric-secure-chaincode/ecc/crypto"
	"github.com/hyperledger/fabric/common/flogging"
)

// #cgo CFLAGS: -I${SRCDIR}/include
// #cgo LDFLAGS: -L${SRCDIR}/lib -ltl
// #include <trusted_ledger.h>
import "C"

const EPID_SIZE = 8
const SPID_SIZE = 16
const MAX_OUTPUT_SIZE = 1024
const MAX_RESPONSET_SIZE = 1024
const SIGNATURE_SIZE = 64
const PUB_KEY_SIZE = 64
const REPORT_SIZE = 432
const TARGET_INFO_SIZE = 512
const CMAC_SIZE = 16

var logger = flogging.MustGetLogger("tl-enclave")

//export golog
func golog(str *C.char) {
	logger.Infof("%s", C.GoString(str))
}

// Stub interface
type Stub interface {
	GetTargetInfo() ([]byte, error)
	// Return report and enclave PK in DER-encoded PKIX format
	GetLocalAttestationReport(targetInfo []byte) ([]byte, []byte, error)
	// Creates an enclave from a given enclave lib file
	Create(enclaveLibFile string) error
	// Init enclave with a given genesis block
	InitWithGenesis(blockBytes []byte) error
	// give enclave next block to validate and append to the ledger
	NextBlock(blockBytes []byte) error
	// verifies state and returns cmac
	GetStateMetadata(key string, nonce []byte, isRangeQuery bool) ([]byte, error)
	// Destroys enclave
	Destroy() error
}

// StubImpl implements the interface
type StubImpl struct {
	eid C.enclave_id_t
}

// NewEnclave starts a new enclave
func NewEnclave() Stub {
	return &StubImpl{}
}

func (e *StubImpl) GetTargetInfo() ([]byte, error) {
	// TODO what is the correct target info size
	targetInfo := make([]byte, TARGET_INFO_SIZE)
	targetInfoPtr := C.CBytes(targetInfo)
	defer C.free(targetInfoPtr)

	C.tlcc_get_target_info(e.eid,
		(*C.target_info_t)(targetInfoPtr))

	return targetInfo, nil
}

func (e *StubImpl) GetLocalAttestationReport(targetInfo []byte) ([]byte, []byte, error) {

	// report
	report := make([]byte, REPORT_SIZE)
	reportPtr := C.CBytes(report)
	defer C.free(reportPtr)

	// pubkey
	pubkey := make([]byte, PUB_KEY_SIZE)
	pubkeyPtr := C.CBytes(pubkey)
	defer C.free(pubkeyPtr)

	// targetInfo
	targetInfoPtr := C.CBytes(targetInfo)
	defer C.free(targetInfoPtr)

	// call enclave
	// TODO read error
	C.tlcc_get_local_attestation_report(e.eid,
		(*C.target_info_t)(targetInfoPtr),
		(*C.report_t)(reportPtr),
		(*C.ec256_public_t)(pubkeyPtr))

	// convert sgx format to DER-encoded PKIX format
	pk, err := crypto.MarshalEnclavePk(C.GoBytes(pubkeyPtr, C.int(PUB_KEY_SIZE)))
	if err != nil {
		return nil, nil, err
	}

	return C.GoBytes(reportPtr, C.int(REPORT_SIZE)), pk, nil
}

func (e *StubImpl) InitWithGenesis(blockBytes []byte) error {
	blockBytesPtr := C.CBytes(blockBytes)
	blockBytesLen := len(blockBytes)
	defer C.free(blockBytesPtr)

	C.tlcc_init_with_genesis(e.eid,
		(*C.uint8_t)(blockBytesPtr), C.uint32_t(blockBytesLen))

	return nil
}

func (e *StubImpl) NextBlock(block []byte) error {
	blockLen := len(block)
	blockPtr := C.CBytes(block)
	defer C.free(blockPtr)

	_, err := C.tlcc_send_block(e.eid,
		(*C.uint8_t)(blockPtr), C.uint32_t(blockLen))

	if err != nil {
		return err
	}

	return nil
}

func (e *StubImpl) GetStateMetadata(key string, nonce []byte, isRangeQuery bool) ([]byte, error) {
	// key
	keyc := C.CString(key)
	defer C.free(unsafe.Pointer(keyc))

	// nonce
	noncePtr := C.CBytes(nonce)
	defer C.free(noncePtr)

	// cmac
	cmac := make([]byte, CMAC_SIZE)
	cmacPtr := C.CBytes(cmac)
	defer C.free(cmacPtr)

	if isRangeQuery {
		C.tlcc_get_multi_state_metadata(e.eid, keyc,
			(*C.uint8_t)(noncePtr),
			(*C.cmac_t)(cmacPtr))
	} else {
		C.tlcc_get_state_metadata(e.eid, keyc,
			(*C.uint8_t)(noncePtr),
			(*C.cmac_t)(cmacPtr))
	}
	return C.GoBytes(cmacPtr, C.int(CMAC_SIZE)), nil
}

// Create starts a new enclave instance
func (e *StubImpl) Create(enclaveLibFile string) error {
	var eid C.enclave_id_t

	f := C.CString(enclaveLibFile)
	defer C.free(unsafe.Pointer(f))

	// todo read error
	C.tlcc_create_enclave(&eid, f)
	e.eid = eid
	return nil
}

// Destroy kills the current enclave instance
func (e *StubImpl) Destroy() error {
	// todo read error
	C.tlcc_destroy_enclave(e.eid)
	return nil
}
