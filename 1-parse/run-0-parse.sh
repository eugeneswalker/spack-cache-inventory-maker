#!/bin/bash -e

DATA_DIR=${DATA_DIR:-../data}
SPECFILE_DIR=${DATA_DIR}/specfiles

./parse \
 -o ${DATA_DIR}/inventory-raw.json \
 -n 12 \
 ${SPECFILE_DIR}
