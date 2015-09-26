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
touch fs/etc/hosts
echo 'nameserver 8.8.4.4' > fs/etc/resolv.conv
echo 'nameserver 8.8.8.8' >> fs/etc/resolv.conv
echo 'hosts: files dns' > fs/etc/nsswitch.conf
cat << EOF > fs/README
GLIBC is covered by the GPL and LGPL licenses.
see http://www.gnu.org/software/libc for more information.
EOF

docker build .