#!/bin/bash

set -e

if [ -n "$(command -v yum)" ]; then
    yum install -y epel-release
    yum groupinstall -y "MATE Desktop"
    yum -y install x2goserver-xsession
fi

if [ -n "$(command -v apt)" ]; then
    # Mate
    export DEBIAN_FRONTEND=noninteractive
    apt-get update
    apt-get install -y xorg mesa-utils
    apt-get install -y mate-desktop-environment
    apt-get --no-install-recommends install -y lightdm
    apt-get install -y sysv-rc-conf
    rm /root/.profile
    cat <<EOT>> /root/.profile
    if [ "$BASH" ]; then
    if [ -f ~/.bashrc ]; then
        . ~/.bashrc
    fi
    fi
    tty -s && mesg n
    EOT

    service lightdm restart

    # x2go
    apt-get install -y software-properties-common dirmngr
    apt-key adv --recv-keys --keyserver keys.gnupg.net 7CDE3A860A53F9FD
    apt-key adv --recv-keys --keyserver keys.gnupg.net E1F958385BFE2B6E
    release=`lsb_release -a | grep Codename | awk '{print $2}'`

    echo "deb http://packages.x2go.org/debian $release extras main" > /etc/apt/sources.list.d/x2go.list
    echo "deb-src http://packages.x2go.org/debian $release extras main" >> /etc/apt/sources.list.d/x2go.list

    # add-apt-repository -y ppa:x2go/stable
    apt-get update
    apt-get install -y x2goserver x2goserver-xsession
fi



