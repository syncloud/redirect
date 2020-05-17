#!/bin/bash -e

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cd ${DIR} 

BUILD_DIR=${DIR}/build

apt -y install ruby ruby-dev
gem install jekyll

mkdir ${BUILD_DIR}
cp -r redirect ${BUILD_DIR}
cd www
jekyll build
cp -r _site ${BUILD_DIR}/www
cd ..

tar czf ${DIR}/redirect.tar.gz -C ${BUILD_DIR} .
