#!/bin/bash
apt-get install -y wget git python3 python3-pip
wget -O /usr/local/bin/k3s https://github.com/rancher/k3s/releases/download/v0.8.1/k3s
chmod +x /usr/local/bin/k3s
wget -O /lib/systemd/system/k3s.service https://raw.githubusercontent.com/rancher/k3s/master/k3s.service
mkdir -p /etc/systemd/system
touch /etc/systemd/system/k3s.service.env
systemctl daemon-reload
systemctl enable k3s
service k3s start
sleep 20
export K3STOKEN=`cat /var/lib/rancher/k3s/server/node-token`
export TOKEN="${GOT_TOKEN}"
/opt/got/goterra-cli --url ${GOT_URL} --deployment ${GOT_DEP} --token $TOKEN put k3stoken ${K3STOKEN}
