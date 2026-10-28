#!/usr/bin/env sh
set -eu
URL='https://raw.githubusercontent.com/aniketnandi/automated-resume-screening/main/data/UpdatedResumeDataSet.csv'
OUT="$(dirname "$0")/../data/UpdatedResumeDataSet.csv"
mkdir -p "$(dirname "$OUT")"
curl -L --fail --retry 2 "$URL" -o "$OUT"
printf 'Downloaded public resume dataset to %s\n' "$OUT"
