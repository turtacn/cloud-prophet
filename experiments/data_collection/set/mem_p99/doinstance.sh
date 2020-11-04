./getinstancesmem.sh
cd tmp
ls i-*.mem_p99 | xargs -n 1 -P 10 ../max.jq.sh
ls c-*.mem_p99 | xargs -n 1 -P 10 ../max.jq.sh
ls d-*.mem_p99 | xargs -n 1 -P 10 ../max.jq.sh
ls pod-*.mem_p99 | xargs -n 1 -P 10 ../max.jq.sh
sed '' i-*.mem_p99.value > ../instances_mem_p99_1603440000_1603944000.data
sed '' d-*.mem_p99.value >> ../instances_mem_p99_1603440000_1603944000.data
sed '' c-*.mem_p99.value >> ../instances_mem_p99_1603440000_1603944000.data
sed '' pod-*.mem_p99.value >> ../instances_mem_p99_1603440000_1603944000.data
