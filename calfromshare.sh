#!/bin/bash
openaps use dexcom iter_calibrations 1 > raw-cgm/latestcal.json
cat raw-cgm/latestcal.json | grep -e scale -e slope -e intercept -e system_time -e meter
echo "./pushderivedcal.sh calfromshare.js `jq .[0].slope raw-cgm/latestcal.json|xargs printf %.0f` `jq .[0].intercept raw-cgm/latestcal.json|xargs printf %.0f` `jq .[0].scale raw-cgm/latestcal.json|xargs printf %.2f` `date "+%s"000`"
echo
