#! /usr/bin/env sh

export ENVIRONMENT=$1

if [ "$1" = "local-dev" ]; then
    export ASSIGNMENT_SERVICE_QUEUE_CONN=amqp://guest:guest@127.0.0.1:5672/
    export ASSIGNMENT_SERVICE_QUEUE_NAME=assignmentQueue
elif [ "$1" = "staging" ]; then
    export ASSIGNMENT_SERVICE_QUEUE_CONN="" # TODO    
elif [ "$1" = "prod" ]; then
    export ASSIGNMENT_SERVICE_QUEUE_CONN="" # TODO
else 
    echo "Error: Invalid environment"
    exit 1
fi
