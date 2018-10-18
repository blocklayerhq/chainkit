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
        echo docker run --rm \
            -v "${data_dir}/${APP_NAME}d:/root/.${APP_NAME}d" \
            -v "${data_dir}/${APP_NAME}cli:/root/.${APP_NAME}cli" \
            ${APP_NAME}:latest ${APP_NAME}d init
    )
}

build() {
    echo "Build."
}

run() {
    echo "Run."
    # invoke build
    #TBD
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