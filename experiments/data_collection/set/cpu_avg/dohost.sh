mkdir -p tmp && /usr/bin/rm -f tmp/host-*.cpu_avg
cat ../host.all.sort.uniq | xargs -n 1 -P 10 ./gethostcpu.sh 
cd tmp && ls host-*.cpu_avg | xargs -n 1 -P 10 ../avg.jq.sh 
sed '' host-*.cpu_avg.value > ../host_cpu_avg_1603440000_1603944000.data
#/usr/bin/rm -rf h c
