package ble

import (
	"bytes"
	"fmt"
	m "machine"
	"strconv"
	"time"

	"github.com/bgould/tinygo-model-m/bluefruit/sdep"
)

type Error uint32

func (err Error) Error() string {
	return strconv.Itoa(int(err))
}

const (
	ErrNone Error = iota
	ErrPacketTooLarge

	ErrSlaveDeviceNotReady     Error = Error(sdep.ErrSlaveDeviceNotReady)
	ErrSlaveDeviceReadOverflow       = Error(sdep.ErrSlaveDeviceReadOverflow)
)

type Mode uint8

const (
	CommandMode Mode = iota
	DataMode

	ResponseBufferSize = 2048
)

type SPIFriend struct {
	bus *m.SPI
	cs  m.Pin
	irq m.Pin
	rst m.Pin

	msg     sdep.Message
	buf     *bytes.Buffer
	mode    Mode
	verbose bool
}

type SPIFriendConfig struct {
	Verbose  bool
	Blocking bool
}

func NewSPIFriend(bus *m.SPI, cs m.Pin, irq m.Pin, rst m.Pin) *SPIFriend {
	return &SPIFriend{
		bus:  bus,
		cs:   cs,
		irq:  irq,
		rst:  rst,
		buf:  bytes.NewBuffer(make([]byte, ResponseBufferSize)),
		mode: CommandMode,
	}
}

func (dev *SPIFriend) Begin(config SPIFriendConfig) (err error) {
	dev.verbose = config.Verbose

	dev.irq.Configure(m.PinConfig{Mode: m.PinInput})

	dev.cs.Configure(m.PinConfig{Mode: m.PinOutput})
	dev.cs.High()

	return dev.Reset()
}

func (dev *SPIFriend) Reset() (err error) {

	// Always try to send Initialize command to reset
	// Bluefruit since user can define but not wiring RST signal
	err = dev.sendInitializePattern()

	if dev.rst != m.NoPin {
		dev.rst.Configure(m.PinConfig{Mode: m.PinOutput})
		dev.rst.High()
		dev.rst.Low()
		time.Sleep(10 * time.Millisecond)
		dev.rst.High()
		err = nil
	}

	// _reset_started_timestamp = millis();

	// Bluefruit takes 1 second to reboot
	if dev.verbose {
		dev.debug("waiting 1 second for reset\r")
	}
	time.Sleep(1 * time.Second)
	if dev.verbose {
		dev.debug("returning from Begin()")
	}

	return

}

type timer struct {
	start    int64
	interval int64
}

func newTimer(interval time.Duration) timer {
	return timer{
		start:    time.Now().UnixNano(),
		interval: int64(interval),
	}
}

func (t *timer) Expired() bool {
	return time.Now().UnixNano() > (t.start + t.interval)
}

func (dev *SPIFriend) SendAT(command string) ([]byte, error) {

	defer dev.cs.High()
	dev.buf.Reset()

	cmd := []byte(command)
	for i, c, more := 0, len(command), true; i < c; i += sdep.MaxPayloadSize {
		j := i + sdep.MaxPayloadSize
		if j >= c {
			j = c
			more = false
		}
		err := dev.sendPacket(sdep.CmdTypeATWrapper, cmd[i:j], more)
		if err != nil {
			return nil, err
		}
	}

	delay()
	dev.cs.Low()
	mandatoryDelay()

	t := newTimer(2 * time.Second)

	for !t.Expired() {
		if !dev.irq.Get() {
			continue
		}
		err := dev.readPacket()
		if err != nil {
			/*
				if e, ok := err.(Error); ok {
					if uint32(e) == uint32(ErrSlaveDeviceNotReady) ||
						uint32(e) == uint32(ErrSlaveDeviceReadOverflow) {
			*/
			dev.cs.High()
			delay()
			dev.cs.Low()
			mandatoryDelay()
			continue
			/*
					}
				}
				return nil, err
			*/
		}
		newlen := dev.buf.Len() + int(dev.msg.Header.GetLength())
		if newlen > ResponseBufferSize {
			panic(fmt.Sprintf("response too large for buffer: %s", newlen))
		}
		dev.buf.Write(dev.msg.GetPayload())
		//payload := dev.msg.GetPayload()
		/*
			if dev.verbose {
				dev.debug("AT response: %s", string(payload))
			}
		*/
		if dev.msg.Header.HasMoreData() {
			continue
		} else {
			return dev.buf.Bytes(), nil
		}
	}
	return nil, fmt.Errorf("read timeout")
}

func (dev *SPIFriend) sendInitializePattern() error {
	if dev.verbose {
		dev.debug("entered sendInitializePattern()\r")
	}
	dev.cs.Low()
	defer dev.cs.High()

	return dev.sendPacket(sdep.CmdTypeInitialize, nil, false)
}

