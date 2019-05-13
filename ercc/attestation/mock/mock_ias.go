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

package mock

import (
	"github.com/hyperledger-labs/fabric-private-chaincode/ercc/attestation"
)

type MockIAS struct {
}

func (ias *MockIAS) RequestAttestationReport(api_key string, quoteAsBytes []byte) (attestation.IASAttestationReport, error) {
	report := attestation.IASAttestationReport{
		IASReportSignature:          "some X-IASReport-Signature",
		IASReportSigningCertificate: "some X-IASReport-Signing-Certificate",
		IASReportBody:               []byte("Some report body"),
	}

	return report, nil
}

func (ias *MockIAS) GetIntelVerificationKey() (interface{}, error) {
	return attestation.PublicKeyFromPem([]byte(attestation.IntelPubPEM))
}
