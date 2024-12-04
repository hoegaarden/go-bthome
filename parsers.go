package bthome

import (
	"fmt"
	"strconv"

	"github.com/ghostiam/binstruct"
)

type ObjectParserFunc func(reader binstruct.Reader, packet *Packet) error

type ObjectParser struct {
	Byte   byte
	Parser ObjectParserFunc
}

var objectParsers = []ObjectParser{
	{
		Byte: 0x00,
		Parser: func(r binstruct.Reader, packet *Packet) error {
			val, err := r.ReadUint8()
			if err != nil {
				return fmt.Errorf("reading uint8 for PacketID: %w", err)
			}
			return setOnce(&packet.ID, val)
		},
	},
	{
		Byte: 0x01,
		Parser: func(r binstruct.Reader, packet *Packet) error {
			val, err := r.ReadUint8()
			if err != nil {
				return fmt.Errorf("reading uint8 for Battery: %w", err)
			}
			return appendTo(&packet.Battery, val)
		},
	},
	{
		Byte: 0x2e,
		Parser: func(r binstruct.Reader, packet *Packet) error {
			val, err := r.ReadUint8()
			if err != nil {
				return fmt.Errorf("reading uint8 for Humidity: %w", err)
			}
			return appendTo(&packet.Humidity, val)
		},
	},
	{
		Byte: 0x45,
		Parser: func(r binstruct.Reader, packet *Packet) error {
			val, err := r.ReadInt16()
			if err != nil {
				return fmt.Errorf("reading int16 for Temperature: %w", err)
			}
			return appendTo(&packet.Temperature, 0.1*float32(val))
		},
	},
	{
		Byte: 0x3a,
		Parser: func(r binstruct.Reader, packet *Packet) error {
			val, err := r.ReadByte()
			if err != nil {
				return fmt.Errorf("reading byte for Button: %w", err)
			}
			return appendTo(&packet.Button, Button(val))
		},
	},
	{
		Byte: 0xf0,
		Parser: func(r binstruct.Reader, packet *Packet) error {
			val, err := r.ReadInt16()
			if err != nil {
				return fmt.Errorf("reading int16 for DeviceTypeID: %w", err)
			}
			return setOnce(&packet.DeviceTypeID, &val)
		},
	},
	{
		Byte:   0xf1,
		Parser: firmwareVersionParser(4),
	},
	{
		Byte:   0xf2,
		Parser: firmwareVersionParser(3),
	},
	{
		Byte: 0x54,
		Parser: func(r binstruct.Reader, packet *Packet) error {
			l, err := r.ReadUint8()
			if err != nil {
				return fmt.Errorf("reading length of Raw: %w", err)
			}
			lr, b, err := r.ReadBytes(int(l))
			if err != nil {
				return fmt.Errorf("reading %d bytes for Raw: %w", l, err)
			}
			if lr != int(l) {
				return fmt.Errorf("reading %d bytes for Raw, only got %d bytes", l, lr)
			}
			return appendTo(&packet.Raw, b)
		},
	},
}

func firmwareVersionParser(length int) ObjectParserFunc {
	return func(r binstruct.Reader, packet *Packet) error {
		ver := ""

		for i := range length {
			val, err := r.ReadUint8()
			if err != nil {
				return fmt.Errorf("reading uint8 %d for firmware version: %w", i, err)
			}
			if i != 0 {
				ver = "." + ver
			}
			ver = strconv.Itoa(int(val)) + ver
		}

		return setOnce(&packet.FirmwareVersion, &ver)
	}
}

func appendTo[T any](dest *[]T, val T) error {
	*dest = append(*dest, val)
	return nil
}

func setOnce[T comparable](dest *T, val T) error {
	var zero T

	if *dest != zero {
		return fmt.Errorf("value already set")
	}

	*dest = val
	return nil
}
