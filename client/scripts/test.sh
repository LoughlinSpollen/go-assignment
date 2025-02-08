#! /usr/bin/env sh

test()
{
    echo "assignment client test"

    local CURRENT_DIR=$(pwd)
    local PROJECT_ROOT=$(git rev-parse --show-toplevel)

    rm -rf $PROJECT_ROOT/test_results/client
    mkdir -p $PROJECT_ROOT/test_results/client

    . $PROJECT_ROOT/shared_lib/scripts/config.sh $1
    cd $PROJECT_ROOT/client
    . ./scripts/config.sh $1
    ginkgo -timeout=120s -r -v --output-dir=../test_results/client -procs=1

    cd $CURRENT_DIR
    echo "assignment client test complete"
}

test $1
