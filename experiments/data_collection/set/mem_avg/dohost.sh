mkdir -p tmp && /usr/bin/rm -f tmp/host-*.mem_avg
cat ../host.all.sort.uniq | xargs -n 1 -P 10 ./gethostmem.sh 
cd tmp && ls host-*.mem_avg | xargs -n 1 -P 10 ../avg.jq.sh 
sed '' host-*.mem_avg.value > ../host_mem_avg_1603440000_1603944000.data
#/usr/bin/rm -rf h c
