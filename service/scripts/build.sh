#! /usr/bin/env sh

build()
{
  local CURRENT_DIR=$(pwd)
  local PROJECT_ROOT=$(git rev-parse --show-toplevel)

  cd $PROJECT_ROOT
  if [ ! -d "build" ]; then
    mkdir -p $PROJECT_ROOT/service/build
  else
    rm -rf $PROJECT_ROOT/service/build/*
  fi

  . $PROJECT_ROOT/shared_lib/scripts/config.sh $1
  cd $PROJECT_ROOT/service
  . scripts/config.sh $1
  go build -ldflags="-X 'main.Version=${ASSIGNMENT_SERVICE_BUILD_VERSION}'" -v -o $PROJECT_ROOT/service/build/assignment_service

  cd $CURRENT_DIR
}

build $1
