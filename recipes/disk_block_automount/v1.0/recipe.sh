#!/bin/bash

set -e

nbvd=0

for vd in /dev/vd*;
do
   vdtype=`file -s $vd`
   if [[ $vdtype =~ \:.data.* ]]; then
       echo "data volume to be created and mounted: $vd"
       nbvd=$((nbvd + 1))
       mkfs.ext4 -m 0 -F -E lazy_itable_init=0,lazy_journal_init=0,discard $vd
       mkdir -p /mnt/disks/data$nbvd
       mount -o discard,defaults $vd /mnt/disks/data$nbvd       
   fi
done

echo "Number of additional disks $nbvd"
if [ $nbvd -ge 1 ];then
    echo "Link first volume to /mnt/data"
    if [ ! -e /mnt/data ]; then
        ln -s /mnt/disks/data1 /mnt/data
    fi
fi

