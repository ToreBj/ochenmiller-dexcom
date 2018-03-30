package main

import (
	"log"
	"time"
	"fmt"

	"github.com/ecc1/gpio"
)

//working version of uart receiver at 25 baud
//var thechar  uint16 = 0

func main() {
//	grts, errrts := gpio.Input(19, "rising", true)
//	if errrts != nil {
//		log.Fatal(errrts)
//	}
	gsigi, errsigi := gpio.Input(19, "rising", true)
	if errsigi != nil {
		log.Fatal(errsigi)
	}
	gsig, errsig := gpio.Input(19, "none", true)
	if errsig != nil {
		log.Fatal(errsig)
	}
	var thechar  uint16 = 0
	charcount := 0
	thestring := ""
//	log.Printf("starting outside loop")
	for {
		thestring = ""
		charcount = 0
//		errrts = grts.Wait(60 * time.Second)
//		if errrts != nil {
//			_, isTimeout := errrts.(gpio.TimeoutError)
//			if isTimeout {
//				log.Print(errrts)
//				continue
//			}
//			log.Fatal(errrts)
//		}
//		log.Print("pin 20 interrupt")
//		brts, errrts := grts.Read()
//			if errrts != nil {
//				log.Fatal(errrts)
//			}
//			if brts {
//			log.Printf("interrupt = one")
//			} else {
//				log.Printf("interrupt = zero")
//			}
		//now read signal
//		time.Sleep(5000 * time.Microsecond)
		startbit := 0
		instartbit := 0
		j := 0
		bitcount := 0
		bsig := true
//		log.Printf("starting inside loop")
//		for i := 0; i < 4096; i++{
        for {
			if startbit == 0 {
				errsigi := gsigi.Wait(10000 * time.Microsecond)
//                errsigi := gsigi.Wait(420 * time.Second)
				if errsigi != nil {
			    	_, isTimeout := errsigi.(gpio.TimeoutError)
			    	if isTimeout {
					//	log.Print(errsigi)
						continue
					}
					log.Fatal(errsigi)
				}
//				log.Printf("interrupt 19 fired")
				startbit = 1
				instartbit = 1
				thechar = 0
//				log.Printf("")
//				log.Printf("STB")
				time.Sleep(26000 * time.Microsecond)
			}
			bsigr, errsig := gsig.Read()
			if errsig != nil {
				log.Fatal(errsig)
			}
			bsig = bsigr
			if bitcount > 15 {
				startbit = 0
				bitcount = 0
				//log.Printf("thechar = %016b", thechar)
				thechar ^=0xFFFF
				foo := uint8(thechar & 1)
				foo |= uint8((thechar & 4)>>1)
				foo |= uint8((thechar & 16)>>2)
				foo |= uint8((thechar & 64)>>3)
				foo |= uint8((thechar & 256)>>4)
				foo |= uint8((thechar & 1024)>>5)
				foo |= uint8((thechar & 4096)>>6)
				foo |= uint8((thechar & 16384)>>7)
				//log.Printf("foo = %08b", foo)
				//log.Printf("char= %d", foo)
				//log.Printf("ithchar = %016b", thechar)
				thestring = fmt.Sprintf("%s%c", thestring, foo)
				//thestring = thestring + string(thechar)
				if foo == 10 {
					log.Printf("%s", thestring)
					thestring = ""
					break
				}
//				log.Printf("%s", thestring) //test
				thechar = 0
			}
			if bsig {
			    j = 0
				if startbit == 1 && instartbit == 0 {
					bitcount++
					//thechar |= 196
					thechar |= uint16(uint(1) << (uint(bitcount)-1))
///					log.Printf("1 - %d", bitcount)
					//log.Printf("2^bitcount = %d", uint16(uint(1) << uint(bitcount)))
				}
				if startbit == 1 && instartbit == 1 {
					instartbit = 0
				}
//				if startbit == 0 {
//					startbit = 1
//					log.Printf("")
///					log.Printf("STB")
//				}
			} else {
			    j++
			    if startbit == 1 {
			    	bitcount++
			    	//thechar = thechar & uint16(uint(1) << uint(bitcount))
///					log.Printf("0 - %d", bitcount)
					//log.Printf("2^bitcount = %d", uint16(uint(1) << uint(bitcount)))
				}
				if j > 20 { 
				log.Printf("breaking")
				break 
				}
			}
			time.Sleep(19000 * time.Microsecond)
			charcount++
		}
        //time.Sleep(1000000 * time.Microsecond)
//        log.Printf("charcount = %d", charcount)
        charcount = 0
	}
}
