package mcp23008

import (
	"fmt"
	"machine"
)

const _debug = false

const (
	Address = 0x20

	// I/O Direction Register
	IODIR = 0x00

	// Input Polarity Register
	IPOL = 0x01

	// Interrupt-on-Change Control Register
	GPINTEN = 0x02

	// Default Compare Register for Interrupt-on-Change
	DEFVAL = 0x03

	// Interrupt Control Register
	INTCON = 0x04

	// Configuration Register
	IOCON = 0x05

	// Pullup Resistor Configuration Register
	GPPU = 0x06

	// Interrupt Flag Register
	INTF = 0x07

	// Interrupt Capture Register
	INTCAP = 0x08

	// Port Register
	GPIO = 0x09

	// Output Latch Register
	OLAT = 0x0A
)

type Device struct {
	bus  *machine.I2C
	buf  []byte
	addr uint8
}

func New(i2c machine.I2C, addressBits uint8) Device {
	return Device{
		bus:  &i2c,
		buf:  make([]byte, 1),
		addr: Address | (addressBits & 0x7),
	}
}

func (d *Device) WriteByte(reg uint8, data byte) {
	d.buf[0] = data
	if _debug {
		fmt.Printf("writing %02X %02X %02X\r\n", d.addr, reg, d.buf[0])
	}
	d.bus.WriteRegister(d.addr, reg, d.buf)
}

func (d *Device) ReadByte(reg uint8) byte {
	d.bus.ReadRegister(d.addr, reg, d.buf)
	if _debug {
		fmt.Printf("reading %02X %02X %02X\r\n", d.addr, reg, d.buf[0])
	}
	return d.buf[0]
}
