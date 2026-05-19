#!/usr/bin/env python3
"""
Copy historical redirect metrics from graphite whisper files into VictoriaMetrics.

Runs on the UAT host (where /opt/graphite/storage/whisper and VM share a box).
For each known whisper path, reads the full history via whisper-fetch.py,
maps to the new Prometheus name + labels, and posts to VM via
/api/v1/import/prometheus.

Usage:
  sudo python3 backfill.py [--vm http://127.0.0.1:8428] [--dry-run]
"""

import argparse
import json
import subprocess
import sys
import urllib.request

WHISPER_ROOT = "/opt/graphite/storage/whisper"
WHISPER_FETCH = "/usr/local/bin/whisper-fetch.py"
FROM_EPOCH = 0  # full history

ENVS = {
    "redirect-prod2": "prod",
    "redirect-uat": "uat",
}

# (relative whisper path) -> (metric name, extra labels in addition to env+job)
DB_GAUGE_PATHS = {
    "db/devices.wsp":           ("redirect_db_devices", {}),
    "db/domains.wsp":           ("redirect_db_domains", {}),
    "db/users/all.wsp":         ("redirect_db_users", {"state": "all"}),
    "db/users/active.wsp":      ("redirect_db_users", {"state": "active"}),
    "db/users/dead.wsp":        ("redirect_db_users", {"state": "dead"}),
    "db/users/subscribed.wsp":  ("redirect_db_users", {"state": "subscribed"}),
    "db/users/online.wsp":      ("redirect_db_users", {"state": "online"}),
}


def fetch_whisper(path):
    out = subprocess.check_output(
        [WHISPER_FETCH, "--json", f"--from={FROM_EPOCH}", path],
        text=True,
    )
    return json.loads(out)


def format_labels(labels):
    return ",".join(f'{k}="{v}"' for k, v in sorted(labels.items()))


def render_lines(metric, labels, start, step, values):
    label_str = format_labels(labels)
    head = f"{metric}{{{label_str}}}"
    ts_ms = start * 1000
    step_ms = step * 1000
    for v in values:
        if v is not None:
            yield f"{head} {v} {ts_ms}"
        ts_ms += step_ms


def post_to_vm(vm_url, body):
    req = urllib.request.Request(
        f"{vm_url}/api/v1/import/prometheus",
        data=body.encode("utf-8"),
        method="POST",
    )
    with urllib.request.urlopen(req, timeout=60) as resp:
        if resp.status >= 300:
            raise RuntimeError(f"VM import failed: {resp.status}")


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--vm", default="http://127.0.0.1:8428")
    ap.add_argument("--dry-run", action="store_true")
    args = ap.parse_args()

    total_points = 0
    for prefix, env in ENVS.items():
        for rel, (metric, extra) in DB_GAUGE_PATHS.items():
            path = f"{WHISPER_ROOT}/{prefix}/{rel}"
            try:
                data = fetch_whisper(path)
            except subprocess.CalledProcessError:
                print(f"skip (no whisper): {prefix}/{rel}", file=sys.stderr)
                continue
            except FileNotFoundError:
                print(f"skip (no file): {prefix}/{rel}", file=sys.stderr)
                continue

            labels = {
                "env": env,
                "instance": "172.17.0.1:9092",
                "job": "redirect-www",
                **extra,
            }
            lines = list(render_lines(
                metric, labels, data["start"], data["step"], data["values"],
            ))
            if not lines:
                print(f"{prefix}/{rel}: 0 non-null points", file=sys.stderr)
                continue

            total_points += len(lines)
            print(f"{prefix}/{rel}: {len(lines)} points -> {metric}{{{format_labels(labels)}}}", file=sys.stderr)

            if not args.dry_run:
                body = "\n".join(lines) + "\n"
                post_to_vm(args.vm, body)

    print(f"\nTotal points: {total_points}", file=sys.stderr)
    if args.dry_run:
        print("(dry-run, nothing written)", file=sys.stderr)


if __name__ == "__main__":
    main()
