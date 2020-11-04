curl -w "\n"  -X POST   http://r.horaedb-vm.xdfdfker892.jdcloud.com/api/query  -H "Content-Type:application/json; charset=utf-8"  -d "{    
	\"start\": 1603094400,
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
		\"metric\": \"cpu_util\"
	}]
}" -o tmp/$1.cpu_avg
