#!/bin/bash

cd "$( dirname "${BASH_SOURCE[0]}" )"
cd ..

if [ "$#" -ne 1 ]; then
    echo "Usage: $0 env"
    exit 1
fi

ENV=$1
GIT_URL=https://github.com/syncloud/redirect
REV_FILE=revision
LATEST_REV=$(git ls-remote ${GIT_URL} refs/heads/master | cut -f1)
if [ "$LATEST_REV" == "" ]; then
  echo "Unable to get latest version"
  exit 1
fi

if [ -f ${REV_FILE} ]; then
  CURRENT_REV=$(<${REV_FILE})
  if [ "$CURRENT_REV" == "$LATEST_REV" ]; then
    exit 1
  fi
fi
echo "$LATEST_REV" > ${REV_FILE}

git pull

sudo pip install -r requirements.txt
sudo service apache2-${ENV} restart

cp ${REV_FILE} www/${REV_FILE}
cd www
jekyll

cd ..
cp deploy.log deploy.log.last