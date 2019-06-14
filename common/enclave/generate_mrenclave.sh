#!/usr/bin/env bash

# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

# abort on error
set -e

build_dir=$1
enclave_dir=$2

hex_out_name=enclave_hash.hex
base64_out_name=mrenclave
tmp_name=tmp_enc_hash
go_name=mrenclave.go

cd $build_dir
sgx_sign gendata -enclave enclave.so -config $enclave_dir/enclave.config.xml -out $tmp_name
dd if=$tmp_name bs=1 skip=188 of=$hex_out_name count=32
hexdump -C $hex_out_name
base64 $hex_out_name > $base64_out_name
rm $hex_out_name
rm $tmp_name
echo "Enclave hash extracted."
cat $base64_out_name

echo "Create go file"
touch $go_name
echo "package enclave" > $go_name
echo "" >> $go_name
echo "// MrEnclave contains hash of enclave code" >> $go_name
echo "const MrEnclave = \"$(cat $base64_out_name)\"" >> $go_name

