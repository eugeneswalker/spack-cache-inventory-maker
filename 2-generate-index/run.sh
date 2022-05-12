#!/bin/bash -e

DATA_DIR=${DATA_DIR:-../data}
INPUT_FILE=${DATA_DIR}/inventory.json
HTML_OUTPUT_DIR=${DATA_DIR}/html
OUTPUT_FILE=${HTML_OUTPUT_DIR}/index.html
PACKAGE_DIR=${HTML_OUTPUT_DIR}/packages

[[ ! -d ${DATA_DIR} ]] \
 && echo Data directory does not exist: ${DATA_DIR} \
 && exit 1

[[ ! -d ${HTML_OUTPUT_DIR} ]] \
 && mkdir ${HTML_OUTPUT_DIR}

[[ ! -d ${PACKAGE_DIR} ]] \
 && mkdir $PACKAGE_DIR

./generate-index \
  -i ${INPUT_FILE} \
  -o ${OUTPUT_FILE} \
  -p ${PACKAGE_DIR}
