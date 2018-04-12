package main

import (
	"flag"
	"log"
	"time"
	"math/bits"

	"github.com/ecc1/cc2500"
)

const (
	baseFrequency = 2425000000

	readingInterval = 5 * time.Minute
	channelInterval = 500 * time.Millisecond

	wakeupMargin = 100 * time.Millisecond

	slowWait = readingInterval + 1*time.Minute
	fastWait = channelInterval + 50*time.Millisecond
	syncWait = wakeupMargin + 100*time.Millisecond
	
	packetLength = 18
)


var (
	transid          = flag.String("txid", "67LDE", "transmitter ID")
	encodedraw       = []byte{0x00, 0x00}
	encodedfilt      = []byte{0x00, 0x00}
	encodedID        = []byte{0x00, 0x00, 0x00, 0x00}
	battery          = flag.Int("bat", 215, "battery voltage")
	sequence         = flag.Int("seq", 32, "sequence 0 ... 63")
	raw              = flag.Int("raw", 144192, "raw value")
	filtered         = flag.Int("filtered", 149760, "filtered value")
	count            = flag.Int("n", 0, "send only `count` packets")
	minPacketSize    = flag.Int("min", 1, "minimum packet `size` in bytes")
	maxPacketSize    = flag.Int("max", 30, "maximum packet `size` in bytes")
	frequency        = flag.Uint("f", 2425000000, "frequency in Hz")
	frequencyone     = flag.Uint("fone", 2450000000, "frequency in Hz")
	frequencytwo     = flag.Uint("ftwo", 2474750000, "frequency in Hz")
	frequencythree   = flag.Uint("fthree", 2477250000, "frequency in Hz")
	interPacketDelay = flag.Duration("delay", time.Second, "inter-packet delay")
)

// Lookup table for CRC-8 calculation with polyomial 0x2F.
var crc8Table = []uint8{
	0x00, 0x2F, 0x5E, 0x71, 0xBC, 0x93, 0xE2, 0xCD,
	0x57, 0x78, 0x09, 0x26, 0xEB, 0xC4, 0xB5, 0x9A,
	0xAE, 0x81, 0xF0, 0xDF, 0x12, 0x3D, 0x4C, 0x63,
	0xF9, 0xD6, 0xA7, 0x88, 0x45, 0x6A, 0x1B, 0x34,
	0x73, 0x5C, 0x2D, 0x02, 0xCF, 0xE0, 0x91, 0xBE,
	0x24, 0x0B, 0x7A, 0x55, 0x98, 0xB7, 0xC6, 0xE9,
	0xDD, 0xF2, 0x83, 0xAC, 0x61, 0x4E, 0x3F, 0x10,
	0x8A, 0xA5, 0xD4, 0xFB, 0x36, 0x19, 0x68, 0x47,
	0xE6, 0xC9, 0xB8, 0x97, 0x5A, 0x75, 0x04, 0x2B,
	0xB1, 0x9E, 0xEF, 0xC0, 0x0D, 0x22, 0x53, 0x7C,
	0x48, 0x67, 0x16, 0x39, 0xF4, 0xDB, 0xAA, 0x85,
	0x1F, 0x30, 0x41, 0x6E, 0xA3, 0x8C, 0xFD, 0xD2,
	0x95, 0xBA, 0xCB, 0xE4, 0x29, 0x06, 0x77, 0x58,
	0xC2, 0xED, 0x9C, 0xB3, 0x7E, 0x51, 0x20, 0x0F,
	0x3B, 0x14, 0x65, 0x4A, 0x87, 0xA8, 0xD9, 0xF6,
	0x6C, 0x43, 0x32, 0x1D, 0xD0, 0xFF, 0x8E, 0xA1,
	0xE3, 0xCC, 0xBD, 0x92, 0x5F, 0x70, 0x01, 0x2E,
	0xB4, 0x9B, 0xEA, 0xC5, 0x08, 0x27, 0x56, 0x79,
	0x4D, 0x62, 0x13, 0x3C, 0xF1, 0xDE, 0xAF, 0x80,
	0x1A, 0x35, 0x44, 0x6B, 0xA6, 0x89, 0xF8, 0xD7,
	0x90, 0xBF, 0xCE, 0xE1, 0x2C, 0x03, 0x72, 0x5D,
	0xC7, 0xE8, 0x99, 0xB6, 0x7B, 0x54, 0x25, 0x0A,
	0x3E, 0x11, 0x60, 0x4F, 0x82, 0xAD, 0xDC, 0xF3,
	0x69, 0x46, 0x37, 0x18, 0xD5, 0xFA, 0x8B, 0xA4,
	0x05, 0x2A, 0x5B, 0x74, 0xB9, 0x96, 0xE7, 0xC8,
	0x52, 0x7D, 0x0C, 0x23, 0xEE, 0xC1, 0xB0, 0x9F,
	0xAB, 0x84, 0xF5, 0xDA, 0x17, 0x38, 0x49, 0x66,
	0xFC, 0xD3, 0xA2, 0x8D, 0x40, 0x6F, 0x1E, 0x31,
	0x76, 0x59, 0x28, 0x07, 0xCA, 0xE5, 0x94, 0xBB,
	0x21, 0x0E, 0x7F, 0x50, 0x9D, 0xB2, 0xC3, 0xEC,
	0xD8, 0xF7, 0x86, 0xA9, 0x64, 0x4B, 0x3A, 0x15,
	0x8F, 0xA0, 0xD1, 0xFE, 0x33, 0x1C, 0x6D, 0x42,
}

