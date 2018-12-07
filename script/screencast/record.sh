#!/bin/bash
set -e

ROOTDIR="$(cd "$(dirname "$0")"; pwd -P)/"

# Dependencies:
# brew install asciinema
# yarn global add svg-term-cli

rm -f "$ROOTDIR/screencast.json" "$ROOTDIR/screencast.svg"

asciinema rec --command "sh $ROOTDIR/script.sh" "$ROOTDIR/screencast.json"
sed -i.bak "s;$HOME;~;g" "$ROOTDIR/screencast.json"
rm -f "$ROOTDIR/screencast.json.bak"

svg-term --window --in "$ROOTDIR/screencast.json" --out "$ROOTDIR/screencast.svg"
