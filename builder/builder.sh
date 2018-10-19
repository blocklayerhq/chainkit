#!/bin/sh

set -e

APP_NAME=myapp
WORKDIR="$(dirname $0)"

init() {
    (
        cd $WORKDIR
        #NOTE: template generation happens here
        echo "### FAKE: generated template in directory $APP_NAME"
        (
            cd $APP_NAME
            echo "### Building code..."
            docker build -t "$APP_NAME" .
        )
        echo "### Generating config and genesis in ./data"
        data_dir="$(pwd)/data"
        docker run --rm \
            -v "${data_dir}/${APP_NAME}d:/root/.${APP_NAME}d" \
            -v "${data_dir}/${APP_NAME}cli:/root/.${APP_NAME}cli" \
            ${APP_NAME}:latest ${APP_NAME}d init
    )
}

build() {
    echo "Build."
}

run_voyager() {
    data_dir="$(pwd)/data"
    export COSMOS_HOME="$data_dir"
    export COSMOS_NODE=localhost
    open /Applications/Cosmos\ Voyager.app
}

run() {
    (
        cd $WORKDIR/$APP_NAME
        data_dir="$(pwd)/data"
        echo "Run."
        run_voyager
        echo docker run --rm -it \
            -v "${data_dir}/${APP_NAME}d:/root/.${APP_NAME}d" \
            -v "${data_dir}/${APP_NAME}cli:/root/.${APP_NAME}cli" \
            -p 26656-26657:26656-26657 \
            ${APP_NAME}:latest ${APP_NAME}d start
        sleep 1
        pkill "Cosmos Voyager"
    )
}

case "$1" in
    init | build | run | console)
        $1
        ;;
    *)
        echo "Unkown command."
        echo "Commands: init, build, run, console"
        ;;
esac
