#!/bin/bash

ID=$1
PORT=$2
LOG=$3

while read line; do 
      if [[ $line == *"$ID"* ]]; then
	TRANSID=`echo $line  | cut -d " " -f 1 | tr -d '\n'`
	RAWBG=`echo $line    | cut -d " " -f 2 | tr -d '\n'`
        FILTERED=`echo $line | cut -d " " -f 3 | tr -d '\n'`
        BATTERY=`echo $line  | cut -d " " -f 4 | tr -d '\n'`
        TRANSREC=`echo $line | cut -d " " -f 6 | tr -d '\n'`
	DATE=`date -Iseconds`
        echo
	echo ./hrawtobg.sh "$RAWBG" \""$DATE"\" Flat $TRANSREC $FILTERED $BATTERY $TRANSID
	     ./hrawtobg.sh "$RAWBG" "$DATE" Flat $TRANSREC $FILTERED $BATTERY $TRANSID
	echo "$line $DATE"
        rm -f /home/pi/myopenaps/monitor/glucose.json && /home/pi/myopenaps/getdxglucose.sh
      fi 
    done < $PORT
