#!/bin/bash

set -e

apt-get update
apt-get install -y nginx jq curl snapd

export PATH=$PATH:/snap/bin
echo "Installing kubernetes cluster"
snap install microk8s --classic
snap install kubectl --classic
mkdir -p $HOME/.kube
microk8s.kubectl config view --raw > $HOME/.kube/config

if [ -e /home/debian ]; then
  mkdir -p /home/debian/.kube
  cp $HOME/.kube/config /home/debian/.kube/
fi

microk8s.enable storage
microk8s.enable rbac
microk8s.enable dns

kubetoken=$(microk8s.kubectl -n kube-system get secret | grep default-token | cut -d " " -f1)


echo "Installing Helm"
curl https://raw.githubusercontent.com/kubernetes/helm/master/scripts/get | bash
wget https://get.helm.sh/helm-v2.15.2-linux-amd64.tar.gz
tar xvfz helm-v2.15.2-linux-amd64.tar.gz
cp linux-amd64/helm /usr/local/bin/
cp linux-amd64/tiller /usr/local/bin/
kubectl --namespace kube-system create serviceaccount tiller
kubectl create clusterrolebinding tiller --clusterrole cluster-admin --serviceaccount=kube-system:tiller
helm init --service-account tiller --wait
kubectl patch deployment tiller-deploy --namespace=kube-system --type=json --patch='[{"op": "add", "path": "/spec/template/spec/containers/0/command", "value": ["/tiller", "--listen=localhost:44134"]}]'


echo "Deploying JupyterHub Helm chart"
HUBTOKEN=`openssl rand -hex 32`
PASSWORD=`openssl rand -hex 12`

cat >config.yaml <<EOL
proxy:
        secretToken: "${HUBTOKEN}"
auth:
  type: dummy
  dummy:
    password: "${PASSWORD}"
  whitelist:
    users:
      - admin
singleuser:
  cloudMetadata:
    enabled: true
  storage:
    dynamic:
      storageClass: microk8s-hostpath

  # Multi profile
  image:
    name: jupyter/minimal-notebook
    tag: 2343e33dec46
    profileList:
      - display_name: "Minimal environment"
        description: "To avoid too much bells and whistles: Python."
        default: true
      - display_name: "Datascience environment"
        description: "If you want the additional bells and whistles: Python, R, and Julia."
        kubespawner_override:
          image: jupyter/datascience-notebook:2343e33dec46
      - display_name: "Learning Data Science"
        description: "Datascience Environment with Sample Notebooks"
        kubespawner_override:
          image: jupyter/datascience-notebook:2343e33dec46
          lifecycle_hooks:
            postStart:
              exec:
                command:
                  - "sh"
                  - "-c"
                  - >
                    gitpuller https://github.com/data-8/materials-fa17 master materials-fa;
EOL


helm repo add jupyterhub https://jupyterhub.github.io/helm-chart/
helm repo update

RELEASE=jhub
NAMESPACE=jhub
helm upgrade --install $RELEASE jupyterhub/jupyterhub \
  --namespace $NAMESPACE  \
  --version=0.8.2 \
  --values config.yaml

echo "Patching JupyterHub for Kubernetes version"
kubectl patch deploy -n jhub hub --type json --patch '[{"op": "replace", "path": "/spec/template/spec/containers/0/command", "value": ["bash", "-c", "\nmkdir -p ~/hotfix\ncp -r /usr/local/lib/python3.6/dist-packages/kubespawner ~/hotfix\nls -R ~/hotfix\npatch ~/hotfix/kubespawner/spawner.py << EOT\n72c72\n<             key=lambda x: x.last_timestamp,\n---\n>             key=lambda x: x.last_timestamp and x.last_timestamp.timestamp() or 0.,\nEOT\n\nPYTHONPATH=$HOME/hotfix jupyterhub --config /srv/jupyterhub_config.py --upgrade-db\n"]}]'

echo "Deploy nginx config"
clusterip=`kubectl get services proxy-public --namespace jhub  -o json | jq '.spec.clusterIP' --raw-output`

cat >/etc/nginx/sites-enabled/default <<EOL
map \$http_upgrade \$connection_upgrade {
    default upgrade;
    ''      close;
}

server {
    listen 80;
    listen [::]:80 default_server;
    server_name _;

    location / {
        proxy_pass http://${clusterip}:80;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header Host \$host;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;

        # websocket headers
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection \$connection_upgrade;
    }
}
EOL

echo "Restarting nginx"
service nginx restart

echo "JupyterHub deployed"

if [ -e /opt/got/goterra-cli ]; then
  /opt/got/goterra-cli --url ${GOT_URL} --deployment ${GOT_DEP} --token $TOKEN put jhub_login admin
  /opt/got/goterra-cli --url ${GOT_URL} --deployment ${GOT_DEP} --token $TOKEN put jhub_password ${PASSWORD}
fi