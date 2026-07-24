#!/bin/bash -e
DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
PROJECT=$1

if [[ -z "$PROJECT" ]]; then
    echo "usage: $0 <desktop|mobile>"
    exit 1
fi

apt-get update && apt-get install -y default-mysql-client sshpass openssh-client
${DIR}/recreatedb

IP=$(getent hosts www.syncloud.test | awk '{print $1}')
echo "$IP syncloud.test api.syncloud.test auth.syncloud.test" >> /etc/hosts

cd ${DIR}/../www
bash ${DIR}/npm.sh ci
EXIT_CODE=0
npx playwright test --project=${PROJECT} || EXIT_CODE=$?

cd ${DIR}/..
OUT=artifact/${PROJECT}
mkdir -p ${OUT}
for dir in www/test-results/*/; do
    name=$(basename ${dir})
    [[ -f ${dir}video.webm ]] && cp ${dir}video.webm ${OUT}/${name}.webm
    [[ -f ${dir}failure-full-page.png ]] && cp ${dir}failure-full-page.png ${OUT}/${name}.png
done
cp -r www/test-results/logs ${OUT}/logs

exit $EXIT_CODE
