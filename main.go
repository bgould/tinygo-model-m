package main

import (
	"fmt"
	"time"

	m "machine"

	"github.com/bgould/tinygo-model-m/bluefruit/ble"
	"github.com/bgould/tinygo-model-m/bluefruit/ezkey"
	"github.com/bgould/tinygo-model-m/keyboard"
	"github.com/bgould/tinygo-model-m/mcp23008"
	"github.com/bgould/tinygo-model-m/modelm"
	"github.com/bgould/tinygo-model-m/timer"
)

const _debug = false

var (
	console = m.UART0

	uart = m.UART1
	tx   = m.D10
	rx   = m.D11

	spi    = &m.SPI0
	csPin  = m.A5
	irqPin = m.A4

	// configure 2 MCP23008 port expanders for reading the columns in each row
	wire  = m.I2C0
	port1 = mcp23008.New(wire, 0x0)
	port2 = mcp23008.New(wire, 0x1)

	// TODO: there are 8 pins and we're reading a byte... should see if there is
	//       a port that could be used to read these in a single operation
	pins = []m.Pin{m.A0, m.A1, m.A2, m.A3, m.D11, m.D10, m.D9, m.D6}

	kbd *keyboard.Keyboard
)

func main() {

	//uart.Configure(m.UARTConfig{TX: tx, RX: rx, BaudRate: 9600})
	//host := &EZKeyHost{ezkey.New(uart, m.NoPin)}

	spi.Configure(m.SPIConfig{LSBFirst: false, Frequency: 1e6})
	spifriend := ble.NewSPIFriend(spi, csPin, irqPin, m.NoPin)
	spifriend.Begin(ble.SPIFriendConfig{Verbose: false})

	host := &BluefruitLEHost{spifriend}
	host.Init()

	matrix := keyboard.NewMatrix(keyboard.RowReaderFunc(ReadRow))
	layers := []keyboard.Keymap{modelm.ANSI101DefaultLayer()}
	kbd = keyboard.New(console, host, matrix, layers).WithDebug(_debug)

	configurePins()
	configurePortExpanders()

	for {
		kbd.Task()
		//time.Sleep(500 * time.Microsecond)
	}

}

// configurePins sets up the pins that will strobe the rows as outputs
func configurePins() {
	for _, pin := range pins {
		pin.Configure(m.PinConfig{Mode: m.PinOutput})
		pin.High()
	}
}

// configurePortExpanders sets up the IO expanders that read the columns
func configurePortExpanders() {

	// set up the I2C bus
	wire.Configure(m.I2CConfig{Frequency: m.TWI_FREQ_400KHZ})

	// enable pullups on all GPIOs
	port1.WriteByte(mcp23008.GPPU, 0xFF)
	port2.WriteByte(mcp23008.GPPU, 0xFF)

	// set all GPIOs as inputs (even though this is power-on default anyhow)
	port1.WriteByte(mcp23008.IODIR, 0xFF)
	port2.WriteByte(mcp23008.IODIR, 0xFF)

}

func ReadRow(rowIndex uint8) keyboard.Row {
	selectRows(uint8(1) << rowIndex)
	delayMicros(50)
	b := readRow(rowIndex)
	selectRows(0)
	return keyboard.Row(b)
}

func selectRows(state uint8) {
	for i, pin := range pins {
		pinState := state&uint8(1<<uint8(i)) == 0
		pin.Set(pinState)
	}
}

func readRow(rowIndex uint8) keyboard.Row {
	var b uint16
	b |= uint16(port1.ReadByte(mcp23008.GPIO))
	b |= uint16(port2.ReadByte(mcp23008.GPIO)) << 8
	return keyboard.Row(^b)
}

type BluefruitLEHost struct {
	spifriend *ble.SPIFriend
}

func (host *BluefruitLEHost) Init() {
	host.spifriend.SendAT("AT+GAPDEVNAME=TinyGo Model M Keyboard")
	host.spifriend.SendAT("AT+BLEKEYBOARDEN=1")
	host.spifriend.SendAT("ATZ")
}

func (host *BluefruitLEHost) Send(rpt *keyboard.Report) {
	cmd := fmt.Sprintf(
		"AT+BLEKEYBOARDCODE=%02X-%02X-%02X-%02X-%02X-%02X-%02X-%02X",
		rpt[0], rpt[1], rpt[2], rpt[3], rpt[4], rpt[5], rpt[6], rpt[7],
	)
	debug("--> %s\r\n", cmd)
	rsp, err := host.spifriend.SendAT(cmd)
	if err != nil {
		debug("<-- (err) %s\r\n", err.Error())
	}
	debug("<-- %s\r\n", string(rsp))
}

type EZKeyHost struct {
	hid *ezkey.HID
}

func (host *EZKeyHost) Send(report *keyboard.Report) {
	rpt := ezkey.Report(*report)
	host.hid.Send(&rpt)
}

//go:inline
func debug(format string, args ...interface{}) {
	if _debug {
		fmt.Fprintf(console, format, args...)
	}
}

func delayMicros(usecs uint32) {
	timer.Wait(time.Duration(usecs) * time.Microsecond)
}
