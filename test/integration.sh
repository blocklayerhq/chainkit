#!/bin/bash

set -eE

TMP_DIR=/tmp/chainkit-integration-tests
PROJECT_NAME=foo
CMD=""

1_test_create() {
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

2_test_build() {
    # Check that you cannot build outside the project dir
    $CMD build || true
    (
        # Test a build that works
        cd $PROJECT_NAME
        $CMD build
    )
}

3_test_start() {
    $CMD start --cwd $PROJECT_NAME > chainkit-start.log 2>&1 &
    # Give some time for the chain to start
    retry 2 10 "curl -s -I -X GET http://localhost:42001 | grep '200 OK'"
}

4_test_cli() {
    (
        cd $PROJECT_NAME
        $CMD cli status
    )
}

5_test_explorer() {
    # Test if the explorer container is running
    retry 2 10 '[ ! -z $(docker ps -qf label=chainkit.cosmos.explorer) ]'
    retry 2 10 'curl -X GET -I http://localhost:42000'
}

6_test_join() {
    # Check that the first node is successfully registered on the network
    retry 3 30 'grep "Node successfully registered" chainkit-start.log'
    network_id=$(grep 'chainkit join' chainkit-start.log | awk '{print $3}')
    [ ! -z "$network_id" ]
    $CMD join $network_id > chainkit-join.log 2>&1 &
    # Check that we discovered the first node
    retry 3 30 'grep "Discovered node " chainkit-join.log'
    # Check that our node id has been spotted in the first node's log
    node_id=$(grep 'Node ID' chainkit-join.log | awk '{print $5}')
    [ ! -z "$node_id" ]
    retry 3 30 'grep "Discovered node $node_id" chainkit-start.log'
    # The 2nd application should be live on port 42011 (port allocation)
    curl -s -I -X GET http://localhost:42011 | grep '200 OK'
}

# Retry a command for 20 sec
retry() {
    unset_trap
    for i in $(seq 1 $2) ; do
        eval "$3" && set_trap && return || true
        sleep $1
    done
    # Reset trap before return
    set_trap
    echo "!!! Retry failed on command: $3"
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
        tail -n 50 chainkit-*.log
    )
    echo "!!! FAILED on line $1"
}

set_trap() {
    trap 'show_logs $LINENO' ERR SIGHUP SIGINT SIGTERM
}

unset_trap() {
    trap - ERR
}

run_tests() {
    CMD="$1"
    # clear-up tmp-dir if exists
    cleanup
    set_trap
    mkdir -p ${TMP_DIR}/src
    (
        cd ${TMP_DIR}/src
        export GOPATH=$TMP_DIR
        test_suite=$(typeset -f | awk '/ \(\) $/ && /^[0-9]+_test_/ {print $1}')
        set -x
        for fn in $test_suite ; do
            echo "############## Running $fn"
            $fn
        done
    )
    echo "All tests passed!"
    cleanup
}

[ -z "$1" ] && {
    echo "Usage: $0 <absolute path to chainkit binary>"
    exit 1
}
run_tests "$1"
