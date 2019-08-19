#!/bin/bash

nbslave=100

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


master=`hostname`
masterip=`ip route get 1 | awk '{print $NF;exit}'`
cat > /etc/slurm-llnl/slurm.conf << EOL
ClusterName=cloud
ControlMachine=${master}
ControlAddr=${masterip}
MailProg=/usr/sbin/sendmail
MpiDefault=none
#MpiParams=ports=#-#
ProctrackType=proctrack/cgroup
ReturnToService=2
SlurmctldPidFile=/var/run/slurm-llnl/slurmctld.pid
#SlurmctldPort=6817
SlurmdPidFile=/var/run/slurm-llnl/slurmd.pid
#SlurmdPort=6818
SlurmdSpoolDir=/var/spool/slurmd
SlurmUser=slurm
#SlurmdUser=root
StateSaveLocation=/var/spool/slurm
SwitchType=switch/none
TaskPlugin=task/cgroup

# SCHEDULING
FastSchedule=1
SchedulerType=sched/backfill
SelectType=select/cons_res
SelectTypeParameters=CR_Core_Memory

# LOGGING AND ACCOUNTING
AccountingStorageType=accounting_storage/filetxt
AccountingStorageLoc=/var/log/slurm-llnl/acct
SlurmctldDebug=4
SlurmctldLogFile=/var/log/slurm-llnl/slurmctld.log
SlurmdLogFile=/var/log/slurm-llnl/slurmd.log

DefMemPerCPU=1000
PartitionName=cloud Nodes=slurm-[1-${nbslave}] Default=YES MaxTime=INFINITE State=UP
EOL

cpu=`grep -c ^processor /proc/cpuinfo`
ram=`grep MemTotal /proc/meminfo | awk '{print $2/1000}'`
rammb=`echo $ram | cut -f1 -d"."`

for i in $(seq 1 ${nbslave})
do
  echo "NodeName=slave-${i} CPUs=${cpu} CoresPerSocket=${cpu} ThreadsPerCore=1 RealMemory=${rammb} State=UNKNOWN " >> /etc/slurm-llnl/slurm.conf
done

service slurmctld start
# service slurmd start

export CONFIG=`cat /etc/slurm-llnl/slurm.conf`
export MUNGE=̀`cat /etc/munge/munge.key | base64`
export TOKEN="${GOT_TOKEN}"
/opt/got/goterra-cli --url ${GOT_URL} --deployment ${GOT_DEP} --token $TOKEN put slurm_config ${CONFIG}
/opt/got/goterra-cli --url ${GOT_URL} --deployment ${GOT_DEP} --token $TOKEN put slurm_munge ${MUNGE}
