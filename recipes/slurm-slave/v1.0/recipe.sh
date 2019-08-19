#!/bin/bash 

export DEBIAN_FRONTEND=noninteractive
apt-get update
apt-get install -y mailutils slurm slurmd slurmctld sview

mkdir -p /var/spool/slurm
chown -R slurm /var/spool/slurm
mkdir -p /var/spool/slurmd
chown -R slurm /var/spool/slurmd

cat > /etc/slurm-llnl/cgroup.conf << EOL
CgroupAutomount=yes
ConstrainCores=yes
ConstrainRAMSpace=yes
EOL

export TOKEN="${GOT_TOKEN}"
CONFIG=`/opt/got/goterra-cli --url ${GOT_URL} --deployment ${GOT_DEP} --token $TOKEN put slurm_config`
MUNGE=`/opt/got/goterra-cli --url ${GOT_URL} --deployment ${GOT_DEP} --token $TOKEN put slurm_munge`
echo $MUNGE | base64 -d > /etc/munge/munge.key
service munge restart
echo $CONFIG > /etc/slurm-llnl/slurm.conf


service slurmd start
