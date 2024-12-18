package main

import (
	"encoding/binary"
	"log"

	"github.com/hoegaarden/go-bthome"
	"tinygo.org/x/bluetooth"
)

var adapter = bluetooth.DefaultAdapter

func main() {
	// Enable BLE interface.
	must("enable BLE stack", adapter.Enable())

	bthomeUUID := bluetooth.New16BitUUID(binary.LittleEndian.Uint16(bthome.BTHomeUUID[:]))
	parser := bthome.NewParser()

	err := parser.AddEncryptionKey("7C:C6:B6:76:58:CE", "A9BC33D24E55F1168C9A909939B51D50")
	if err != nil {
		panic(err)
	}

	// Start scanning.
	log.Println("scanning...")
	err = adapter.Scan(func(adapter *bluetooth.Adapter, device bluetooth.ScanResult) {
		if !device.HasServiceUUID(bthomeUUID) {
			return
		}

		addr := device.Address.String()

		for _, sd := range device.ServiceData() {
			packets, err := parser.Parse(addr, nil, sd.Data)
			if err != nil {
				log.Printf("[%s] error: %v\n", addr, err)
				continue
			}
			for _, p := range packets {
				log.Printf("[%s] %s\n", addr, p)
			}
		}
	})
	must("start scan", err)
}

func must(action string, err error) {
	if err != nil {
		panic("failed to " + action + ": " + err.Error())
	}
}
