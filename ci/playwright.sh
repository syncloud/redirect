#!/bin/bash -e
DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
PROJECT=$1

if [[ -z "$PROJECT" ]]; then
    echo "usage: $0 <desktop|mobile>"
    exit 1
fi

apt-get update && apt-get install -y default-mysql-client
${DIR}/recreatedb

cd ${DIR}/../www
npm ci
EXIT_CODE=0
npx playwright test --project=${PROJECT} || EXIT_CODE=$?

cd ${DIR}/..
mkdir -p artifact
cp -r www/playwright-report artifact/playwright-report-${PROJECT} 2>/dev/null || true
cp -r www/test-results artifact/playwright-results-${PROJECT} 2>/dev/null || true

exit $EXIT_CODE
