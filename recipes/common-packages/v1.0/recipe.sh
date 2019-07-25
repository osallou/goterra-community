#!/bin/bash

set -e

if [ -n "$(command -v yum)" ]; then
    yum -y install git ansible vim nano python3 httpie
fi

if [ -n "$(command -v apt)" ]; then
    apt-get install -y git ansible vim nano python3 httpie
fi


