mkdir -p tmp
./getinstancescpu.sh
cd tmp
ls i-*.cpu_avg | xargs -n 1 -P 10 ../avg.jq.sh
ls c-*.cpu_avg | xargs -n 1 -P 10 ../avg.jq.sh
ls d-*.cpu_avg | xargs -n 1 -P 10 ../avg.jq.sh
ls pod-*.cpu_avg | xargs -n 1 -P 10 ../avg.jq.sh
sed '' i-*.cpu_avg.value > ../instances_cpu_avg_1603440000_1603944000.data
sed '' d-*.cpu_avg.value >> ../instances_cpu_avg_1603440000_1603944000.data
sed '' c-*.cpu_avg.value >> ../instances_cpu_avg_1603440000_1603944000.data
sed '' pod-*.cpu_avg.value >> ../instances_cpu_avg_1603440000_1603944000.data
#paste i c > instance_cpu_avg_1603440000_1603944000.data
#/usr/bin/rm -rf  i c
