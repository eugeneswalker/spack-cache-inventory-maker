#!/bin/bash -e

DATA_DIR=${DATA_DIR:-../data/22.11}
SPECFILE_DIR=${SPECFILE_DIR:-${DATA_DIR}/specfiles}
PARALLELISM=12

./parse \
 -o ${DATA_DIR}/inventory-raw.json \
 -n ${PARALLELISM} \
 ${SPECFILE_DIR}
