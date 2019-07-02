#!/bin/bash      
apt-get install -y wget git python3 python3-pip
wget -O /usr/local/bin/k3s https://github.com/rancher/k3s/releases/download/v0.5.0/k3s
chmod +x /usr/local/bin/k3s
wget -O /lib/systemd/system/k3s.service https://raw.githubusercontent.com/rancher/k3s/master/k3s.service
mkdir -p /etc/systemd/system
touch /etc/systemd/system/k3s.service.env
export TOKEN="${GOT_TOKEN}"
masterip=`/opt/got/goterra-cli --deployment ${GOT_DEP} --url ${GOT_URL} --token $TOKEN get masterip`
mastertoken=`/opt/got/goterra-cli --deployment ${GOT_DEP} --url ${GOT_URL} --token $TOKEN get k3stoken`

sed -i "s;ExecStart=/usr/local/bin/k3s server;ExecStart=/usr/local/bin/k3s agent --server https://$masterip:6443 --token ${mastertoken};g" /lib/systemd/system/k3s.service
systemctl daemon-reload
systemctl enable k3s
service k3s start