// CRC8 computes the 8-bit CRC of the given data.
func CRC8(msg []byte) byte {
	res := byte(0)
	for _, b := range msg {
		res = crc8Table[res^b]
	}
	return res
}

var transmitterIDChar = []byte{
	'0', '1', '2', '3', '4', '5', '6', '7',
	'8', '9', 'A', 'B', 'C', 'D', 'E', 'F',
	'G', 'H', 'J', 'K', 'L', 'M', 'N', 'P',
	'Q', 'R', 'S', 'T', 'U', 'W', 'X', 'Y',
}

func getsrcvalue(v byte) uint32 {
	var i uint32
	for i = 0; i < 32; i++ {
		if transmitterIDChar[i]==v { break }
	}
	log.Printf("idsrc %c == trasmitterID location %d", v, i)
	return i
	}

func marshalTransmitterID(v string) []byte {
	var i uint8
	var transmitterID uint32
	var arr []byte
	//var srcval byte
	var srcinx uint32
	arr = []byte(v)
	for i = 0; i < 5; i++ {
		log.Printf("char %c", arr[i])
		srcinx = getsrcvalue(arr[i])
		transmitterID = transmitterID | srcinx<<((4-i)*5)
		log.Printf("char srcinx %d", srcinx)
	}
	//transmitterID = bits.Reverse32(transmitterID)
	//transmitterID = transmitterID<<7 
	u0, u1, u2, u3 := uint8(transmitterID), uint8(transmitterID>>8), uint8(transmitterID>>16), uint8(transmitterID>>24)
	//u0, u1, u2, u3 := bits.Reverse8(uint8(transmitterID)), bits.Reverse8(uint8(transmitterID>>8)), bits.Reverse8(uint8(transmitterID>>16)), bits.Reverse8(uint8(transmitterID>>24))
	log.Printf("TransmitterID %02X %02X %02X %02X", u0, u1, u2, u3)
	return []byte{u0, u1, u2, u3}
}

//func unmarshalTransmitterID(v []byte) string {
//	u := unmarshalUint32(v)
//	id := make([]byte, 5)
//	for i := 0; i < 5; i++ {
//		n := byte(u>>uint(20-5*i)) & 0x1F
//		id[i] = transmitterIDChar[n]
//	}
//	return string(id)
//}

// Unmarshal a little-endian uint16.
//func unmarshalUint16(v []byte) uint16 {
//	return uint16(v[0]) | uint16(v[1])<<8
//}

// Unmarshal a little-endian uint32.
//func unmarshalUint32(v []byte) uint32 {
//	return uint32(unmarshalUint16(v[0:2])) | uint32(unmarshalUint16(v[2:4]))<<16
//}

//uint32 getSrcValue(char srcVal)
//{
//	uint8 i = 0;
//	for(i = 0; i < 32; i++)
//	{
//			if (SrcNameTable[i]==srcVal) break;
//	}
//	//printf("getSrcVal: %c %u\r\n",srcVal, i);
//	return i & 0xFF;
//}

//uint32 asciiToDexcomSrc(char addr[6])
//{
//	// prepare a uint32 variable for our return value
//	uint32 src = 0;
//	// look up the first character, and shift it 20 bits left.
//	src |= (getSrcValue(addr[0]) << 20);
//	// look up the second character, and shift it 20 bits left.
//	src |= (getSrcValue(addr[1]) << 15);
//	// look up the third character, and shift it 20 bits left.
//	src |= (getSrcValue(addr[2]) << 10);
//	// look up the fourth character, and shift it 20 bits left.
//	src |= (getSrcValue(addr[3]) << 5);
//	// look up the fifth character, and shift it 20 bits left.
//	src |= getSrcValue(addr[4]);
//	//printf("asciiToDexcomSrc: val=%u, src=%u\r\n", val, src);
//	return src;
//}


