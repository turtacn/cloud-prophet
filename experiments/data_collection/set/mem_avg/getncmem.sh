curl -w "\n"  -X POST   http://r.horaedb-vm.xdfdfker892.jdcloud.com/api/query  -H "Content-Type:application/json; charset=utf-8"  -d "{    
	\"start\": 1603440000,
	\"end\": 1603944000,
	\"queries\": [{
		\"aggregator\": \"avg\",
        \"downsample\": \"10m-avg\",
		\"filters\": [{
			\"groupBy\": true,
			\"filter\": \"$1\",
			\"type\": \"literal_or\",
			\"tagk\": \"resourceId\"
		}],
		\"metric\": \"nativecontainer.memory.usage\"
	}]
}" -o tmp/$1.mem_avg
