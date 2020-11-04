#!/bin/bash

for dir in "cpu_avg" "cpu_max" "cpu_p95" "cpu_p99" "mem_avg" "mem_max" "mem_p95" "mem_p99"
do 
   cd $dir
   /usr/bin/rm   -rf   *.data     tmp
   ./dohost.sh >&log&
   ./doinstance.sh >&log&
   cd ../
done 
tail  -f  cpu_*/log  mem_*/log
