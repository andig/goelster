package goelster

import (
	"log"
	"time"

	"github.com/brutella/can"
)

func CanDump(bus *can.Bus) {
	bus.SubscribeFunc(logFrame)
	bus.ConnectAndPublish()
}

// makeScanMatcher creates a function that matches send/recv/reg for incoming frames
func makeScanMatcher(c chan can.Frame, sender uint16, receiver uint16, register uint16) func(frm can.Frame) {
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

func createReadFrame(sender uint16, receiver uint16, r ElsterReading) *can.Frame {
	frm := can.Frame{
		ID:     uint32(sender),
		Length: 8,
		Data:   [8]uint8{},
	}
	copy(frm.Data[:], RequestFrame(receiver, r))
	return &frm
}

func readRegister(bus *can.Bus, sender uint16, receiver uint16, r ElsterReading) {
	// signalling channel
	c := make(chan can.Frame)
	frm := createReadFrame(sender, receiver, r)

	handler := can.NewHandler(makeScanMatcher(c, sender, receiver, r.Index))
	bus.Subscribe(handler)
	bus.Publish(*frm)

	select {
	case <-time.After(100 * time.Millisecond):
		// timeout
	case frm := <-c:
		// result
		logFrame(frm)
	}

	bus.Unsubscribe(handler)
}

func CanScan(bus *can.Bus, sender uint16, receiver uint16) {
	go bus.ConnectAndPublish()

	for _, r := range ElsterReadings {
		readRegister(bus, sender, receiver, r)
	}

	bus.Disconnect()
}

func CanRead(bus *can.Bus, sender uint16, receiver uint16, register uint16) {
	r := Reading(register)
	if r == nil {
		log.Fatalf("Unknown register '%X'", register)
	}

	go bus.ConnectAndPublish()
	readRegister(bus, sender, receiver, *r)
	bus.Disconnect()
}

func CanWrite(bus *can.Bus, sender uint16, receiver uint16, register uint16, value uint16) {

}
