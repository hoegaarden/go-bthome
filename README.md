# BTHome & golang

## What

This allows you to parse advertisements from BTHome compatible devices, e.g. a
[Shelly BLU H&T sensor][blu-ht]. It is an (incomplete) implementation of the [BTHome v2
format][bthomev2]. It relies on other low-level bits to do the bluetooth / BLE
communication, e.g. [go-ble].

## Status

- incomplete:
  - I only implemented what I needed for my sensors
  - Encryption is not implemented at all
- PoC: This is just a quick & dirty PoC, use at your on risk


## Examples

In the example in [cmd/readSensors][example] you find an implementation of a
tool which listens for advertisements of sensors and just prints the parsed
BTHome data:

```terminal
: sudo go run cmd/readSensors/main.go -mac 7c:c6:b6:71:9d:b5 -mac 7c:c6:b6:71:98:d7
2024/11/01 15:40:14 Scan on default starting
2024/11/01 15:40:28 [7c:c6:b6:71:9d:b5] BTHomePacket{Encrypted: false, Trigger: Button, BTHomeVersion: 2, ID: 222, Battery: [100], Humidity: [59], Temperature: [20.80], Button: [Press]}
2024/11/01 15:40:37 [7c:c6:b6:71:98:d7] BTHomePacket{Encrypted: false, Trigger: Button, BTHomeVersion: 2, ID: 208, Battery: [100], Humidity: [53], Temperature: [22.00]}
^C
2024/11/01 15:40:39 Scan on device default stopped
: 
```

There is another example in [cmd/bluezScan][exampleBluez], which uses
[tiny-go/bluetooth][tgb] which in turn uses [BlueZ] under the hood (on Linux),
and implements a similar tool as the example above:

```terminal
: go run cmd/bluezScan/main.go
2024/11/02 11:31:20 scanning...
2024/11/02 11:31:37 [7C:C6:B6:71:98:D7] BTHomePacket{Encrypted: false, Trigger: Button, BTHomeVersion: 2, ID: 119, Battery: [100], Humidity: [55], Temperature: [19.90]}
2024/11/02 11:32:09 [7C:C6:B6:71:9D:B5] BTHomePacket{Encrypted: false, Trigger: Button, BTHomeVersion: 2, ID: 144, Battery: [100], Humidity: [56], Temperature: [20.10]}
^Csignal: interrupt
: 
```

[blu-ht]: https://www.shelly.com/products/shelly-blu-h-t-black
[example]: ./cmd/readSensors/main.go
[go-ble]: https://github.com/go-ble/ble/
[bthomev2]: https://bthome.io/format/
[exampleBluez]: ./cmd/bluezScan/main.go
[tgb]: https://github.com/tinygo-org/bluetooth
[BlueZ]: https://www.bluez.org/
