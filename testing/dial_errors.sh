#!/bin/bash

set -e

loopback2 -config=sample.json &
sleep 1
ip link set dev i0-0 up
ip link set dev i1-0 up
ip link set dev i2-0 up
ip link set dev i3-0 up
sleep 1
ip0=$(ip -f inet6 -o addr show i0-0 |cut -d\  -f 7 | cut -d/ -f 1)
ip1=$(ip -f inet6 -o addr show i1-0 |cut -d\  -f 7 | cut -d/ -f 1)
ip2=$(ip -f inet6 -o addr show i2-0 |cut -d\  -f 7 | cut -d/ -f 1)
ip3=$(ip -f inet6 -o addr show i3-0 |cut -d\  -f 7 | cut -d/ -f 1)

echo $ip0
echo $ip1
echo $ip2
echo $ip3

sleep 1
autoroute -fake_money -auto=true -nolisten=true -devs='i0-0' &
sleep 1
autoroute -fake_money -auto=true -listen "[$ip1%i1-0]:31337" -devs='i1-0' &
# sleep 1
# autoroute -auto=true -nolisten=true -devs='i3-0' &
sleep 5
trap 'kill $(jobs -pr)' SIGINT SIGTERM EXIT
