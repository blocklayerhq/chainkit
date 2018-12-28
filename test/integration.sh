#!/bin/bash

set -eE

TMP_DIR=/tmp/chainkit-integration-tests
PROJECT_NAME=foo
CMD=""

test_create() {
    $CMD create $PROJECT_NAME
    (
        # Check that key files have been created
        cd $PROJECT_NAME
        [ -f chainkit.yml ]
        [ -f Dockerfile ]
        [ -f app.go ]
        [ -d cmd ]
    )
    # Check that creating the same project fails
    $CMD create $PROJECT_NAME || true
}

test_build() {
    # Check that you cannot build outside the project dir
    $CMD build || true
    (
        # Test a build that works
        cd $PROJECT_NAME
        $CMD build
    )
}

test_start() {
    $CMD start --cwd $PROJECT_NAME > chainkit-start.log 2>&1 &
    # Give some time for the chain to start
    retry "curl -s -I -X GET http://localhost:42001 | grep '200 OK'"
}

test_cli() {
    (
        cd $PROJECT_NAME
        $CMD cli status
    )
}

test_explorer() {
    # Test if the explorer container is running
    retry '[ ! -z $(docker ps -qf label=chainkit.cosmos.explorer) ]'
    retry 'curl -X GET -I http://localhost:42000'
}

# Retry a command for 20 sec
retry() {
    for i in $(seq 1 5) ; do
        eval "$1" && return || true
        sleep 4
    done
    false
}

cleanup() {
    echo "Cleaning up..."
    docker ps -aq -f "label=chainkit.project=$PROJECT_NAME" | xargs docker rm -f || true
    docker rmi "chainkit-$PROJECT_NAME" || true
    rm -rf $TMP_DIR 2>/dev/null || true
}

show_logs() {
    (
        cd ${TMP_DIR}/src
        tail -n 50 chainkit-start.log
    )
}

run_tests() {
    CMD="$1"
    # clear-up tmp-dir if exists
    cleanup
    trap show_logs ERR SIGHUP SIGINT SIGTERM
    mkdir -p ${TMP_DIR}/src
    (
        cd ${TMP_DIR}/src
        export GOPATH=$TMP_DIR
        set -x
        test_create
        test_build
        test_start
        test_cli
        test_explorer
    )
    echo "All tests passed!"
    cleanup
}

[ -z "$1" ] && {
    echo "Usage: $0 <absolute path to chainkit binary>"
    exit 1
}
run_tests "$1"