// Marshal a uint32 as a 16-bit float (13-bit mantissa, 3-bit exponent) byte reversed
//func marshalReading(v int) uint16 {
func marshalReading(v int) []byte {
var usExponent int
	for x := 0; x < 7; x++ {
		loopMantissa := v >> uint(x)
		if loopMantissa < 8192 { break }
		usExponent++
	}
usReversed := uint16(uint(usExponent) << 13 | uint(v) >> uint(usExponent))
u0, u1 := bits.Reverse8(uint8(usReversed)), bits.Reverse8(uint8(usReversed>>8))
//log.Printf("u0: %02X", u0)
//log.Printf("u1: %02X", u1)
return []byte{u0, u1}
}

func main() {
	log.SetFlags(log.Ltime | log.Lmicroseconds | log.LUTC)
	flag.Parse()
	r := cc2500.Open()
	if r.Error() != nil {
		log.Fatal(r.Error())
	}
	defer r.Close()

	n := 18
	pkts := 0
	data := make([]byte, *maxPacketSize)
	for r.Error() == nil {
		if *count != 0 && pkts == *count {
			return
		}
		
		//	0..3: destination address (always FF FF FF FF = broadcast)
		data[0] = byte(0xff) 
		data[1] = byte(0xff)
		data[2] = byte(0xff)
		data[3] = byte(0xff)
		
		encodedID = marshalTransmitterID(*transid)
		
		//	4..7: transmitter ID
		data[4] = byte(encodedID[0])
		data[5] = byte(encodedID[1])
		data[6] = byte(encodedID[2])
		data[7] = byte(encodedID[3])
		
		//	8: port? (always 3F)
		data[8] = byte(0x3f)
		
		//	9: device info? (always 03)
		data[9] = byte(0x03)
		
		//	10: sequence number
		data[10] = byte(*sequence)<<2
		
		//	11..12: raw reading
		encodedraw = marshalReading(*raw)
		data[11] = byte(encodedraw[0])
		data[12] = byte(encodedraw[1])
		
		//	13..14: filtered reading
		encodedfilt = marshalReading(*filtered/2)
		data[13] = byte(encodedfilt[0])
		data[14] = byte(encodedfilt[1])
		
		//	15: battery level
		//*battery = 100
		data[15] = byte(*battery)
		//	16: unknown
		data[16] = byte(0x00)
		//	17: checksum
		calcCRC := CRC8(data[11 : packetLength-1])
		
		//marshalfilt := marshalReading(*filtered/2)
		//log.Printf("marshalfilt: %02X", marshalfilt)
		
		//log.Printf("calcCRC: %02X", calcCRC)
		data[17] = byte(calcCRC)
		packet := data[:n]
		
		// send channel 0
		log.Printf("setting frequency to %d", *frequency)
	    r.Init(uint32(*frequency))
	    log.Printf("actual frequency: %d", r.Frequency())
		log.Printf("data channel 0: % X", packet)
		r.Send(packet)
		
		//send channel 1
		//log.Printf("setting frequency to %d", *frequencyone)
		//r.Init(uint32(*frequencyone))
	    //log.Printf("actual frequency: %d", r.Frequency())
	    //time.Sleep(fastWait)	    
	    //log.Printf("data channel 1: % X", packet)
	    //r.Send(packet)

	    //send channel 2
	    //log.Printf("setting frequency to %d", *frequencytwo)
	    //r.Init(uint32(*frequencytwo))
	    //log.Printf("actual frequency: %d", r.Frequency())
	    //time.Sleep(fastWait)	    
	    //log.Printf("data channel 2: % X", packet)
	    //r.Send(packet)

	    //send channel 3
	    //log.Printf("setting frequency to %d", *frequencythree)
	    //r.Init(uint32(*frequencythree))
	    //log.Printf("actual frequency: %d", r.Frequency())
	    //time.Sleep(fastWait)	    
	    //log.Printf("data  channel 3: % X", packet)
	    //r.Send(packet)

		pkts++
		if n > *maxPacketSize {
			n = *minPacketSize
		}
		time.Sleep(*interPacketDelay)
	}
	log.Fatal(r.Error())
}
