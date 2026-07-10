#!/bin/bash -e
# Robust npm wrapper: retry on transient CI install flakes that retry-settings
# alone don't cover -- cacache EEXIST/ENOENT rename races, esbuild ETXTBSY
# (text file busy) postinstall races, and network blips.
for attempt in 1 2 3; do
    if npm "$@"; then
        exit 0
    fi
    echo "npm $* failed (attempt ${attempt}/3); cleaning node_modules + cache, retrying..." >&2
    rm -rf node_modules
    npm cache clean --force >/dev/null 2>&1 || true
    sleep 5
done
echo "npm $* failed after 3 attempts" >&2
exit 1
