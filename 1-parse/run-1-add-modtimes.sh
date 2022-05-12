#!/bin/bash -e

DATA_DIR=${DATA_DIR:-../data}
INPUT_FILE=${DATA_DIR}/inventory-raw.json
OUTPUT_FILE=${DATA_DIR}/inventory.json
S3_SECRET_FILE=../s3.secrets

[[ ! -f ${S3_SECRET_FILE} ]] \
 && echo Error: S3 secret file not found: ${S3_SECRET_FILE} \
 && exit 1

. ${S3_SECRET_FILE}

time \
 ./add-modification-times.py ${INPUT_FILE} ${OUTPUT_FILE}
