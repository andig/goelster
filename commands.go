package goelster

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/brutella/can"
)

func CanDump(bus *can.Bus) {
	bus.SubscribeFunc(LogFrame)
	bus.ConnectAndPublish()
}

// makeScanMatcher creates a function that matches send/recv/reg for incoming frames
func makeScanMatcher(
	c chan can.Frame,
	sender uint16,
	receiver uint16,
	register uint16,
) func(frm can.Frame) {
	return func(frm can.Frame) {
		frmReceiver := ReceiverId(frm.Data[:2])

		// frame sent back from receiver to sender?
		if uint16(frm.ID) == receiver && frmReceiver == sender {
			// requested register?
			if reg, _ := Payload(frm.Data[:]); reg == register {
				// data frame?
				if frm.Data[0]&Data != 0 {
					c <- frm
				}
			}
		}
	}
}

// createReadFrame constructs a CAN bus request frame
func createReadFrame(sender uint16, receiver uint16, r ElsterReading) *can.Frame {
	frm := can.Frame{
		ID:     uint32(sender),
		Length: 8,
		Data:   [8]uint8{},
	}
	copy(frm.Data[:], RequestFrame(receiver, r))
	return &frm
}

func readRegister(
	bus *can.Bus,
	sender uint16,
	receiver uint16,
	r ElsterReading,
) *can.Frame {
	c := make(chan can.Frame) // signalling channel
	frm := createReadFrame(sender, receiver, r)

	handler := can.NewHandler(makeScanMatcher(c, sender, receiver, r.Index))
	bus.Subscribe(handler)
	defer bus.Unsubscribe(handler)

	bus.Publish(*frm)
	select {
	case <-time.After(100 * time.Millisecond):
		// timeout
		return nil
	case frm := <-c:
		// result
		return &frm
	}
}

func CanScan(bus *can.Bus, sender uint16, receiver uint16) {
	go bus.ConnectAndPublish()
	defer bus.Disconnect()

	for _, r := range ElsterReadings {
		if frm := readRegister(bus, sender, receiver, r); frm != nil {
			_, payload := Payload(frm.Data[:])
			val := DecodeValue(payload, r.Type)

			if RawLog || val != nil {
				LogFrame(*frm)
			} else {
				LogRegisterValue(val, r)
			}
		}
	}
}

func CanRead(bus *can.Bus, sender uint16, receiver uint16, register uint16) {
	r := Reading(register)
	if r == nil {
		log.Fatalf("Unknown register '%X'", register)
	}

	go bus.ConnectAndPublish()
	defer bus.Disconnect()

	frm := readRegister(bus, sender, receiver, *r)
	if frm == nil {
		os.Exit(1)
	}

	if RawLog {
		LogFrame(*frm)
	} else {
		_, payload := Payload(frm.Data[:])
		val := DecodeValue(payload, r.Type)
		valStr := ValueString(val)
		fmt.Println(strings.Trim(valStr, " "))
	}
}

func CanWrite(bus *can.Bus, sender uint16, receiver uint16, register uint16, value uint16) {
	log.Fatal("Not implemented")
}
