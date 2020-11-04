./getinstancescpu.sh
cd tmp
ls i-*.cpu_p99 | xargs -n 1 -P 10 ../p99.jq.sh
ls c-*.cpu_p99 | xargs -n 1 -P 10 ../p99.jq.sh
ls d-*.cpu_p99 | xargs -n 1 -P 10 ../p99.jq.sh
ls pod-*.cpu_p99 | xargs -n 1 -P 10 ../p99.jq.sh
sed '' i-*.cpu_p99.value > ../instances_cpu_p99_1603440000_1603944000.data
sed '' d-*.cpu_p99.value >> ../instances_cpu_p99_1603440000_1603944000.data
sed '' c-*.cpu_p99.value >> ../instances_cpu_p99_1603440000_1603944000.data
sed '' pod-*.cpu_p99.value >> ../instances_cpu_p99_1603440000_1603944000.data
#paste i c > instance_cpu_p99_1603440000_1603944000.data
#/usr/bin/rm -rf  i c
