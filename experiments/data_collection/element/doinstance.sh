#2020/11/3 19:0:0
#override endtime
Now=`date +%s`
D=`date --date=@${Now}`
Before=`date -d "$D - 3 hours" +%s`
COUNTER=0
while [ $COUNTER -lt 100 ]
do
    COUNTER=`expr $COUNTER + 1`
    echo $Before $Now
    ./utility-getall-instance.sh $Before $Now $1
    Now=$Before
    D=`date --date=@${Now}`
    Before=`date -d "$D - 3 hours" +%s`
done