func (dev *SPIFriend) sendPacket(cmd uint16, buf []byte, moreData bool) error {

	dev.cs.Low()
	defer dev.cs.High()

	if dev.verbose {
		dev.debug("Sending command packet: %04X", cmd)
	}

	// flush old response before sending the new command, but only if we're *not*
	// in DATA mode, as the RX FIFO may containg incoming UART data that hasn't
	// been read yet
	//if (more_data == 0 && _mode != BLUEFRUIT_MODE_DATA) flush();

	length := 0
	if buf != nil {
		length = len(buf)
		if length > sdep.MaxPacketSize {
			return ErrPacketTooLarge
		}
	}

	dev.msg.Header.Type = sdep.MsgTypeCommand
	dev.msg.Header.ID = cmd
	dev.msg.Header.Size = 0

	if buf != nil {
		n := uint8(len(buf))
		if n > sdep.MaxPayloadSize {
			return ErrPacketTooLarge
		}
		dev.msg.Header.Size = n
		if n == sdep.MaxPayloadSize {
			if moreData {
				dev.msg.Header.Size |= uint8(1 << 7)
			}
		}
		copy(dev.msg.Payload[:n], buf)
	}

	if dev.verbose {
		dev.debug("sending %s", dev.msg.String())
	}

	var b byte
	// Loop until Bluefruit is ready
	for i := 0; i < 25; i++ {
		b, _ := dev.bus.Transfer(byte(dev.msg.Header.Type))
		if b != uint8(ErrSlaveDeviceNotReady) {
			break
		}
		if dev.verbose {
			dev.debug("bluefruit not ready")
		}
		dev.cs.High()
		mandatoryDelay()
		dev.cs.Low()
	}
	if b == uint8(ErrSlaveDeviceNotReady) {
		return fmt.Errorf("write timeout")
	}

	// send the rest of the data
	dev.bus.Transfer(byte(uint8(dev.msg.Header.ID & 0xFF)))
	dev.bus.Transfer(byte(uint8(dev.msg.Header.ID >> 0x8)))
	if length > 0 {
		dev.bus.Transfer(byte(dev.msg.Header.Size))
		dev.bus.Tx(buf, nil)
	} else {
		dev.bus.Transfer(0xFF)
	}

	if dev.verbose {
		dev.debug("Finished sending command packet")
	}

	return nil
}

func (dev *SPIFriend) readPacket() (err error) {
	if dev.verbose {
		dev.debug("Attempting to read packet")
	}

	dev.msg.Header.Type, _ = dev.bus.Transfer(0xFF)
	dev.msg.Header.ID = 0
	dev.msg.Header.Size = 0

	if dev.verbose {
		dev.debug("read type byte: %02X", dev.msg.Header.Type)
	}

	switch dev.msg.Header.Type {
	case sdep.MsgTypeCommand:
		err = fmt.Errorf("Unexpected message type: command")
	case sdep.MsgTypeResponse:
		err = nil
	case sdep.MsgTypeAlert:
		err = fmt.Errorf("Unexpected message type: alert")
	case sdep.MsgTypeError:
		err = fmt.Errorf("Unexpected message type: error")
	case sdep.ErrSlaveDeviceNotReady:
		return ErrSlaveDeviceNotReady
	case sdep.ErrSlaveDeviceReadOverflow:
		return ErrSlaveDeviceReadOverflow
	default:
		return fmt.Errorf("Unexpected byte from slave device: %02X", dev.msg.Header.Type)
	}

	buf := make([]byte, 3)
	dev.bus.Tx(nil, buf)
	dev.msg.Header.ID = uint16(buf[0])
	dev.msg.Header.ID |= uint16(buf[1]) << 8
	dev.msg.Header.Size = uint8(buf[2])
	if dev.msg.Header.Type != sdep.MsgTypeError {
		length := dev.msg.Header.GetLength()
		if length > sdep.MaxPayloadSize {
			length = sdep.MaxPayloadSize // TODO: this should probably be handled better
		}
		if length > 0 {
			dev.bus.Tx(nil, dev.msg.Payload[:length])
		}
	}

	return
}

func (dev *SPIFriend) debug(format string, args ...interface{}) {
	fmt.Printf("[SPIFRIEND %d] ", time.Now().UnixNano())
	fmt.Printf(format, args...)
	println("\r")
}

func mandatoryDelay() {
	delayMicros(100)
	//	time.Sleep(250 * time.Microsecond)
}

func delay() {
	delayMicros(10)
}

func delayMicros(usecs uint32) {
	t := newTimer(time.Duration(usecs) * time.Microsecond)
	for !t.Expired() {
	}
}

/*
// this is probably stupid (also probably wrong)
// TODO: actually take the time to count out the cycles and loop in asm
func delayMicros(usecs uint32) {
	for ; usecs > 0; usecs-- {
		arm.Asm("nop")
		arm.Asm("nop")
		arm.Asm("nop")
		arm.Asm("nop")
		arm.Asm("nop")
		arm.Asm("nop")
		arm.Asm("nop")
		arm.Asm("nop")
		arm.Asm("nop")
		arm.Asm("nop")
		arm.Asm("nop")
		arm.Asm("nop")
		arm.Asm("nop")
		arm.Asm("nop")
		arm.Asm("nop")
		arm.Asm("nop")
		arm.Asm("nop")
		arm.Asm("nop")
		arm.Asm("nop")
		arm.Asm("nop")
		arm.Asm("nop")
		arm.Asm("nop")
		arm.Asm("nop")
		arm.Asm("nop")
		arm.Asm("nop")
		arm.Asm("nop")
		arm.Asm("nop")
		arm.Asm("nop")
		arm.Asm("nop")
		arm.Asm("nop")
		arm.Asm("nop")
		arm.Asm("nop")
		arm.Asm("nop")
		arm.Asm("nop")
		arm.Asm("nop")
		arm.Asm("nop")
		arm.Asm("nop")
		arm.Asm("nop")
		arm.Asm("nop")
		arm.Asm("nop")
		arm.Asm("nop")
		arm.Asm("nop")
		arm.Asm("nop")
		arm.Asm("nop")
		arm.Asm("nop")
		arm.Asm("nop")
		arm.Asm("nop")
	}
}
*/
