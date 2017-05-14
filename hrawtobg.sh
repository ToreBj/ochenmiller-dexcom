pi@mlkh3:~/newopenaps $ cat hrawtobg.sh 
#!/bin/bash
# dexterity record uploader
# takes raw/filtered, direction, transmitter record number (0-63 = 05:20 loop. Handy for determining missed records)
# calibration data and mongo credentials are in dexterity.cfg
#
RAW=$1
AT=$2
DIRECTION=$3
TRANSREC=$4
FILTERED=$5
BATTERY=$6
TRANSID=$7
#
#node grablatestcal.js > cal.tmp
curl -s -k -m 30 https://XXX.azurewebsites.net/pebble | sed 's/.*slope/slope/' | sed 's/\"//g' | sed 's/}//g' | sed 's/]//g' | sed 's/:/=/g' | sed 's/,/\n/g' > cal.test
if grep scale cal.test > /dev/null; then mv cal.test cal.cfg; else echo "no cal record to grab from azure"; fi

source mongo.cfg
source cal.cfg
source thist.cfg
#

# create SGV from raw/filtered and calibration data
SGV=`awk -v r=$RAW -v f=$FILTERED -v s=$slope -v i=$intercept -v c=$scale 'BEGIN { printf "%2.0f\n", c*(((r+r+r+r+f)/5)-i)/s }'`

if [ $SGV -lt 40 ]
then
echo "Bad site? SGV=$SGV"
exit
fi

OPTS=""
if [ -n "$AT" ]; then
  OPTS="--date ${AT}"
fi
#ISO=$(date -Iseconds $OPTS)
#UNIX=$(date +%s $OPTS)000
ISO=$(date -Iseconds --date="$2")
UNIX=$(date +%s --date="$2")000
#RECKEY=`echo $UNIX | cut -c1-6`
#RECKEY=`expr $UNIX / 300000`
RECKEY=`expr \( $UNIX - 60000 \) / 300000`

#QUERY="%7B\"unfiltered\"%3A\"$RAW\"%2C\"filtered\"%3A\"$FILTERED\"%2C\"transrec\"%3A\"$TRANSREC\"%2C\"reckey\"%3A$RECKEY%7D"
QUERY="%7B\"reckey\"%3A$RECKEY%7D"
echo $QUERY

# derive DIRECTION from SGV and historical records

let DIFF=$SGV-$SGV1
echo DIFFERENCE=$DIFF
let TIMEDIFF=$UNIX-$DATE1
echo TIMEDIFF=$TIMEDIFF

PROPDIFF=`awk -v t=$TIMEDIFF -v d=$DIFF 'BEGIN { printf "%3.0f\n", ((d/t)*300000) }'`
echo PROPDIFF=$PROPDIFF

if [ $PROPDIFF -lt -15 ]
then
DIRECTION="DoubleDown"
TREND=7
fi

if [ $PROPDIFF -ge -15 ]
then
DIRECTION="SingleDown"
TREND=6
fi

if [ $PROPDIFF -ge -10 ]
then
DIRECTION="FortyFiveDown"
TREND=5
fi

if [ $PROPDIFF -ge -5 ]
then
DIRECTION="Flat"
TREND=4
fi

if [ $PROPDIFF -gt 5 ]
then
DIRECTION="FortyFiveUp"
TREND=3
fi

if [ $PROPDIFF -gt 10 ]
then
DIRECTION="SingleUp"
TREND=2
fi

if [ $PROPDIFF -gt 15 ]
then
DIRECTION="DoubleUp"
TREND=1
fi

#
set=\$set

DATA=`(
cat <<EOF
{ "$set" :  { "sgv"        :  $SGV        ,
              "unfiltered" : "$RAW"       ,
              "filtered"   : "$FILTERED"  ,
              "device"     : "kh"         ,
              "kh"         : "kh"         ,
              "psgv"       :  $SGV        ,
              "direction"  : "$DIRECTION" ,
              "trend"      : $TREND     ,
              "dateString" : "$ISO"       ,
              "date"       :  $UNIX       ,
              "battery"    : "$BATTERY"   ,
              "transid"    : "$TRANSID"   ,
              "transrec"   : "$TRANSREC"  ,
              "slope"      : "$slope"     ,
              "intercept"  : "$intercept" ,
              "scale"      : "$scale"     ,     
              "type"       : "sgv"        }
}
EOF
) | awk 1 ORS=' '`
echo $DATA
echo $DATA | sed 's/{ "$set" : {/{/' | sed 's/} }/}/' >> dexterity.json
echo

echo "SGV1=$SGV"     > thist.cfg
echo "SGV2=$SGV1"   >> thist.cfg
echo "SGV3=$SGV2"   >> thist.cfg
echo "DATE1=$UNIX"  >> thist.cfg
echo "DATE2=$DATE1" >> thist.cfg
echo "DATE3=$DATE2" >> thist.cfg

#echo "$QUERY $DATA" | sed 's/hj/idex/g'  > /dev/ttyAMA0
echo "$QUERY $DATA"

curl -s -k -m 30 -H "Accept: application/json" -H "Content-type: application/json" -X PUT -d "$DATA" "https://api.mongolab.com/api/1/databases/XXX?apiKey=$APIKEY&q=$QUERY&u=true"
echo
