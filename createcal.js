#!/usr/bin/node
'use strict';
var sprintf=require("sprintf-js").sprintf;

var db = require("mongodb");
var uri = "mongodb://XXX";

// print process.argv
var hours = 72; //how many hours back

process.argv.forEach(function (val, index, array) {
//  console.log(index + ': ' + val);
  if ( index == 2 ) {
  hours   = parseFloat(val);
  }
});

var now = Date.now();
//time interval should be set to x days OR latest sensor restart + ??minutes
var interval = (now-(hours*3600*1000));
var mbgrecord = 0;
var calrecord = 0;
var currentsgv;
var nextsgv;
var currentmbg;
var insertdate;
var knownmbg = [0];
var knownfiltered = [30000];
//console.log('now= ' + now + '\n');
//console.log('twodaysago= ' + twodaysago + '\n');

db.MongoClient.connect(uri, function(err, db) {
    if (err) {
    	console.log("Error connecting to mongo, ERROR: %j", err);
    	throw err;
    }
    var ochenmiller = db.collection('ochenmiller');
      
    ochenmiller
    .find({ date : { "$exists": true, "$gt": interval } })
    .sort({ date: -1 })
//  .limit(6)
    .toArray
    (function (err, recs) {
		if(err) throw err;
		var sgvrec;
        recs.forEach(function (rec) {
        	if (rec['sgv'] && rec['filtered']) {
//       		console.log('SGV:' + rec['sgv'] + ',' + rec['raw'] + ',' + rec['filtered'] + ',' + rec['device'] + ',' + rec['date']);
        		currentsgv = rec;
//        		sgvrec['foo'] = 50;
//				sgvrec = '';
//        		console.log(sgvrec);
        	}
        	if (rec['slp']) {
//        		console.log('CAL:' + rec['slp'] + ',' + rec['int']  + ',' + rec['scl'] + ',' + rec['date']);
        		calrecord = 1;
        	}
        	if (rec['mbg']) {
//        		console.log('MBG:' + rec['mbg'] + ',' + rec['device'] + ',' + rec['date']);
        		
        		currentmbg = rec;
        		if (mbgrecord == 0) {
        			mbgrecord = 1;
        			insertdate = currentmbg['date']+1000;
        		}
        		if ( (currentsgv['date'] - currentmbg['date']) <= (600*1000)) {
      		console.log(currentmbg['mbg'] + ' ' + currentsgv['filtered'] + ' ' + currentmbg['date']);
        		knownmbg.push(parseInt(currentmbg['mbg']));
        		knownfiltered.push(parseInt(currentsgv['filtered']));
        		}
        	}
        });

        // Only close the connection when your app is terminating.
        db.close(function (err) {
            if(err) throw err;
        });
        
        // create cal record
//		console.log(knownfiltered);
//		console.log(knownmbg);
		var lr = linearRegression(knownfiltered, knownmbg);
		var scl = (1-((1-(knownmbg[1]/((knownfiltered[1]-lr.intercept)/lr.slope)))/2));
		var normalslope = (lr.slope/scl);
		
		console.log('./pushderivedcal.sh createcal.js ' +
//					sprintf("%-0.0f", lr.slope) + ' ' +
					sprintf("%-0.0f", normalslope) + ' ' +
					sprintf("%-0.0f",lr.intercept) + ' ' +
					'1  `date +%s`000'
//					sprintf("%-0.3f",scl) + ' ' + insertdate
		);

    });
           
});

function linearRegression(y,x){
		var lr = {};
		var n = y.length;
		var sum_x = 0;
		var sum_y = 0;
		var sum_xy = 0;
		var sum_xx = 0;
		var sum_yy = 0;
		
		for (var i = 0; i < y.length; i++) {
			
			sum_x += x[i];
			sum_y += y[i];
			sum_xy += (x[i]*y[i]);
			sum_xx += (x[i]*x[i]);
			sum_yy += (y[i]*y[i]);
		} 
		
		lr['slope'] = (n * sum_xy - sum_x * sum_y) / (n*sum_xx - sum_x * sum_x);
		lr['intercept'] = (sum_y - lr.slope * sum_x)/n;
		lr['r2'] = Math.pow((n*sum_xy - sum_x*sum_y)/Math.sqrt((n*sum_xx-sum_x*sum_x)*(n*sum_yy-sum_y*sum_y)),2);
		
		return lr;
}
