#!/bin/bash -e

BUCKET=${BUCKET:-cache.e4s.io}
PARALLELISM=${PARALLELISM:-32}
DATA_DIR=${DATA_DIR:-../data}
SPECFILE_DIR=${DATA_DIR}/specfiles

mkdir -p ${DATA_DIR}
mkdir -p ${SPECFILE_DIR}

./download \
 -b ${BUCKET} \
 -n ${PARALLELISM} \
 -d ${SPECFILE_DIR}
