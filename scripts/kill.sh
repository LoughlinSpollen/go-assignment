#! /usr/bin/env sh

ps aux | grep '/build/assignment_service' | grep -v grep | awk '{print $2}' | xargs kill -9
brew services stop rabbitmq
