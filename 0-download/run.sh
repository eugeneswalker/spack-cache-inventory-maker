#!/bin/bash -e

BUCKET=${BUCKET:-cache.e4s.io/}
PARALLELISM=${PARALLELISM:-32}
DATA_DIR=${DATA_DIR:-../data/22.11}
SPECFILE_DIR=${DATA_DIR}/specfiles
BUILDCACHE_PREFIX=${BUILDCACHE_PREFIX:-build_cache/}

mkdir -p ${DATA_DIR}
mkdir -p ${SPECFILE_DIR}

./download \
 -b ${BUCKET} \
 -p ${BUILDCACHE_PREFIX} \
 -n ${PARALLELISM} \
 -d ${SPECFILE_DIR}
