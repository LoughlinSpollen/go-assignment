#! /usr/bin/env sh

brew services restart rabbitmq
echo "Waiting 5s for RabbitMQ to initialise"
sleep 5

CURRENT_DIR=$(pwd)
PROJECT_ROOT=$(git rev-parse --show-toplevel)
cd $PROJECT_ROOT/service
. ./scripts/install.sh
. ./scripts/config.sh $1
. ./scripts/build.sh

./build/assignment_service &

cd $CURRENT_DIR
