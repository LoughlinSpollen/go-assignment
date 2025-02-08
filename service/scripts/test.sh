#! /usr/bin/env sh

test()
{
    echo "assignment service test"

    local CURRENT_DIR=$(pwd)
    local PROJECT_ROOT=$(git rev-parse --show-toplevel)

    rm -rf $PROJECT_ROOT/test_results/service
    mkdir -p $PROJECT_ROOT/test_results/service

    . $PROJECT_ROOT/shared_lib/scripts/config.sh $1
    cd $PROJECT_ROOT/service
    . ./scripts/config.sh $1
    ginkgo -timeout=120s -r -v --output-dir=../test_results/service -procs=1

    cd $CURRENT_DIR
    echo "assignment service test complete"
}

test $1
