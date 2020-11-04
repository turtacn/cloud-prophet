./getinstancesmem.sh
cd tmp
ls i-*.mem_p95 | xargs -n 1 -P 10 ../max.jq.sh
ls c-*.mem_p95 | xargs -n 1 -P 10 ../max.jq.sh
ls d-*.mem_p95 | xargs -n 1 -P 10 ../max.jq.sh
ls pod-*.mem_p95 | xargs -n 1 -P 10 ../max.jq.sh
sed '' i-*.mem_p95.value > ../instances_mem_p95_1603440000_1603944000.data
sed '' d-*.mem_p95.value >> ../instances_mem_p95_1603440000_1603944000.data
sed '' c-*.mem_p95.value >> ../instances_mem_p95_1603440000_1603944000.data
sed '' pod-*.mem_p95.value >> ../instances_mem_p95_1603440000_1603944000.data
