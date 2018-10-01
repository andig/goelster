package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/brutella/can"
)

// logCANFrame logs a frame with the same format as candump from can-utils.
func logCANFrame(frm can.Frame) {
	data := trimSuffix(frm.Data[:], 0x00)
	length := fmt.Sprintf("[%x]", frm.Length)
	log.Printf("%-3s %-4x %-3s % -24X '%s'\n", *i, frm.ID, length, data, printableString(data[:]))

	logElster(frm, data)
}

// trim returns a subslice of s by slicing off all trailing b bytes.
func trimSuffix(s []byte, b byte) []byte {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] != b {
			return s[:i+1]
		}
	}

	return []byte{}
}

// printableString creates a string from s and replaces non-printable bytes (i.e. 0-32, 127)
// with '.' – similar how candump from can-utils does it.
func printableString(s []byte) string {
	var ascii []byte
	for _, b := range s {
		if b < 32 || b > 126 {
			b = byte('.')

		}
		ascii = append(ascii, b)
	}

	return string(ascii)
}

type CanSendReceive int

const (
	send CanSendReceive = iota
	receive
)

func bytes2id(b []byte) uint16 {
	return uint16(b[0]&0xF0)<<8 + uint16(b[1]&0x0F)
}

func id2bytes(id uint16, sr CanSendReceive) []byte {
	b := make([]byte, 2)

	b[0] = byte(id>>8) & 0xF0
	b[1] = byte(id) & 0xFF

	sendreceive := byte(1)
	if sr == receive {
		sendreceive = 2
	}

	b[0] = b[0] | sendreceive

	return b
}

/*
	00 00 06 80  05 00 00 00  31 00  fa  09 31
	00 00 01 80  07 00 00 00  d2 00  fa  09 31  00 27
	00 00 06 80  05 00 00 00  31 00  fa  09 30
	00 00 01 80  07 00 00 00  d2 00  fa  09 30  00 73
	|---------|  ||           |---|  ||  |---|  |---|
	1)           2)           3)     4)  5)     6)

	1) Sender CAN ID: 180 or 680
	2) No of bytes of data - 5 for queries, 7 for replies
	3) Receiver CAN ID of the communications partner and type of message
		For queries the second digit is 1.
		Pattern: n1 0m with n = 180 / 80 = 3 (hex) and m = 180 mod 8 = 0
		(hex) Partner ID = 30 * 8 (hex) + 00 = 180
		Responses follow a similar pattern using second digit 2:
		Partner ID is: d0 * 8 + 00 = 680
	4) 0xFA indicates that the Elster index is greater than ff.
	5) Index (parameter) queried for: 0930 for kWh and 0931 for MWh
	6) Value returned 27h=39,73h=115
*/

func logElster(frm can.Frame, data []byte) {
	id := bytes2id(data[:2])

	var reg uint16
	payload := make([]byte, 2)

	if data[2] == 0xFA {
		reg = binary.BigEndian.Uint16(data[3:5])
		if copy(payload, data[5:7]) != 2 {
			log.Panic("Invalid copy length")
		}
	} else {
		reg = uint16(data[2])
		if copy(payload, data[3:5]) != 2 {
			log.Panic("Invalid copy length")
		}
	}

	log.Printf("id %x\n, reg %x, payload % x", id, reg, payload)
}

// func loopElster() {
// 	for _, e := range ElsterReadings {
// 		frm := can.Frame{
// 			ID:     0x0180,
// 			Length: 0,
// 			Flags:  0,
// 			Res0:   0,
// 			Res1:   0,
// 			Data:   [8]uint8{0, 0, 0, 0, 0, 0, 0, 0},
// 		}
// 	}
// }

var i = flag.String("if", "", "network interface name")

func main() {
	flag.Parse()
	if len(*i) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	// iface, err := net.InterfaceByName(*i)
	// if err != nil {
	// 	log.Fatalf("Could not find network interface %s (%v)", *i, err)
	// }

	// conn, err := can.NewReadWriteCloserForInterface(iface)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// bus := can.NewBus(conn)

	bus, err := can.NewBusForInterfaceWithName(*i)
	if err != nil {
		log.Fatal(err)
	}

	bus.SubscribeFunc(logCANFrame)

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, os.Kill)

	go func() {
		select {
		case <-c:
			bus.Disconnect()
			os.Exit(1)
		}
	}()

	bus.ConnectAndPublish()

	// loopElster()
}
