#!/bin/bash -e

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cd ${DIR} 

RUBY_VERSION=2.4.1
BUILD_DIR=${DIR}/build

mkdir ${BUILD_DIR}

#apt update
#apt -y install ruby ruby-dev
command curl -sSL https://rvm.io/pkuczynski.asc | gpg2 --import -
curl -sSL https://get.rvm.io | bash -s stable --path ${DIR}/ruby
source ${DIR}/ruby/scripts/rvm
rvm install ${RUBY_VERSION}
gem install jekyll
cd www
jekyll build
cp -r _site ${BUILD_DIR}/www

cd ${DIR}
cp -r redirect ${BUILD_DIR}

tar czf ${DIR}/redirect.tar.gz -C ${BUILD_DIR} .
