#!/bin/bash

set -e

loopback2 &
sleep 1
ip link set dev looptap0-0 up
ip link set dev looptap0-1 up
sleep 1
ip0=$(ip -f inet6 -o addr show looptap0-0 |cut -d\  -f 7 | cut -d/ -f 1)
ip1=$(ip -f inet6 -o addr show looptap0-1 |cut -d\  -f 7 | cut -d/ -f 1)
echo $ip0
echo $ip1
autoroute -listen "[$ip0%looptap0-0]:31337" -auto=true -interface='looptap0-0' &
sleep 1
autoroute -listen "[$ip1%looptap0-1]:31337" -auto=true -interface='looptap0-1'

