#!/bin/bash

export DEBIAN_FRONTEND=noninteractive
apt-get update

# Consul install (DNS)
CONSUL_VERSION=1.5.3
apt-get install -y unzip wget bind9 bind9utils
wget https://releases.hashicorp.com/consul/${CONSUL_VERSION}_/consul_${CONSUL_VERSION}_linux_amd64.zip
unzip consul_${CONSUL_VERSION}__linux_amd64.zip
chmod +x ./consul
mv ./consul /usr/bin/
rm consul_${CONSUL_VERSION}__linux_amd64.zip
mkdir -p /opt/consul

cat > /etc/bind/named.conf.options << EOL
options {
   directory "/var/cache/bind";
   recursion yes;
   allow-query { localhost; };
   
   forwarders {
      8.8.8.8;
      8.8.4.4;
   };
   dnssec-enable no;
   dnssec-validation no;
   auth-nxdomain no; # conform to RFC1035
   listen-on-v6 { any; };
};
include "/etc/bind/consul.conf";
EOL

cat > /etc/bind/consul.conf << EOL
zone "consul" IN {
   type forward;
   forward only;
   forwarders { 127.0.0.1 port 8600; };
};
EOL

service bind9 restart

cat > /lib/systemd/system/consul.service << EOL
# Consul systemd service unit file
[Unit]
Description=Consul Service Discovery Agent
Documentation=https://www.consul.io/
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStart=/usr/bin/consul agent \
        --server \
        --data-dir /opt/consul \
        --bootstrap-expect 1

ExecReload=/bin/kill -HUP $MAINPID
KillSignal=SIGINT
TimeoutStopSec=5
Restart=on-failure
SyslogIdentifier=consul

[Install]
WantedBy=multi-user.target
EOL

systemctl daemon-reload
service consul start

cat > /etc/resolv.conf << EOL
domain node.consul
search node.consul
nameserver 127.0.0.1
EOL

consulip=`ip route get 1 | awk '{print $NF;exit}'`
/opt/got/goterra-cli --url ${GOT_URL} --deployment ${GOT_DEP} --token $TOKEN put consul_advertise ${consulip}
