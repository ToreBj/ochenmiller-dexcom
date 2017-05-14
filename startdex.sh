#!/bin/bash
# PFM start up dexterity
cd /home/pi/newopenaps
if `pgrep dexterity.sh > /dev/null` 
then echo 
     echo "`date`: dexterity running" 
else echo 
     echo "starting dexterity"
     ./dexterity.sh $1 /dev/wixel >> dexterity.log &
fi
