#!/bin/bash
set -e

# Dependencies:
# brew install asciinema
# yarn global add svg-term-cli

rm -f "$(dirname $0)/screencast.json" "$(dirname $0)/screencast.svg"

asciinema rec --command 'sh "$(dirname $0)/script.sh"' "$(dirname $0)/screencast.json"
sed -i.bak "s;$HOME;~;g" screencast.json
rm -f "$(dirname $0)/screencast.json.bak"

svg-term --window --in "$(dirname $0)/screencast.json" --out "$(dirname $0)/screencast.svg"
