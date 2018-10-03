package goelster

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
)

/*
   Elster frame decoding
   ---------------------

   06 80  05  31 00  fa  09 31
   01 80  07  d2 00  fa  09 31  00 27
   06 80  05  31 00  fa  09 30
   01 80  07  d2 00  fa  09 30  00 73
   |---|  ||  |---|  ||  |---|  |---|
   1)     2)  3)     4)  5)     6)

   1) Sender CAN ID: 180 or 680
   2) No of bytes of data - 5 for queries, 7 for replies
   3) Receiver CAN ID of the communications partner and type of message
       Queries:   2nd digit is 1
	   Pattern:   n1 0m with n = 0x30, m = 0x00
                  Partner ID: 0x30 * 8 + 0x00 = 0x180
       Responses: 2nd digit is 2
                  Partner ID: 0xd0 * 8 + 0x00 = 0x680
   4) 0xFA indicates that the Elster index is greater than ff.
   5) Index (parameter) queried for: 0930 for kWh and 0931 for MWh
   6) Value returned 27h=39,73h=115
*/

const (
	// byte 0
	Request byte = 0x01
	Data    byte = 0x02
	// byte 1
	Broadcast byte = 0x79
)

func ReceiverId(b []byte) uint16 {
	return uint16(b[0]&0xF0)<<3 + uint16(b[1]&0x0F)
}

func Payload(data []byte) (reg uint16, payload []byte) {
	if data[2] == 0xFA {
		reg = binary.BigEndian.Uint16(data[3:5])
		payload = data[5:7]
	} else {
		reg = uint16(data[2])
		payload = data[3:5]
	}
	return reg, payload
}

func Reading(register uint16) *ElsterReading {
	for _, r := range ElsterReadings {
		if r.Index == register {
			return &r
		}
	}
	return nil
}

func DecodeValue(b []byte, t ElsterType) interface{} {
	if bytes.Equal(b, []byte{0x80, 0x00}) {
		return nil
	}

	switch t {
	case et_little_endian:
		return binary.LittleEndian.Uint16(b)
	case et_dec_val:
		return float64(binary.BigEndian.Uint16(b)) / 10
	case et_cent_val:
		return float64(binary.BigEndian.Uint16(b)) / 100
	case et_mil_val:
		return float64(binary.BigEndian.Uint16(b)) / 1000

	case et_byte:
		return b[0]

	case et_zeit:
		val := binary.BigEndian.Uint16(b)
		return fmt.Sprintf("%2.2d:%2.2d", val&0xff, val>>8)
	case et_datum:
		val := binary.BigEndian.Uint16(b)
		return fmt.Sprintf("%2.2d.%2.2d.", val>>8, val&0xff)
	case et_time_domain:
		val := binary.BigEndian.Uint16(b)
		if val&0x8080 != 0 {
			return nil
		}
		return fmt.Sprintf("%2.2d:%2.2d-%2.2d:%2.2d",
			(val>>8)/4, 15*((val>>8)%4),
			(val&0xff)/4, 15*(val%4))

	case et_little_bool:
		if bytes.Equal(b, []byte{0x01, 0x00}) {
			return true
		} else if bytes.Equal(b, []byte{0x00, 0x00}) {
			return false
		}
		log.Fatalf("Invalid little bool % x", b)
		return nil

	case et_bool:
		if bytes.Equal(b, []byte{0x00, 0x01}) {
			return true
		} else if bytes.Equal(b, []byte{0x00, 0x00}) {
			return false
		}
		log.Fatalf("Invalid bool % x", b)
		return nil
	}

	// default
	return b
}

func EncodeValue(val interface{}, t ElsterType) []byte {
	b := []byte{0, 0}

	switch t {
	case et_little_endian:
		u := uint16(val.(float64))
		binary.LittleEndian.PutUint16(b, u)
	case et_dec_val:
		u := uint16(val.(float64) * 10)
		binary.BigEndian.PutUint16(b, u)
	case et_cent_val:
		u := uint16(val.(float64) * 100)
		binary.BigEndian.PutUint16(b, u)
	case et_mil_val:
		u := uint16(val.(float64) * 1000)
		binary.BigEndian.PutUint16(b, u)

	case et_byte:
		b[0] = val.(byte)
	case et_little_bool:
		if val.(bool) {
			b[0] = 0x01
		}
	case et_bool:
		if val.(bool) {
			b[1] = 0x01
		}
	}

	// default
	return b
}

func EncodeRegister(b []byte, register uint16) int {
	if register > 0xFF {
		b[2] = 0xFA
		binary.BigEndian.PutUint16(b[3:], register)
		return 5
	}

	b[2] = byte(register)
	return 3
}

func EncodeReceiver(b []byte, receiverId uint16, requestType byte) {
	b[0] = byte(receiverId>>3)&0xF0 | requestType
	b[1] = byte(receiverId) & 0x0F
}

func RequestFrame(receiverId uint16, reading ElsterReading) []byte {
	b := make([]byte, 8)

	EncodeReceiver(b, receiverId, Request)
	EncodeRegister(b, reading.Index)

	return b
}

func DataFrame(receiverId uint16, val interface{}, reading ElsterReading) []byte {
	b := make([]byte, 8)

	EncodeReceiver(b, receiverId, Data)
	valIdx := EncodeRegister(b, reading.Index)
	copy(b[valIdx:], EncodeValue(val, reading.Type))

	return b
}
