package main

import (
	"bytes"
	"testing"
)

func TestDecodeReceiverId(t *testing.T) {
	rcvr := ReceiverId([]byte{0xD2, 0x0F})
	if rcvr != 0x68f {
		t.Errorf("Receiver id incorrect, got: %x, want: %x.", rcvr, 0x68f)
	}
	rcvr = ReceiverId([]byte{0xA1, 0x00})
	if rcvr != 0x500 {
		t.Errorf("Receiver id incorrect, got: %x, want: %x.", rcvr, 0x500)
	}
}

func TestDecodePayload(t *testing.T) {
	b := []byte{0x00, 0x00, 0x0C, 0x01, 0x02}
	reg, payload := Payload(b)
	if reg != 0x0C {
		t.Errorf("Register incorrect, got: %x, want: %x.", reg, 0x0C)
	}
	if !bytes.Equal(payload, []byte{0x01, 0x02}) {
		t.Errorf("Register incorrect, got: %x, want: %x.", payload, []byte{0x01, 0x02})
	}

	b = []byte{0x00, 0x00, 0xFA, 0x03, 0x04, 0x01, 0x02}
	reg, payload = Payload(b)
	if reg != 0x0304 {
		t.Errorf("Register incorrect, got: %x, want: %x.", reg, 0x0304)
	}
	if !bytes.Equal(payload, []byte{0x01, 0x02}) {
		t.Errorf("Register incorrect, got: %x, want: %x.", payload, []byte{0x01, 0x02})
	}
}
