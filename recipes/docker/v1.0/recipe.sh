#!/bin/bash

set -e

if [ -n "$(command -v yum)" ]; then
    yum install -y yum-utils \
        device-mapper-persistent-data \
        lvm2

    yum-config-manager \
        --add-repo \
        https://download.docker.com/linux/centos/docker-ce.repo

    yum install -y docker-ce docker-ce-cli containerd.io

    systemctl enable docker
fi

if [ -n "$(command -v apt)" ]; then
    apt-get install -y \
        apt-transport-https \
        ca-certificates \
        curl \
        gnupg2 \
        software-properties-common

    curl -fsSL https://download.docker.com/linux/debian/gpg | sudo apt-key add -

    add-apt-repository \
        "deb [arch=amd64] https://download.docker.com/linux/debian \
        $(lsb_release -cs) \
        stable"

    apt-get update

    apt-get install -y docker-ce docker-ce-cli containerd.io

    systemctl enable docker
fi

