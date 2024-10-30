package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/go-ble/ble"
	"github.com/go-ble/ble/examples/lib/dev"
	"github.com/hoegaarden/go-bthome"
)

func main() {
	var (
		device string
		macs   strings
	)

	flag.StringVar(&device, "device", "default", "implementation of ble")
	flag.Var(&macs, "mac", "MACs of the BLE devices to listen for")

	flag.Parse()

	d, err := dev.NewDevice(device)
	if err != nil {
		log.Fatalf("can't new device : %s", err)
	}
	ble.SetDefaultDevice(d)

	ctx := ble.WithSigHandler(context.WithCancel(context.Background()))

	log.Printf("Scan on %s starting", device)
	defer func() {
		log.Printf("Scan on device %s stopped", device)
	}()
	err = ble.Scan(ctx, true, advHandler(), advFilter(macs))

	if err != nil && err != context.Canceled {
		panic(err)
	}
}

func toServiceDataPairs(serviceDatas []ble.ServiceData) [][]byte {
	data := [][]byte{}
	for _, serviceData := range serviceDatas {
		data = append(data, serviceData.UUID, serviceData.Data)
	}
	return data
}

func advHandler() func(a ble.Advertisement) {
	parser := bthome.NewParser()

	return func(a ble.Advertisement) {
		addr := a.Addr().String()

		packets, err := parser.Parse(addr, toServiceDataPairs(a.ServiceData())...)
		if err != nil {
			log.Printf("[%s] Error: parsing advertisement: %v", addr, err)
		}

		for _, packet := range packets {
			log.Printf("[%s] %s", addr, packet)
		}
	}
}

func advFilter(macs strings) func(ble.Advertisement) bool {
	if len(macs) == 0 {
		return nil
	}

	macMap := make(map[string]struct{})
	for _, mac := range macs {
		macMap[mac] = struct{}{}
	}

	return func(a ble.Advertisement) bool {
		_, ok := macMap[a.Addr().String()]
		return ok
	}
}

type strings []string

func (s *strings) String() string {
	return fmt.Sprintf("%v", *s)
}

func (s *strings) Set(value string) error {
	*s = append(*s, value)
	return nil
}
