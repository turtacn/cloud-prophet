#!/bin/sh
mkdir -p tmp/

#默认指标，vm
cpu_util_percent="cpu_util"
mem_util_percent="memory.usage"
net_in="vm.network.dev.bytes.in"
net_out="vm.network.dev.bytes.out"
disk_usage_percent_read="vm.disk.dev.bytes.read"
disk_usage_percent_write="vm.disk.dev.bytes.write"
instance="vm"
if [ -z "$3" ]
then
      exit
else
    instype=$(echo $3 | cut -d"-" -f 1)
    case "$instype" in
              "i")   instance="vm"      ;;
              "d")   instance="docker"  ;;
              "c")   instance="nc"      ;;
              "pod") instance="pod"     ;;
              *) printf "非法实例==>$3\n\n"; exit 123 ;;
    esac
fi

curl -w "\n"  -X POST   http://r.horaedb-vm.xdfdfker892.jdcloud.com/api/query  -H "Content-Type:application/json; charset=utf-8"  -d "{
    \"start\": $1,
    \"end\": $2,
    \"queries\": [{
        \"aggregator\": \"max\",
        \"downsample\": \"1m-avg\",
        \"filters\": [{
            \"groupBy\": true,
            \"filter\": \"$3\",
            \"type\": \"literal_or\",
            \"tagk\": \"resourceId\"
        }],
        \"metric\": \"$cpu_util_percent\"
    }]
}" -o tmp/$instance-$1-$2-$3.cpu_util_percent

curl -w "\n"  -X POST   http://r.horaedb-vm.xdfdfker892.jdcloud.com/api/query  -H "Content-Type:application/json; charset=utf-8"  -d "{
    \"start\": $1,
    \"end\": $2,
    \"queries\": [{
        \"aggregator\": \"max\",
        \"downsample\": \"1m-avg\",
        \"filters\": [{
            \"groupBy\": true,
            \"filter\": \"$3\",
            \"type\": \"literal_or\",
            \"tagk\": \"resourceId\"
        }],
        \"metric\": \"$mem_util_percent\"
    }]
}" -o tmp/$instance-$1-$2-$3.mem_util_percent

curl -w "\n"  -X POST   http://r.horaedb-vm.xdfdfker892.jdcloud.com/api/query  -H "Content-Type:application/json; charset=utf-8"  -d "{
    \"start\": $1,
    \"end\": $2,
    \"queries\": [{
        \"aggregator\": \"max\",
        \"downsample\": \"1m-avg\",
        \"tags\":{\"devName\":\"eth0\",\"resourceId\":\"$3\"},
        \"filters\": [{
            \"groupBy\": true,
            \"filter\": \"$3\",
            \"type\": \"literal_or\",
            \"tagk\": \"resourceId\"
        }],
        \"metric\": \"$net_in\"
    }]
}" -o tmp/$instance-$1-$2-$3.net_in

curl -w "\n"  -X POST   http://r.horaedb-vm.xdfdfker892.jdcloud.com/api/query  -H "Content-Type:application/json; charset=utf-8"  -d "{
    \"start\": $1,
    \"end\": $2,
    \"queries\": [{
        \"aggregator\": \"max\",
        \"downsample\": \"1m-avg\",
        \"tags\":{\"devName\":\"eth0\",\"resourceId\":\"$3\"},
        \"filters\": [{
            \"groupBy\": true,
            \"filter\": \"$3\",
            \"type\": \"literal_or\",
            \"tagk\": \"resourceId\"
        }],
        \"metric\": \"$net_out\"
    }]
}" -o tmp/$instance-$1-$2-$3.net_out

curl -w "\n"  -X POST   http://r.horaedb-vm.xdfdfker892.jdcloud.com/api/query  -H "Content-Type:application/json; charset=utf-8"  -d "{
    \"start\": $1,
    \"end\": $2,
    \"queries\": [{
        \"aggregator\": \"max\",
        \"downsample\": \"1m-avg\",
        \"tags\":{\"devName\":\"/vda1\",\"resourceId\":\"$3\"},
        \"filters\": [{
            \"groupBy\": true,
            \"filter\": \"$3\",
            \"type\": \"literal_or\",
            \"tagk\": \"resourceId\"
        }],
        \"metric\": \"$disk_usage_percent_read\"
    }]
}" -o tmp/$instance-$1-$2-$3.disk_usage_percent_read

curl -w "\n"  -X POST   http://r.horaedb-vm.xdfdfker892.jdcloud.com/api/query  -H "Content-Type:application/json; charset=utf-8"  -d "{
    \"start\": $1,
    \"end\": $2,
    \"queries\": [{
        \"aggregator\": \"max\",
        \"downsample\": \"1m-avg\",
        \"tags\":{\"devName\":\"/vda1\",\"resourceId\":\"$3\"},
        \"filters\": [{
            \"groupBy\": true,
            \"filter\": \"$3\",
            \"type\": \"literal_or\",
            \"tagk\": \"resourceId\"
        }],
        \"metric\": \"$disk_usage_percent_write\"
    }]
}" -o tmp/$instance-$1-$2-$3.disk_usage_percent_write

cat tmp/$instance-$1-$2-$3.cpu_util_percent         |  jq .[].dps | jq .[] > c1-$instance-$1-$2-$3
cat tmp/$instance-$1-$2-$3.mem_util_percent         |  jq .[].dps | jq .[] > c2-$instance-$1-$2-$3
cat tmp/$instance-$1-$2-$3.net_in                   |  jq .[].dps | jq .[] > c3-$instance-$1-$2-$3
cat tmp/$instance-$1-$2-$3.net_out                  |  jq .[].dps | jq .[] > c4-$instance-$1-$2-$3
cat tmp/$instance-$1-$2-$3.disk_usage_percent_read  |  jq .[].dps | jq .[] > c5-$instance-$1-$2-$3
cat tmp/$instance-$1-$2-$3.disk_usage_percent_write |  jq .[].dps | jq .[] > c6-$instance-$1-$2-$3
cat tmp/$instance-$1-$2-$3.disk_usage_percent_write |  jq .[].dps | jq  -r keys[] > k-$instance-$1-$2-$3
paste -d  ','  \
    k-$instance-$1-$2-$3  \
    c1-$instance-$1-$2-$3 \
    c2-$instance-$1-$2-$3 \
    c3-$instance-$1-$2-$3 \
    c4-$instance-$1-$2-$3 \
    c5-$instance-$1-$2-$3 \
    c6-$instance-$1-$2-$3 > $instance-$1-$2-$3-usage.csv
rm  k-$instance-$1-$2-$3  \
    c1-$instance-$1-$2-$3 \
    c2-$instance-$1-$2-$3 \
    c3-$instance-$1-$2-$3 \
    c4-$instance-$1-$2-$3 \
    c5-$instance-$1-$2-$3 \
    c6-$instance-$1-$2-$3
