mkdir -p tmp && /usr/bin/rm -f tmp/host-*.cpu_max
cat ../host.all.sort.uniq | xargs -n 1 -P 10 ./gethostcpu.sh 
cd tmp && ls host-*.cpu_max | xargs -n 1 -P 10 ../max.jq.sh 
sed '' host-*.cpu_max.value > ../host_cpu_max_1603440000_1603944000.data
#/usr/bin/rm -rf h c
