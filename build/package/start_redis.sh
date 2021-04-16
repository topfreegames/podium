#!/bin/bash

sed "s/<<PORT>>/$PORT/g" /redis_model.conf > /redis.conf

redis-server /redis.conf
