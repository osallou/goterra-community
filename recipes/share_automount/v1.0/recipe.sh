#!/bin/bash

set -e

if [ "a${shared_path}" == "a" ]; then
  echo "no shared_path defined, exiting"
  exit 0
fi

mkdir -p /mnt/share

if [ -n "$(command -v yum)" ]; then
    yum -y install nfs-utils
fi

if [ -n "$(command -v apt)" ]; then
    apt-get install -y nfs-common
fi


mount -t nfs ${shared_path} /mnt/share

