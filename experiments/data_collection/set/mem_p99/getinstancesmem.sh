# vm instance
cut -d ' ' -f 2 ../instances | grep -o 'i-..........' | xargs -n 1 -P 10 ./getvmmem.sh
# docker instance
cut -d ' ' -f 2 ../instances | grep -o 'd-..........' | xargs -n 1 -P 10 ./getdockermem.sh
# pod instance
cut -d ' ' -f 2 ../instances | grep -o 'pod-..........' | xargs -n 1 -P 10 ./getpodmem.sh
# nc instance
cut -d ' ' -f 2 ../instances | grep -o 'c-..........' | xargs -n 1 -P 10 ./getncmem.sh

