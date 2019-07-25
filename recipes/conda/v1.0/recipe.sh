#!/bin/bash

set -e

if [ -n "$(command -v yum)" ]; then
    yum -y install wget
fi

if [ -n "$(command -v apt)" ]; then
    apt-get install -y wget
fi


wget https://repo.anaconda.com/miniconda/Miniconda3-latest-Linux-x86_64.sh -O /root/miniconda.sh
bash /root/miniconda.sh -b -p /root/miniconda

echo "export PATH=/root/miniconda/bin:\$PATH" > /etc/profile
