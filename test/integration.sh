#!/bin/sh

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
    sleep 5
    tail chainkit-start.log
    curl -I -X GET --fail http://localhost:42001
}

cleanup() {
    docker rm -f chainkit-$PROJECT_NAME
    docker rmi chainkit-$PROJECT_NAME
}

run_tests() {
    CMD="$1"
    # clear-up tmp-dir if exists
    rm -rf $TMP_DIR
    mkdir -p ${TMP_DIR}/src
    (
        cd ${TMP_DIR}/src
        export GOPATH=$TMP_DIR
        set -e
        set -x
        test_create
        test_build
        test_start
        set +x
        set +e
        cleanup
    )
}

[ -z "$1" ] && {
    echo "Usage: $0 <absolute path to chainkit binary>"
    exit 1
}
run_tests "$1"
