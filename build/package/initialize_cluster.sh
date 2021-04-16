#!/bin/bash

HOSTS=$(echo $CLUSTER_HOSTS | tr " " "\n")
HOST_PORTS=()
for element in $HOSTS; do
  HOST_NAME=$(echo $element | cut -d ":" -f 1)
  PORT=$(echo $element | cut -d ":" -f 2)
  IP=$(host ${HOST_NAME} | grep "has address" | cut -d " " -f 4)

  HOST_PORTS=("${HOST_PORTS[@]}" "$(echo $IP:$PORT)")
done

redis-cli --cluster create ${HOST_PORTS[@]} --cluster-yes
