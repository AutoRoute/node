#!/bin/bash
set -e
rm -r fs || true
go build github.com/AutoRoute/node/autoroute
mkdir fs
mkdir fs/usr
mkdir fs/usr/lib
cp /usr/lib/libpthread.so.0 fs/usr/lib
cp /usr/lib/libc.so.6 fs/usr/lib
cp /usr/lib/libnss_files.so.2 fs/usr/lib
cp /usr/lib/libnss_dns.so.2 fs/usr/lib
mkdir fs/lib64
cp /lib64/ld-linux-x86-64.so.2 fs/lib64
mkdir fs/etc
cp /etc/host.conf fs/etc
cp /etc/nsswitch.conf fs/etc
cp /etc/hosts fs/etc
cp /etc/resolv.conf fs/etc
docker build .
