package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/brutella/can"
)

var scanFrame can.Frame

// logCANFrame logs a frame with the same format as candump from can-utils.
func logCANFrame(frm can.Frame) {
	data := trimSuffix(frm.Data[:], 0x00)
	length := fmt.Sprintf("[%x]", frm.Length)

	chars := fmt.Sprintf("'%s'", printableString(data[:]))
	rcvr := ReceiverId(frm.Data[:2])
	formatted := fmt.Sprintf("%-3s %-4x %-3s % -24X %-10s %6x ", *i, frm.ID, length, data, chars, rcvr)

	reg, payload := Payload(data)
	formatted += fmt.Sprintf("%04X ", reg)

	if data[0]&Data != 0 {
		if r := Reading(reg); r != nil {
			val := DecodeValue(payload, r.Type)
			valStr := payloadString(val)

			formatted += fmt.Sprintf("%-20s %8s", left(r.Name, 20), valStr)
		}
	}

	log.Println(formatted)
}

// logCANFrame logs a frame with the same format as candump from can-utils.
func logCANFrame2(frm can.Frame) {
	rcvr := ReceiverId(frm.Data[:2])
	scanRcvr := ReceiverId(scanFrame.Data[:2])

	var match bool
	if frm.ID == scanFrame.ID && rcvr == scanRcvr {
		match = true
	} else if uint16(frm.ID) == scanRcvr && rcvr == uint16(scanFrame.ID) {
		match = true
	}

	if !match {
		return
	}

	data := trimSuffix(frm.Data[:], 0x00)
	chars := fmt.Sprintf("'%s'", printableString(data[:]))
	length := fmt.Sprintf("[%x]", frm.Length)
	formatted := fmt.Sprintf("%-3s %-4x %-3s % -24X %-10s %6x ", *i, frm.ID, length, data, chars, rcvr)

	reg, payload := Payload(data)
	formatted += fmt.Sprintf("%04X ", reg)

	if data[0]&Data != 0 {
		if r := Reading(reg); r != nil {
			if val := DecodeValue(payload, r.Type); val != nil {
				valStr := payloadString(val)

				formatted += fmt.Sprintf("%-20s %8s", left(r.Name, 20), valStr)
				log.Println(formatted)
			}
			return
		}
	}
}

func payloadString(val interface{}) string {
	if _, ok := val.(float64); ok {
		return fmt.Sprintf("%6.1f", val)
	}

	return fmt.Sprintf("%v", val)
}

func left(s string, chars int) string {
	l := len(s)
	if chars < l {
		l = chars
	}
	return s[:l]
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

func loopElster(bus *can.Bus) {
	for _, r := range ElsterReadings {
		scanFrame = can.Frame{
			ID:     0x0680,
			Length: 8,
			Flags:  0,
			Res0:   0,
			Res1:   0,
			Data:   [8]uint8{},
		}
		copy(scanFrame.Data[:], RequestFrame(0x180, r))
		bus.Publish(scanFrame)
	}
}

var i = flag.String("if", "", "network interface name")

func main() {
	flag.Parse()
	if len(*i) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	bus, err := can.NewBusForInterfaceWithName(*i)
	if err != nil {
		log.Fatal(err)
	}

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

	go bus.ConnectAndPublish()

	// bus.SubscribeFunc(logCANFrame)
	bus.SubscribeFunc(logCANFrame2)
	loopElster(bus)

	for {
		time.Sleep(40 * time.Millisecond)
	}
}
