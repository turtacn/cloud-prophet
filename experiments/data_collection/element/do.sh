#2020/11/3 19:0:0
Now=1604401200 #`date +%s`
D=`date --date=@${Now}`
Before=`date -d "$D - 3 hours" +%s`
COUNTER=0
while [ $COUNTER -lt 100 ]
do
    COUNTER=`expr $COUNTER + 1`
    echo $Before $Now
    ./utility-getall-host.sh $Before $Now 172.19.9.104
    Now=$Before
    D=`date --date=@${Now}`
    Before=`date -d "$D - 3 hours" +%s`
done
