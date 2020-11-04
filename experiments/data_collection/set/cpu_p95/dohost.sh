mkdir -p tmp && /usr/bin/rm -f tmp/host-*.cpu_p95
cat ../host.all.sort.uniq | xargs -n 1 -P 10 ./gethostcpu.sh 
cd tmp && ls host-*.cpu_p95 | xargs -n 1 -P 10 ../p95.jq.sh 
sed '' host-*.cpu_p95.value > ../host_cpu_p95_1603440000_1603944000.data
#/usr/bin/rm -rf h c
