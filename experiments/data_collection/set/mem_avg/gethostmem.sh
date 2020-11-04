curl -w "\n"  -X POST  http://horaedb-read.jcloud.com/api/query  -H "Content-Type:application/json; charset=utf-8"  -d "{    
    \"start\": 1603440000,
    \"end\":1603944000,
    \"queries\": [{
        \"aggregator\": \"avg\",
        \"downsample\": \"10m-avg\",
        \"filters\": [{
            \"groupBy\": true,
            \"filter\": \"$1\",
            \"type\": \"literal_or\",
            \"tagk\": \"ns\"
        }],
        \"metric\": \"mem.use.percent\"
    }]
}" -o tmp/host-$1.mem_avg
