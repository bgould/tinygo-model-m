package sdep

import (
	"bytes"
	"fmt"
)

const (
	MaxPacketSize  = 20
	MaxPayloadSize = 16

	CmdTypeInitialize = 0xBEEF
	CmdTypeATWrapper  = 0x0A00
	CmdTypeBLEUARTTx  = 0x0A01
	CmdTypeBLEUARTRx  = 0x0A02

	MsgTypeCommand  = 0x10
	MsgTypeResponse = 0x20
	MsgTypeAlert    = 0x40
	MsgTypeError    = 0x80

	ErrSlaveDeviceNotReady     = 0xFE
	ErrSlaveDeviceReadOverflow = 0xFF
)

type Header struct {
	Type uint8
	ID   uint16
	Size uint8
}

func (header *Header) HasMoreData() bool {
	return header.Size>>7 > 0
}

func (header *Header) GetLength() uint8 {
	return header.Size & 31
}

type Message struct {
	Header  Header
	Payload [MaxPayloadSize]byte
}

func (msg *Message) GetPayload() []byte {
	return msg.Payload[:msg.Header.GetLength()]
}

func (msg *Message) String() string {
	buf := &bytes.Buffer{}
	fmt.Fprintf(buf, "SDEP Packet:\r\n----------\r\n")
	fmt.Fprintf(buf, "Header:\r\n")
	fmt.Fprintf(buf, "  Type: %02X\r\n", msg.Header.Type)
	fmt.Fprintf(buf, "  ID:   %04X\r\n", msg.Header.ID)
	fmt.Fprintf(buf, "  Size: %d (length: %d, moreData: %t)\r\n",
		msg.Header.Size,
		msg.Header.GetLength(),
		msg.Header.HasMoreData(),
	)
	if payload := msg.GetPayload(); payload != nil && len(payload) > 0 {
		fmt.Fprintf(buf, "Payload: % X    \r\n", payload)
	}
	fmt.Fprintf(buf, "----------\r\n")
	return buf.String()
}
