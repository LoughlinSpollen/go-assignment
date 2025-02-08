#!/usr/bin/env sh

CURRENT_DIR=$(pwd)
PROJECT_ROOT=$(git rev-parse --show-toplevel)

cd $PROJECT_ROOT/shared_lib
ginkgo --timeout=120s -r -v --output-dir=../test_results/shared_lib -procs=1 

cd $CURRENT_DIR