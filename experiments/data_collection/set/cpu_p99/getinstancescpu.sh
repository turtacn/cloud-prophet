# vm instance
cut -d ' ' -f 2 ../instances | grep -o 'i-..........' | xargs -n 1 -P 10 ./getvmcpu.sh
# docker instance
cut -d ' ' -f 2 ../instances | grep -o 'd-..........' | xargs -n 1 -P 10 ./getdockercpu.sh
# pod instance
cut -d ' ' -f 2 ../instances | grep -o 'pod-..........' | xargs -n 1 -P 10 ./getpodcpu.sh
# nc instance
cut -d ' ' -f 2 ../instances | grep -o 'c-..........' | xargs -n 1 -P 10 ./getnccpu.sh

