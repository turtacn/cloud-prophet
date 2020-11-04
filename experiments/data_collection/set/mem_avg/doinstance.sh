./getinstancesmem.sh
cd tmp
ls i-*.mem_avg | xargs -n 1 -P 10 ../avg.jq.sh
ls c-*.mem_avg | xargs -n 1 -P 10 ../avg.jq.sh
ls d-*.mem_avg | xargs -n 1 -P 10 ../avg.jq.sh
ls pod-*.mem_avg | xargs -n 1 -P 10 ../avg.jq.sh
sed '' i-*.mem_avg.value > ../instances_mem_avg_1603440000_1603944000.data
sed '' d-*.mem_avg.value >> ../instances_mem_avg_1603440000_1603944000.data
sed '' c-*.mem_avg.value >> ../instances_mem_avg_1603440000_1603944000.data
sed '' pod-*.mem_avg.value >> ../instances_mem_avg_1603440000_1603944000.data
