#!/bin/bash -e

DATA_DIR=${DATA_DIR:-../data/22.11}
INPUT_FILE=${DATA_DIR}/inventory-raw.json
OUTPUT_FILE=${DATA_DIR}/inventory.json
S3_SECRET_FILE=../s3.secrets
BUILDCACHE_PREFIX=${S3_PREFIX:-build_cache/}

[[ ! -f ${S3_SECRET_FILE} ]] \
 && echo Error: S3 secret file not found: ${S3_SECRET_FILE} \
 && exit 1

. ${S3_SECRET_FILE}

time \
 ./add-modification-times.py ${INPUT_FILE} ${OUTPUT_FILE}
