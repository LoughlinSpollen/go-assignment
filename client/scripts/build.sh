#! /usr/bin/env sh

build()
{
  local CURRENT_DIR=$(pwd)
  local PROJECT_ROOT=$(git rev-parse --show-toplevel)

  cd $PROJECT_ROOT
  if [ ! -d "build" ]; then
    mkdir -p $PROJECT_ROOT/client/build
  else
    rm -rf $PROJECT_ROOT/client/build/*
  fi

  . $PROJECT_ROOT/shared_lib/scripts/config.sh $1
  cd $PROJECT_ROOT/client
  . scripts/config.sh $1
  go build -ldflags="-X 'main.Version=${ASSIGNMENT_CLIENT_BUILD_VERSION}'" -v -o $PROJECT_ROOT/client/build/assignment_client

  cd $CURRENT_DIR
}

build $1
