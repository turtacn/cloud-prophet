./getinstancescpu.sh
cd tmp
ls i-*.cpu_p95 | xargs -n 1 -P 10 ../p95.jq.sh
ls c-*.cpu_p95 | xargs -n 1 -P 10 ../p95.jq.sh
ls d-*.cpu_p95 | xargs -n 1 -P 10 ../p95.jq.sh
ls pod-*.cpu_p95 | xargs -n 1 -P 10 ../p95.jq.sh
sed '' i-*.cpu_p95.value > ../instances_cpu_p95_1603440000_1603944000.data
sed '' d-*.cpu_p95.value >> ../instances_cpu_p95_1603440000_1603944000.data
sed '' c-*.cpu_p95.value >> ../instances_cpu_p95_1603440000_1603944000.data
sed '' pod-*.cpu_p95.value >> ../instances_cpu_p95_1603440000_1603944000.data
#paste i c > instance_cpu_p95_1603440000_1603944000.data
#/usr/bin/rm -rf  i c
