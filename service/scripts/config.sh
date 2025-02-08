#! /usr/bin/env sh

PROJECT_ROOT=$(git rev-parse --show-toplevel)
. $PROJECT_ROOT/shared_lib/scripts/config.sh $1

export ASSIGNMENT_SERVICE_NAME=assignment_service
export ASSIGNMENT_SERVICE_BUILD_VERSION=0.0.1

