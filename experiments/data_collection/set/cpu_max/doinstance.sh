./getinstancescpu.sh
cd tmp
ls i-*.cpu_max | xargs -n 1 -P 10 ../max.jq.sh
ls c-*.cpu_max | xargs -n 1 -P 10 ../max.jq.sh
ls d-*.cpu_max | xargs -n 1 -P 10 ../max.jq.sh
ls pod-*.cpu_max | xargs -n 1 -P 10 ../max.jq.sh
sed '' i-*.cpu_max.value > ../instances_cpu_max_1603440000_1603944000.data
sed '' d-*.cpu_max.value >> ../instances_cpu_max_1603440000_1603944000.data
sed '' c-*.cpu_max.value >> ../instances_cpu_max_1603440000_1603944000.data
sed '' pod-*.cpu_max.value >> ../instances_cpu_max_1603440000_1603944000.data
#paste i c > instance_cpu_max_1603440000_1603944000.data
#/usr/bin/rm -rf  i c
