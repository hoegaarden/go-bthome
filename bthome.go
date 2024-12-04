package bthome

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/ghostiam/binstruct"
)

type Parser struct {
	objectParsers map[byte]ObjectParserFunc
	lastPacket    map[string]uint8
}

// NewParser creates a new BTHome parser with the default object parsers
// registered.
func NewParser() *Parser {
	parser := &Parser{}

	parser.lastPacket = map[string]uint8{}

	for _, p := range objectParsers {
		parser.RegisterObjectParser(p)
	}

	return parser
}

// RegisterObjectParser registers an object parser, which parsers parts of the
// BTHome format. If a parser for the same byte is already registered, it will
// be overwritten.
func (p *Parser) RegisterObjectParser(op ObjectParser) {
	if p.objectParsers == nil {
		p.objectParsers = map[byte]ObjectParserFunc{}
	}
	p.objectParsers[op.Byte] = op.Parser
}

// BTHomeUUID is the Service Data UUID of BTHome packets.
var BTHomeUUID = [...]byte{0xD2, 0xFC}

// Parse parses advertisement data as a BTHome packet.
// If the address is not empty, duplicate packets (with the same BTHome packet
// ID) will be ignored.
// Service data must be pairs of UUID and raw service data. If any of UUIDs is
// not a BTHome UUID, it will be ignored. If the UUD is nil, it will be treated
// as if it was a BTHome UUID, this might be useful if the ServiceData has
// already been filtered for BTHome and we don't need to check for that again.
func (p *Parser) Parse(address string, serviceData ...[]byte) ([]Packet, error) {
	l := len(serviceData)
	if l%2 != 0 {
		return nil, fmt.Errorf("serviceData must be pairs of UUID and service data")
	}

	packets := []Packet{}

	for i := 0; i < l; i += 2 {
		uuid, data := serviceData[i], serviceData[i+1]

		if uuid != nil && !isBTHome(uuid) {
			continue
		}

		packet, err := p.parseSingle(data)
		if err != nil {
			return nil, fmt.Errorf("parsing packet: %w", err)
		}

		if address != "" {
			lastPacketID, ok := p.lastPacket[address]
			if isDup := ok && lastPacketID == packet.ID; isDup {
				continue
			}
			p.lastPacket[address] = packet.ID
		}

		packets = append(packets, packet)
	}

	return packets, nil
}

func (p *Parser) parseSingle(b []byte) (Packet, error) {
	r := binstruct.NewReaderFromBytes(b, binary.LittleEndian, false)
	packet := Packet{}

	header, err := r.ReadByte()
	if err != nil {
		return packet, fmt.Errorf("reading header byte: %w", err)
	}

	packet.Encrypted = getBit(header, 0)
	packet.Trigger = Trigger(getBit(header, 2))
	packet.BTHomeVersion = header >> 5

	for {
		objectID, err := r.ReadByte()
		if err == io.EOF {
			break
		}
		if err != nil {
			return packet, fmt.Errorf("reading next byte: %w", err)
		}

		op, ok := p.objectParsers[objectID]
		if !ok {
			return packet, fmt.Errorf("unknown object ID: %x", objectID)
		}

		if err := op(r, &packet); err != nil {
			return packet, fmt.Errorf("parsing object ID %x: %w", objectID, err)
		}
	}

	return packet, nil
}

func isBTHome(serviceDataUUID []byte) bool {
	return bytes.Equal(serviceDataUUID, BTHomeUUID[:])
}

func getBit(b byte, pos int) bool {
	return (b>>pos)&1 == 1
}
