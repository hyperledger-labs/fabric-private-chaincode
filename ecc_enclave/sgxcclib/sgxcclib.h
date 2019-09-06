/*
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#ifndef _SGXCCLIB_H_
#define _SGXCCLIB_H_

#include "types.h"

#ifdef __cplusplus
extern "C" {
#endif

int sgxcc_create_enclave(enclave_id_t* eid, const char* enclave_file);

int sgxcc_destroy_enclave(enclave_id_t eid);

uint32_t sgxcc_get_quote_size(void);

int sgxcc_get_local_attestation_report(
    enclave_id_t eid, target_info_t* target_info, report_t* report, ec256_public_t* pubkey);

int sgxcc_get_remote_attestation_report(
    enclave_id_t eid, quote_t* quote, uint32_t quote_size, ec256_public_t* pubkey, spid_t* spid);

int sgxcc_get_target_info(enclave_id_t eid, target_info_t* target_info);

int sgxcc_bind(enclave_id_t eid, report_t* report, ec256_public_t* pubkey);

int sgxcc_init(enclave_id_t eid,
    const char* args,
    // Note: no PK, init only/mainly useful iff it involves explicit org-approval and then should be
    // public
    uint8_t* response,
    uint32_t response_len_in,
    uint32_t* response_len_out,
    ec256_signature_t* signature,
    void* ctx);

int sgxcc_invoke(enclave_id_t eid,
    const char* args,
    const char* pk,  // client pk used for args encryption, if null
                     // no encryption used
    uint8_t* response,
    uint32_t response_len_in,
    uint32_t* response_len_out,
    ec256_signature_t* signature,
    void* ctx);

int sgxcc_get_pk(enclave_id_t eid, ec256_public_t* pubkey);

#ifdef __cplusplus
}
#endif /* __cplusplus */

#endif /* !_SGXCCLIB_H_ */
