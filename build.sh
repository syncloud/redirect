#!/bin/bash -e

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cd ${DIR} 

if [[ -z "$1" ]]; then
    echo "usage $0 version"
    exit 1
fi

VERSION=$1
BUILD_DIR=${DIR}/build

mkdir ${BUILD_DIR}

cd www
mkdir .jekyll-cache
jekyll build
cp -r _site ${BUILD_DIR}/www

cd ${DIR}
cp -r redirect ${BUILD_DIR}
cp requirements.txt ${BUILD_DIR}
cp -r config ${BUILD_DIR}
cp -r apache ${BUILD_DIR}
cp -r emails ${BUILD_DIR}
cp redirect_*.wsgi  ${BUILD_DIR}

mkdir ${DIR}/artifact
tar czf ${DIR}/artifact/redirect-${VERSION}.tar.gz -C ${BUILD_DIR} .
