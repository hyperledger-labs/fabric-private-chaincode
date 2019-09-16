/*
 * Copyright 2019 Intel Corporation
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#ifndef _COMMON_SGXCCLIB_H_
#define _COMMON_SGXCCLIB_H_

#include "types.h"

#ifdef __cplusplus
extern "C" {
#endif

int sgxcc_create_enclave(enclave_id_t* eid, const char* enclave_file);
int sgxcc_destroy_enclave(enclave_id_t eid);
int sgxcc_get_quote_size(uint8_t* p_sig_rl, uint32_t sig_rl_size, uint32_t* p_quote_size);
int sgxcc_get_target_info(enclave_id_t eid, target_info_t* target_info);
int sgxcc_get_local_attestation_report(
    enclave_id_t eid, target_info_t* target_info, report_t* report, ec256_public_t* pubkey);
int sgxcc_get_remote_attestation_report(enclave_id_t eid,
    quote_t* quote,
    uint32_t quote_size,
    ec256_public_t* pubkey,
    spid_t* spid,
    uint8_t* p_sig_rl,
    uint32_t sig_rl_size);
int sgxcc_get_pk(enclave_id_t eid, ec256_public_t* pubkey);
int sgxcc_get_egid(unsigned int* p_egid);

#define ERROR_CHECK_RET(func_name)                             \
    if (ret != SGX_SUCCESS)                                    \
    {                                                          \
        LOG_ERROR("Lib: ERROR - " #func_name ": ret=%d", ret); \
        return ret;                                            \
    }

#define ERROR_CHECK(func_name)                                                 \
    ERROR_CHECK_RET(func_name)                                                 \
    if (enclave_ret != SGX_SUCCESS)                                            \
    {                                                                          \
        LOG_ERROR("Lib: ERROR - " #func_name ": enclave_ret=%d", enclave_ret); \
        return enclave_ret;                                                    \
    }

#define ERROR_CHECK_AND_RETURN(func_name) \
    ERROR_CHECK(func_name)                \
    return SGX_SUCCESS;

#ifdef __cplusplus
}
#endif /* __cplusplus */

#endif /* !_COMMON_SGXCCLIB_H_ */
