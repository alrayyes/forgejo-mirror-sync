#!/usr/bin/env bash
# Assembles site/reports/ from the test job's junit.xml and coverage.out, in
# the layout rules/published-reports.md describes. CI runs it on every pull
# request so a broken conversion fails before the merge; only the deploy is
# main-only.
#
# usage: build-reports-site.sh <dir holding junit.xml and coverage.out>
set -euo pipefail

cd "$(dirname "$0")/.."

in="${1:?usage: build-reports-site.sh <dir with junit.xml and coverage.out>}"
in="$(cd "$in" && pwd)"
out=site/reports
commit="${GITHUB_SHA:-$(git rev-parse HEAD)}"
built="$(date -u +%Y-%m-%dT%H:%MZ)"

rm -rf site
mkdir -p "$out/tests" "$out/coverage"

cp "$in/junit.xml" "$out/tests/unit.xml"
cp "$in/coverage.out" "$out/coverage/coverage.out"

# Pinned exact: this is the one non-Go-module tool the job runs.
go run github.com/boumenot/gocover-cobertura@v1.5.0 \
  <"$in/coverage.out" >"$out/coverage/coverage.xml"
grep -q '<coverage ' "$out/coverage/coverage.xml"

go tool cover -html="$in/coverage.out" -o "$out/coverage/index.html"

cat >"$out/index.html" <<HTML
<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>forgejo-mirror-sync reports</title>
<style>
  :root { color-scheme: light dark; }
  body { font: 1rem/1.5 system-ui, sans-serif; max-width: 40rem; margin: 2rem auto; padding: 0 1rem; }
</style>
</head>
<body>
<main>
<h1>forgejo-mirror-sync reports</h1>
<p>Built from commit <code>${commit:0:7}</code> on <time datetime="${built}">${built}</time>.</p>
<ul>
  <li><a href="tests/unit.xml">Test results</a> (JUnit XML)</li>
  <li><a href="coverage/">Coverage</a> (HTML)</li>
  <li><a href="coverage/coverage.xml">Coverage</a> (Cobertura XML)</li>
  <li><a href="coverage/coverage.out">Coverage</a> (Go cover profile)</li>
</ul>
</main>
</body>
</html>
HTML
