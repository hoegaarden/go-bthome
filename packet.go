package bthome

import "fmt"

type Packet struct {
	Encrypted     bool
	Trigger       Trigger
	BTHomeVersion uint8
	ID            uint8

	FirmwareVersion *string
	DeviceTypeID    *int16

	Battery     []uint8
	Humidity    []uint8
	Temperature []float32
	Button      []Button
}

func (p Packet) String() string {
	s := "BTHomePacket{"

	s += fmt.Sprintf("Encrypted: %t, Trigger: %s, BTHomeVersion: %d, ID: %d", p.Encrypted, p.Trigger, p.BTHomeVersion, p.ID)

	if p.FirmwareVersion != nil {
		s += fmt.Sprintf(", FirmwareVersion: %s", *p.FirmwareVersion)
	}
	if p.DeviceTypeID != nil {
		s += fmt.Sprintf(", DeviceTypeID: %d", *p.DeviceTypeID)
	}
	if len(p.Battery) > 0 {
		s += fmt.Sprintf(", Battery: %d", p.Battery)
	}
	if len(p.Humidity) > 0 {
		s += fmt.Sprintf(", Humidity: %d", p.Humidity)
	}
	if len(p.Temperature) > 0 {
		s += fmt.Sprintf(", Temperature: %0.2f", p.Temperature)
	}
	if len(p.Button) > 0 {
		s += fmt.Sprintf(", Button: %v", p.Button)
	}

	s += "}"
	return s
}
