package goelster

import (
	"bytes"
	"testing"
)

func TestDecodeReceiverId(t *testing.T) {
	rcvr := ReceiverId([]byte{0xD2, 0x0F})
	if rcvr != 0x68f {
		t.Errorf("Receiver id incorrect, got: %X, want: %X.", rcvr, 0x68f)
	}
	rcvr = ReceiverId([]byte{0xA1, 0x00})
	if rcvr != 0x500 {
		t.Errorf("Receiver id incorrect, got: %X, want: %X.", rcvr, 0x500)
	}
}

func TestDecodePayload(t *testing.T) {
	b := []byte{0x00, 0x00, 0x0C, 0x01, 0x02}
	reg, payload := Payload(b)
	if reg != 0x0C {
		t.Errorf("Register incorrect, got: %X, want: %X.", reg, 0x0C)
	}
	if !bytes.Equal(payload, []byte{0x01, 0x02}) {
		t.Errorf("Register incorrect, got: %X, want: %X.", payload, []byte{0x01, 0x02})
	}

	b = []byte{0x00, 0x00, 0xFA, 0x03, 0x04, 0x01, 0x02}
	reg, payload = Payload(b)
	if reg != 0x0304 {
		t.Errorf("Register incorrect, got: %X, want: %X.", reg, 0x0304)
	}
	if !bytes.Equal(payload, []byte{0x01, 0x02}) {
		t.Errorf("Register incorrect, got: %X, want: %X.", payload, []byte{0x01, 0x02})
	}
}

func TestEncodeFrame(t *testing.T) {
	r := *Reading(0x0002) // decimal value
	val := 32.1           // 312 -> 0x0141
	rcvr := uint16(0x500)
	frame := RequestFrame(rcvr, r)

	expected := []byte{0xA1, 0x00, 0x02, 0x00, 0x00, 0x00, 0x00, 0x00}
	if !bytes.Equal(frame, expected) {
		t.Errorf("Frame incorrect, got: % X, want: % X.", frame, expected)
	}

	r = *Reading(0x0002) // decimal value
	val = 32.1           // 312 -> 0x0141
	rcvr = uint16(0x500)
	frame = DataFrame(rcvr, val, r)

	expected = []byte{0xA2, 0x00, 0x02, 0x01, 0x41, 0x00, 0x00, 0x00}
	if !bytes.Equal(frame, expected) {
		t.Errorf("Frame incorrect, got: % X, want: % X.", frame, expected)
	}

	r = *Reading(0x010c) // decimal value
	val = 32.1           // 321 -> 0x0141
	rcvr = uint16(0x68f)
	frame = DataFrame(rcvr, val, r)

	expected = []byte{0xD2, 0x0F, 0xFA, 0x01, 0x0C, 0x01, 0x41, 0x00}
	if !bytes.Equal(frame, expected) {
		t.Errorf("Frame incorrect, got: % X, want: % X.", frame, expected)
	}
}
