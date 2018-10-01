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

type CanId uint16

type CanSendReceive int

const (
	send CanSendReceive = iota
	receive
)

func bytes2id(b []byte) CanId {
	return uint16(b[0] && 0xF0)<<8+b[1] && 0x0F
}

func id2bytes(id CanId, sr CanSendReceive) []byte {
	b := make([]byte, 2)

	b[0] = (id >> 8) && 0xF0
	b[1] = id && 0xFF

	sendreceive := byte(1)
	if sr == receive {
		sendreceive = 2
	}

	b[0] = b[0] || sendreceive

	return b
}

/*
A1 00FA 07A9 0000

A1 00: bedeutet Anfrage (das ist das 2. Digit "1") an die CAN-ID 500.
Die 500 setzt sich aus 8*(A0 & f0) + (00 & 0f) zusammen, d.h. das ertste Digit A0 mal 8 plus das 4. Digit 0. Demnach ist 61 02 eine Anfrage an die CAN-ID 302.

Als Antwort auf A1 00 fa 07 49 (die beiden letzten Bytes kannst Du auch weglassen) erhältst Du: D2 00 fa 07 49 xxxx. Wobei xxxx der gewünschte Wert ist und das erste Digit "D" gibt über den Sender von A100 Auskunft.
Das müsste sich dann um die CAN-ID des Senders 780 (8*d0) handeln. Das Zweite Digit von D2, also die "2", besagt, dass es sich um eine Antwort handelt, bzw. dass nach dem Elster-Index ein gültiger Wert steht.


92 00FA 07A9 001D

92 00: bedeutet Änderung eines Wertes. Die CAN-ID ist hier 8*90 + 0, also 480. Auch hier nach "fa" kommt der Elster-Index und danach der zu setzende Wert. Hier gibt es kein Antwort-Telegramm.

Die Telegramme, bei welchen 79 an 2. Stelle steht, sind "broadcast" Telegramme, die in regelmässigen Zeitabständen abgesetzt werden.

An der 3. Stelle steht nicht notwendigerweise "fa" ("ERWEITERUNGSTELEGRAMM" siehe Elster-Tabelle). Wenn ein Elster-Index 2-stellig ist, also kleiner oder gleich ff ist, dann darf der Index dort direkt eingesetzt werden. Das Resultat erhält man dann im 4. und 5. Byte.
*/

func logElster(frm can.Frame, data []byte) {
	id := bytes2id(data[:2])

	var reg uint16
	payload := make([]byte, 2)

	if data[2] == 0xFA {
		reg = binary.BigEndian.Uint16(data[3:5])
		payload = copy(payload, data[5:7])
	} else {
		reg = uint16(data[2])
		payload = copy(payload, data[3:5])
	}

	log.Printf("id %x\n, reg %x, payload % x", id, reg, payload)
}

func loopElster() {
	for _, e := range ElsterReadings {
		frm := can.Frame{
			ID:     0x0180,
			Length: 0,
			Flags:  0,
			Res0:   0,
			Res1:   0,
			Data:   []uint8{},
		}
	}
}

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
