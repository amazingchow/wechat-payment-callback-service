#!/bin/bash

set -e

function run
{
    echo -n -e \
    "wx8888888888888888\n1414561699\n5K8264ILTKCH16CQ2502SI8ZNMTM67VS\nprepay_id=wx201410272009395522657a690389285100\n" \
    | openssl dgst -sha256 -sign ./fixtures/private.key\
    | openssl base64 -A
}

run $@
