#! /usr/bin/env sh

PROJECT_ROOT=$(git rev-parse --show-toplevel)
. $PROJECT_ROOT/shared_lib/scripts/config.sh $1

export ASSIGNMENT_CLIENT_NAME=assignment_client
export ASSIGNMENT_CLIENT_BUILD_VERSION=0.0.0

