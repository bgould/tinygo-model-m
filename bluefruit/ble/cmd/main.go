package main

import (
	"fmt"
	m "machine"
	"strconv"
	"strings"
	"time"

	"github.com/bgould/tinygo-model-m/bluefruit/ble"
	"github.com/tinygo-org/tinygo/src/machine"
)

const (
	csPin  = m.D10
	irqPin = m.D11
)

var (
	spifriend *ble.SPIFriend
)

func main() {

	//var err error

	spi := &m.SPI0
	spi.Configure(m.SPIConfig{LSBFirst: false, Frequency: 1e6})

	println("Starting up...\r")
	time.Sleep(3 * time.Second)

	spifriend = ble.NewSPIFriend(&m.SPI0, csPin, irqPin, m.NoPin)
	spifriend.Begin(ble.SPIFriendConfig{Verbose: false})

	for {
		cli()
	}
}

var (
	console = machine.UART0
	state   = StateInput

	commands map[string]cmdfunc = map[string]cmdfunc{
		"":     cmdfunc(noop),
		"dbg":  cmdfunc(dbg),
		"cs":   cmdfunc(cs),
		"irq":  cmdfunc(irq),
		"send": cmdfunc(send),
		//		"read":  cmdfunc(read),
		//		"echo":  cmdfunc(echo),
		//		"info":  cmdfunc(info),
		//		"reset": cmdfunc(reset),
		//		"check": cmdfunc(check),
	}

	input [consoleBufLen]byte
	debug bool
)

func send(argv []string) {
	if len(argv) < 2 {
		println("Usage: send <AT command>\r")
	}
	command := strings.Join(argv[1:], " ")
	fmt.Printf("\r\n--> %s\r\n", command)
	response, err := spifriend.SendAT(command)
	if err != nil {
		println("Error: ", err.Error(), "\r")
	} else {
		println("\r\n<--\r")
		console.Write(response)
		println("\r")
	}
}

func cs(argv []string) {
	state := csPin.Get()
	if state {
		println("CS is high\r")
	} else {
		println("CS is low\r")
	}
}

func irq(argv []string) {
	state := irqPin.Get()
	if state {
		println("IRQ is high\r")
	} else {
		println("IRQ is low\r")
	}
}

/*
func check(argv []string) {
	if err := spifriend.Check(); err != nil {
		println("Error during check transaction: ", err.Error(), "\r")
	}
}

func read(argv []string) {
	if !irqPin.Get() {
		println("Error: IRQ pin is not asserted\r")
	}
	msg, err := spifriend.ReadPacket()
	if err != nil {
		println("ReadPacket() error: ", err.Error(), "\r")
	}
	println(msg.String(), "\r")
}

func echo(argv []string) {
	if len(argv) != 2 {
		println("Usage: echo <on|off>\r")
	}
	newState := argv[1]
	if newState == "on" {
		spifriend.Echo(true)
		return
	}
	if newState == "off" {
		spifriend.Echo(false)
		return
	}
	println("Usage: echo <on|off>\r")
}

func info(argv []string) {
	var i ble.Info
	var err error
	println("Requesting Bluefruit info...\r")
	if i, err = spifriend.Info(); err != nil {
		println("Error getting Bluefruit info: ", err, "\r")
	}
	fmt.Printf("%s\r\n", i.String())
}

func reset(argv []string) {
	println("Resetting Bluefruit... \r")
	if err := spifriend.ATZ(); err != nil {
		println("error during Bluefruit reset: ", err.Error(), "\r")
	} else {
		println("issued reset successfully\r")
	}
}
*/

type cmdfunc func(argv []string)

const consoleBufLen = 64
const storageBufLen = 512

const (
	StateInput = iota
	StateEscape
	StateEscBrc
	StateCSI
)

func cli() {

	prompt()

	for i := 0; ; {
		if console.Buffered() > 0 {
			data, _ := console.ReadByte()
			if debug {
				fmt.Printf("\rdata: %x\r\n\r", data)
				prompt()
				console.Write(input[:i])
			}
			switch state {
			case StateInput:
				switch data {
				case 0x8:
					fallthrough
				case 0x7f: // this is probably wrong... works on my machine tho :)
					// backspace
					if i > 0 {
						i -= 1
						console.Write([]byte{0x8, 0x20, 0x8})
					}
				case 13:
					// return key
					console.Write([]byte("\r\n"))
					runCommand(string(input[:i]))
					return
				case 27:
					// escape
					state = StateEscape
				default:
					// anything else, just echo the character if it is printable
					if strconv.IsPrint(rune(data)) {
						if i < (consoleBufLen - 1) {
							console.WriteByte(data)
							input[i] = data
							i++
						}
					}
				}
			case StateEscape:
				switch data {
				case 0x5b:
					state = StateEscBrc
				default:
					state = StateInput
				}
			default:
				// TODO: handle escape sequences
				state = StateInput
			}
		}
		//time.Sleep(10 * time.Millisecond)
	}

}

func noop(argv []string) {}

func prompt() {
	print("==> ")
}

func runCommand(line string) {
	argv := strings.SplitN(strings.TrimSpace(line), " ", -1)
	cmd := argv[0]
	cmdfn, ok := commands[cmd]
	if !ok {
		println("unknown command: " + line)
		return
	}
	cmdfn(argv)
}

func dbg(argv []string) {
	if debug {
		debug = false
		println("Console debugging off")
	} else {
		debug = true
		println("Console debugging on")
	}
}
