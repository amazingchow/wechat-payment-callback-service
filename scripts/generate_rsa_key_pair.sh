#!/usr/bin/env bash

# 遇到执行出错，直接终止脚本的执行
set -o errexit

function logger_print
{
    local prefix="[$(date +%Y/%m/%d\ %H:%M:%S)]"
    echo "${prefix}$@" >&2
}

function run
{
    openssl genrsa -out ./fixtures/keypair.pem 4096
    logger_print "[INFO]" "generated RSA key pair."
    openssl rsa -in ./fixtures/keypair.pem -pubout -out ./fixtures/public.crt
    logger_print "[INFO]" "generated public key."
    openssl pkcs8 -topk8 -inform PEM -outform PEM -nocrypt -in ./fixtures/keypair.pem -out ./fixtures/private.key
    logger_print "[INFO]" "generated private key."
}

run $@
