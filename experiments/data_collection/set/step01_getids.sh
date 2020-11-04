export JVIRT_ADMIN_TOKEN=07765634-933c-4cea-9f63-a586cd15cbb8
jvirt host-list -a > host.all
grep -oE "\b([0-9]{1,3}\.){3}[0-9]{1,3}\b" host.all | sort -u > host.all.sort.uniq
/usr/bin/rm -rf tmp
mkdir -p tmp
cat host.all.sort.uniq | xargs -n 1 -P 10 ./jvirt_instance_list
sed '' tmp/host-*-vm.list >  jvirt.instances 
sed '' tmp/host-*-docker.list >> jvirt.instances
sed '' tmp/host-*-pod.list >> jvirt.instances
sed '' tmp/host-*-nc.list >> jvirt.instances
grep   ' i-\| d-\| c-\| pod-'  jvirt.instances > instances
