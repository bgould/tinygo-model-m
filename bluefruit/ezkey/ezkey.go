package ezkey

import "github.com/tinygo-org/tinygo/src/machine"

const debug = false

type HID struct {
	bus   machine.UART
	reset machine.Pin
}

type Report [8]byte

type KeyboardModifier byte

const (
	KbdModNone       KeyboardModifier = 0x0
	KbdModCtrlLeft   KeyboardModifier = 1 << 0
	KbdModShiftLeft  KeyboardModifier = 1 << 1
	KbdModAltLeft    KeyboardModifier = 1 << 2
	KbdModGuiLeft    KeyboardModifier = 1 << 3
	KbdModCtrlRight  KeyboardModifier = 1 << 4
	KbdModShiftRight KeyboardModifier = 1 << 5
	KbdModAltRight   KeyboardModifier = 1 << 6
	KbdModGuiRight   KeyboardModifier = 1 << 7
)

type MouseButton byte

const (
	MouseBtnLeft   MouseButton = 0x01
	MouseBtnRight  MouseButton = 0x02
	MouseBtnMiddle MouseButton = 0x04
)

type ConsumerKey uint16

const (
	ConsKeyHome       ConsumerKey = 0x0100
	ConsKeyKbdLayout  ConsumerKey = 0x0200
	ConsKeySearch     ConsumerKey = 0x0400
	ConsKeySnapshot   ConsumerKey = 0x0800
	ConsKeyVolUp      ConsumerKey = 0x1000
	ConsKeyVolDown    ConsumerKey = 0x2000
	ConsKeyPlayPause  ConsumerKey = 0x4000
	ConsKeyFastFwd    ConsumerKey = 0x8000
	ConsKeyRewind     ConsumerKey = 0x0001
	ConsKeyNextTrack  ConsumerKey = 0x0002
	ConsKeyPrevTrack  ConsumerKey = 0x0004
	ConsKeyRandomPlay ConsumerKey = 0x0008
	ConsKeyStop       ConsumerKey = 0x0010
)

func (r *Report) Keyboard(mod KeyboardModifier, keys ...byte) *Report {
	r[0] = byte(mod)
	r[1] = 0x0
	for i, c := 0, len(keys); i < 6; i++ {
		if i < c {
			r[i+2] = keys[i]
		} else {
			r[i+2] = 0x0
		}
	}
	return r
}

func (r *Report) Mouse(buttons MouseButton, x int8, y int8) *Report {
	r[0] = 0x0
	r[1] = 0x3
	r[2] = byte(buttons)
	r[3] = byte(x)
	r[4] = byte(y)
	r[5] = 0x0
	r[6] = 0x0
	r[7] = 0x0
	return r
}

func (r *Report) Consumer(key ConsumerKey) *Report {
	r[0] = 0x0
	r[1] = 0x2
	r[2] = byte(key >> 8)
	r[3] = byte(key & 0xFF)
	r[4] = 0x0
	r[5] = 0x0
	r[6] = 0x0
	r[7] = 0x0
	return r
}

func New(uart machine.UART, resetPin machine.Pin) *HID {
	return &HID{
		bus:   uart,
		reset: resetPin,
	}
}

func (hid *HID) Send(rpt *Report) {
	if debug {
		println(rpt[0], rpt[1], rpt[2], rpt[3], rpt[4], rpt[5], rpt[6], rpt[7])
	}
	hid.bus.WriteByte(0xFD)
	hid.bus.Write(rpt[:])
}
