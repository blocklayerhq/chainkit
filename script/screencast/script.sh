#!/bin/bash
set -e

CHAINKIT=/Users/al/go/src/github.com/blocklayerhq/chainkit/chainkit

rm -rf ~/go/src/github.com/demoapp
cd ~/go/src/github.com

printf '\e[32m%s\e[m' "$ "
echo "chainkit create demoapp" | pv -qL $[10+(-2 + RANDOM%5)]
$CHAINKIT create demoapp

printf '\e[32m%s\e[m' "$ "
sleep 1
echo "cd demoapp" | pv -qL $[10+(-2 + RANDOM%5)]
cd demoapp

printf '\e[32m%s\e[m' "$ "
sleep 1
echo "chainkit start" | pv -qL $[10+(-2 + RANDOM%5)]
($CHAINKIT start --no-color=false | tee output &)
trap "kill 0" EXIT

while [ -z "$(grep "Node successfully registered" ~/go/src/github.com/demoapp/output)" ]; do
	sleep 1
done

sleep 3
