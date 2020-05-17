#!/bin/bash -e

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cd ${DIR} 

BUILD_DIR=${DIR}/build

mkdir ${BUILD_DIR}
cp -r redirect ${BUILD_DIR}
cd www
jekyl build
cp -r _site ${BUILD_DIR}/www
cd ..

tar czf ${DIR}/redirect.tar.gz -C ${BUILD_DIR} .
