#!/bin/bash

PASS=$(grep REDIS_PASSWORD_DOCKER= config/secret.env | cut -d'=' -f2 | tr -d ' ')

docker-compose build

sed "s/\$REDIS_PASSWORD/$PASS/" docker-compose.yml | docker-compose -f - up
