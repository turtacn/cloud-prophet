mkdir -p tmp && /usr/bin/rm -f tmp/host-*.mem_p99
cat ../host.all.sort.uniq | xargs -n 1 -P 10 ./gethostmem.sh 
cd tmp && ls host-*.mem_p99 | xargs -n 1 -P 10 ../max.jq.sh 
sed '' host-*.mem_p99.value > ../host_mem_p99_1603440000_1603944000.data
#/usr/bin/rm -rf h c
