#!/bin/sh

set -x 

BUILD_DIR=$(pwd)/build
INFO_DIR=${BUILD_DIR}/info

go build -buildvcs=false -o ${BUILD_DIR}/ ./...
