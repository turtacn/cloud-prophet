mkdir -p tmp/
curl -w "\n"  -X POST  http://horaedb-read.jcloud.com/api/query  -H "Content-Type:application/json; charset=utf-8"  -d "{
    \"start\": $1,
    \"end\": $2,
    \"queries\": [{
        \"aggregator\": \"max\",
        \"downsample\": \"1m-avg\",
        \"filters\": [{
            \"groupBy\": true,
            \"filter\": \"$3\",
            \"type\": \"literal_or\",
            \"tagk\": \"ns\"
        }],
        \"metric\": \"mem.use.percent\"
    }]
}" -o tmp/host-$1-$2-$3.mem_util_percent

curl -w "\n"  -X POST  http://horaedb-read.jcloud.com/api/query  -H "Content-Type:application/json; charset=utf-8"  -d "{
    \"start\": $1,
    \"end\": $2,
    \"queries\": [{
        \"aggregator\": \"max\",
        \"downsample\": \"1m-avg\",
        \"filters\": [{
            \"groupBy\": true,
            \"filter\": \"$3\",
            \"type\": \"literal_or\",
            \"tagk\": \"ns\"
        }],
        \"metric\": \"cpu.use\"
    }]
}" -o tmp/host-$1-$2-$3.cpu_util_percent

curl -w "\n"  -X POST  http://horaedb-read.jcloud.com/api/query  -H "Content-Type:application/json; charset=utf-8"  -d "{
    \"start\": $1,
    \"end\": $2,
    \"queries\": [{
        \"aggregator\": \"max\",
        \"downsample\": \"1m-avg\",
        \"tags\":{\"iface\":\"eth0\",\"ns\":\"$3\"},
        \"filters\": [{
            \"groupBy\": true,
            \"filter\": \"$3\",
            \"type\": \"literal_or\",
            \"tagk\": \"ns\"
        }],
        \"metric\": \"net.out.bps\"
    }]
}" -o tmp/host-$1-$2-$3.net_out


curl -w "\n"  -X POST  http://horaedb-read.jcloud.com/api/query  -H "Content-Type:application/json; charset=utf-8"  -d "{
    \"start\": $1,
    \"end\": $2,
    \"queries\": [{
        \"aggregator\": \"max\",
        \"downsample\": \"1m-avg\",
        \"tags\":{\"iface\":\"eth0\",\"ns\":\"$3\"},
        \"filters\": [{
            \"groupBy\": true,
            \"filter\": \"$3\",
            \"type\": \"literal_or\",
            \"tagk\": \"ns\"
        }],
        \"metric\": \"net.in.bps\"
    }]
}" -o tmp/host-$1-$2-$3.net_in

curl -w "\n"  -X POST  http://horaedb-read.jcloud.com/api/query  -H "Content-Type:application/json; charset=utf-8"  -d "{
    \"start\": $1,
    \"end\": $2,
    \"queries\": [{
        \"aggregator\": \"max\",
        \"downsample\": \"1m-avg\",
        \"tags\":{\"device\":\"sda\",\"ns\":\"$3\"},
        \"filters\": [{
            \"groupBy\": true,
            \"filter\": \"$3\",
            \"type\": \"literal_or\",
            \"tagk\": \"ns\"
        }],
        \"metric\": \"disk.io.util\"
    }]
}" -o tmp/host-$1-$2-$3.disk_usage_percent

cat tmp/host-$1-$2-$3.cpu_util_percent    |  jq .[].dps | jq .[] > c1-$1-$2-$3
cat tmp/host-$1-$2-$3.mem_util_percent    |  jq .[].dps | jq .[] > c2-$1-$2-$3
cat tmp/host-$1-$2-$3.net_in              |  jq .[].dps | jq .[] > c3-$1-$2-$3
cat tmp/host-$1-$2-$3.net_out             |  jq .[].dps | jq .[] > c4-$1-$2-$3
cat tmp/host-$1-$2-$3.disk_usage_percent  |  jq .[].dps | jq .[] > c5-$1-$2-$3
cat tmp/host-$1-$2-$3.disk_usage_percent  |  jq .[].dps | jq  -r keys[] > k-$1-$2-$3
paste -d  ','  k-$1-$2-$3 c1-$1-$2-$3 c2-$1-$2-$3 c3-$1-$2-$3 c4-$1-$2-$3 c5-$1-$2-$3  > host-$1-$2-$3-usage.csv
rm c1-$1-$2-$3 c2-$1-$2-$3 c3-$1-$2-$3 c4-$1-$2-$3 c5-$1-$2-$3 k-$1-$2-$3
