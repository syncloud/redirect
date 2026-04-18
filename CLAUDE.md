# Debugging CI failures

When a CI build fails, start by identifying the failing step:
```sh
curl -s "http://ci.syncloud.org:8080/api/repos/syncloud/redirect/builds/{N}" | python3 -c "
import json,sys
b=json.load(sys.stdin)
for stage in b.get('stages', []):
    for step in stage.get('steps', []):
        if step.get('status') == 'failure':
            print(f\"stage {stage.get('number')} step {step.get('number')}: {step.get('name')}\")
"
```

Get the log for a specific stage/step:
```sh
curl -s "http://ci.syncloud.org:8080/api/repos/syncloud/redirect/builds/{N}/logs/{stage}/{step}" | python3 -c "
import json,sys
for line in json.load(sys.stdin):
    print(line.get('out', ''), end='')
" | tail -120
```

Example from build `1216`:
- failing pipeline: `redirect-amd64`
- failing step: `stage 1 step 6`
- failure reason: Jest picked up `www/e2e/*.spec.js` and executed Playwright files during `npm run test`

# CI

Web UI:
```text
http://ci.syncloud.org:8080/syncloud/redirect
```

Drone API examples:
```sh
curl -s "http://ci.syncloud.org:8080/api/repos/syncloud/redirect/builds?limit=5"
curl -s "http://ci.syncloud.org:8080/api/repos/syncloud/redirect/builds/{N}"
curl -s "http://ci.syncloud.org:8080/api/repos/syncloud/redirect/builds/{N}/logs/{stage}/{step}"
```

## Build structure

The main repo pipeline currently does:
1. `build web`
   Runs `npm install`, `npm run test`, `npm run lint`, `npm run build`
2. `build backend`
3. `package`
4. `test-integration`
5. `test-ui-playwright`
6. `artifact`

The UI test migration separates test runners:
- `npm run test` = Jest unit tests only
- `npm run test:e2e` = Playwright end-to-end tests only

If Jest starts executing `www/e2e/*.spec.js`, CI will fail before Playwright even starts.
Keep `www/jest.config.js` ignoring `e2e/`.

## Playwright notes

Playwright tests live in:
```text
www/e2e/
```

Configuration:
```text
www/playwright.config.js
```

Current intent:
- shared `*.spec.js` flows run on both `desktop` and `mobile`
- `*.mobile.spec.js` holds mobile-only assertions
- screenshots are taken explicitly on failure in `www/e2e/fixtures.js`
- videos are retained in `www/test-results`

This mirrors the old Selenium confidence model:
- full desktop run
- full mobile run
- extra mobile-specific checks where needed

## CI artifacts

Artifacts are uploaded by the `artifact` step and served by an nginx file browser at `http://ci.syncloud.org:8081`.

List build artifacts:
```sh
curl -s "http://ci.syncloud.org:8081/files/redirect/{N}/"
```

Typical layout for a redirect build:
```
{N}/
  redirect-{N}.tar.gz
  distro/
  playwright-report/
  playwright-results/
    <test-slug>/
      error-context.md
      failure-full-page.png
      trace.zip
      video.webm
```

Fetch a Playwright failure's page snapshot + error:
```sh
curl -s "http://ci.syncloud.org:8081/files/redirect/{N}/playwright-results/<slug>/error-context.md"
```

Download the trace and inspect URLs or network calls locally:
```sh
curl -O "http://ci.syncloud.org:8081/files/redirect/{N}/playwright-results/<slug>/trace.zip"
unzip -q trace.zip -d trace/
grep -oE "/some-path\?[^\" ]*" trace/0-trace.trace | head
```

Those directories are populated from:
- `www/playwright-report/`
- `www/test-results/`

# Project structure

- `backend/` — Go backend binaries and services
- `www/` — Vue 3 frontend using Vite
- `integration/` — Python integration tests and legacy Selenium UI tests
- `.drone.jsonnet` — Drone pipelines

## Frontend test split

- `www/tests/unit/` — Jest unit/component tests
- `www/e2e/` — Playwright browser tests

Do not mix them:
- Jest should not discover `e2e/`
- Playwright should not depend on Jest config or setup files

# Local limitations

Playwright does not run on this Termux/Android host. Syntax checks and repo edits can be done locally, but real Playwright execution must be validated in Drone's Linux environment.
