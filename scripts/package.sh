#!/bin/sh

BASEDIR=$(dirname "$0")
SCRIPT_DIR=$(cd $BASEDIR && pwd)
PROJECT_DIR=$(dirname $SCRIPT_DIR)
BUILD_DIR=${PROJECT_DIR}/build
PACKAGE_DIR=${PROJECT_DIR}/package
DIST_DIR=${PROJECT_DIR}/dist

. ${BUILD_DIR}/buildinfo

rm -rf ${PACKAGE_DIR} ${DIST_DIR}
mkdir -p ${PACKAGE_DIR} ${DIST_DIR}

cd ${PACKAGE_DIR}
cp ${BUILD_DIR}/Responder .
cp ${BUILD_DIR}/buildinfo .

zip ${DIST_DIR}/${ZIPFILE} *