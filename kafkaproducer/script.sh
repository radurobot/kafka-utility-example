#!/bin/sh

# export env vars
# export KAFKA_BROKERS="localhost"
export SECRET_KEY="SECRET"

# run the app
go build -o kafka-producer-example
./kafka-producer-example