./getinstancesmem.sh
cd tmp
ls i-*.mem_max | xargs -n 1 -P 10 ../max.jq.sh
ls c-*.mem_max | xargs -n 1 -P 10 ../max.jq.sh
ls d-*.mem_max | xargs -n 1 -P 10 ../max.jq.sh
ls pod-*.mem_max | xargs -n 1 -P 10 ../max.jq.sh
sed '' i-*.mem_max.value > ../instances_mem_max_1603440000_1603944000.data
sed '' d-*.mem_max.value >> ../instances_mem_max_1603440000_1603944000.data
sed '' c-*.mem_max.value >> ../instances_mem_max_1603440000_1603944000.data
sed '' pod-*.mem_max.value >> ../instances_mem_max_1603440000_1603944000.data
