#!/bin/sh

# set verbose mode
set -x

echo "Set up docker-compose"

KAFKA_TOPICS=$(echo $KAFKA_TOPICS | sed 's/,/:1:1,/g')
KAFKA_TOPICS="$KAFKA_TOPICS:1:1"
echo "KAFKA TOPICS: $KAFKA_TOPICS"
echo "Default KAFKA_BROKER: $KAFKA_BROKER"

# replace in docker-compose.yml
sed -i "" "s/<KAFKA_BROKER>/$KAFKA_BROKER/g" docker-composer.yml
sed -i "" "s/<KAFKA_TOPICS>/$KAFKA_TOPICS/g" docker-composer.yml

# run docker-compose
docker compose -f "docker-composer.yml" up -d --build