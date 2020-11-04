export JVIRT_ADMIN_TOKEN=07765634-933c-4cea-9f63-a586cd15cbb8
grep -oE "\b([0-9]{1,3}\.){3}[0-9]{1,3}\b" host.all | sort -u > host.all.sort.uniq
cat host.all.sort.uniq | xargs -n 1 -P 1 jvirt instance-list  -a  --host-ip > jcs.instances
cat host.all.sort.uniq | xargs -n 1 -P 1 jvirt container-list -a  --host-ip >> jcs.instances
cat host.all.sort.uniq | xargs -n 1 -P 1 jvirt-jks pod-list   -a  --host-ip > jks.instances
cat host.all.sort.uniq | xargs -n 1 -P 1 jvirt-jks nc-list    -a  --host-ip >> jks.instances
