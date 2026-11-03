# Project structure

- `backend/` — Go binaries and services (api, www, metrics, cleaner)
- `web/` — Vue 3 frontend using Vite
- `test/` — Python integration tests against the API
- `caddy-smtp/`, `dns-faker/`, `device-faker/`, `ses-faker/`, `paypal-faker/`,
  `stripe-faker/` — stand-ins the CI environment runs instead of the real thing
- `.drone.jsonnet` — pipelines

The frontend directory is `web/`, not `www/`. Paths under `www/` in older notes
or commit messages are stale.

## Build structure

The main pipeline runs, in order: `build web` (`npm install`, `npm run test`,
`npm run lint`, `npm run build`), `test backend`, `test caddy plugin`,
`build backend`, `package`, `docker`, `build test api`, the fakers,
`deploy test`, `test-api`, `test-ui-desktop`, `test-ui-mobile`, then
`deploy uat`, `deploy prod` and `artifact`.

`deploy prod` runs only on a push to `stable`, after the tests above it.

## Frontend test split

- `web/tests/unit/` — Jest unit and component tests (`npm run test`)
- `web/e2e/` — Playwright browser tests (`npm run test:e2e`)

Do not mix them. Jest must not discover `e2e/`, and Playwright must not depend
on Jest config or setup files — keep `web/jest.config.js` ignoring `e2e/`. If
Jest starts executing `web/e2e/*.spec.js`, CI fails before Playwright starts.

Playwright layout: shared `*.spec.js` flows run on both `desktop` and `mobile`,
`*.mobile.spec.js` holds mobile-only assertions, screenshots are taken on
failure in `web/e2e/fixtures.js`, and videos land in `web/test-results`.

Playwright cannot run on this Termux host, so a Playwright change is only
really verified once Drone has run it.

## Database

Schema is golang-migrate, with migrations embedded into the binaries from
`backend/db/migrations` (`//go:embed` in `backend/db/migrations.go`). The api
applies pending migrations on startup; `backend/cmd/migrate` is the manual CLI
(`up`, `version`, `force`).

Add a migration as `0000NN_name.up.sql` + `0000NN_name.down.sql`, numbered from
the highest existing prefix.

## Metrics

`backend/metrics` is a single `prometheus.Collector` passed to
`metrics.NewServer`. A new counter has to be added in three places or it is
silently absent from `/metrics`: the struct, `New()`, and both `Describe` and
`Collect`.

Counters labelled with anything a browser sends must bound the label values —
`event.Events` for event names and `label.Allowed` for languages both reject or
fold unknown values, so a crafted request cannot invent new time series.

## Artifacts

Uploaded by the `artifact` step, served at `http://ci.syncloud.org:8081`:

```sh
curl -s "http://ci.syncloud.org:8081/files/redirect/{N}/"
```

A redirect build holds `redirect-{N}.tar.gz`, `distro/`, `playwright-report/`
and `playwright-results/<test-slug>/` with `error-context.md`,
`failure-full-page.png`, `trace.zip` and `video.webm`. Those come from
`web/playwright-report/` and `web/test-results/`.

To read a Playwright failure without a browser:

```sh
curl -s "http://ci.syncloud.org:8081/files/redirect/{N}/playwright-results/<slug>/error-context.md"
curl -O "http://ci.syncloud.org:8081/files/redirect/{N}/playwright-results/<slug>/trace.zip"
unzip -q trace.zip -d trace/ && grep -oE "/some-path\?[^\" ]*" trace/0-trace.trace | head
```
