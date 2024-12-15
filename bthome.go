package bthome

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/ghostiam/binstruct"
	ccm "gitlab.com/go-extension/aes-ccm"
)

type Parser struct {
	objectParsers map[byte]ObjectParserFunc
	lastPacket    map[string]uint8
	cipherBlocks  map[string]cipher.Block
}

// NewParser creates a new BTHome parser with the default object parsers
// registered.
func NewParser() *Parser {
	parser := &Parser{}

	parser.lastPacket = map[string]uint8{}
	parser.cipherBlocks = map[string]cipher.Block{}

	for _, p := range objectParsers {
		parser.RegisterObjectParser(p)
	}

	return parser
}

// AddEncryptionKey adds an encryption key for a specific address.
func (p *Parser) AddEncryptionKey(address string, key string) error {
	keyRaw, err := hex.DecodeString(key)
	if err != nil {
		return fmt.Errorf("decoding key: %w", err)
	}

	block, err := aes.NewCipher(keyRaw)
	if err != nil {
		return fmt.Errorf("creating AES cipher: %w", err)
	}

	if p.cipherBlocks == nil {
		p.cipherBlocks = map[string]cipher.Block{}
	}

	p.cipherBlocks[strings.ToLower(address)] = block
	return nil
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

		packet := Packet{}
		bytes := binstruct.NewReaderFromBytes(data, binary.LittleEndian, false)

		header, err := bytes.ReadByte()
		if err != nil {
			return nil, fmt.Errorf("reading header byte: %w", err)
		}

		packet.Encrypted = getBit(header, 0)
		packet.Trigger = Trigger(getBit(header, 2))
		packet.BTHomeVersion = header >> 5

		if packet.Encrypted {
			encryptedBlob, err := bytes.ReadAll()
			if err != nil {
				return nil, fmt.Errorf("reading encrypted data: %w", err)
			}

			decryptedData, err := p.decrypt(address, header, encryptedBlob)
			if err != nil {
				return nil, fmt.Errorf("decrypting packet: %w", err)
			}

			// replace the bytes reader with a reader holding the decrypted data
			bytes = binstruct.NewReaderFromBytes(decryptedData, binary.LittleEndian, false)
		}

		err = p.parseData(bytes, &packet)
		if err != nil {
			return nil, fmt.Errorf("parsing packet data: %w", err)
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

func (p *Parser) parseData(r binstruct.Reader, packet *Packet) error {
	for {
		objectID, err := r.ReadByte()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("reading next byte: %w", err)
		}

		op, ok := p.objectParsers[objectID]
		if !ok {
			return fmt.Errorf("unknown object ID: %x", objectID)
		}

		if err := op(r, packet); err != nil {
			return fmt.Errorf("parsing object ID %x: %w", objectID, err)
		}
	}

	return nil
}

func (p *Parser) decrypt(address string, header byte, encryptedBlob []byte) ([]byte, error) {
	if address == "" {
		return nil, fmt.Errorf("no address provided for decryption")
	}

	addrBytes, err := hex.DecodeString(strings.ReplaceAll(address, ":", ""))
	if err != nil {
		return nil, fmt.Errorf("decoding MAC address: %w", err)
	}

	// +--------------------------+-------------------+---------------+
	// | Encrypted Data (n bytes) | Counter (4 bytes) | MIC (4 bytes) |
	// +--------------------------+-------------------+---------------+

	l := len(encryptedBlob)
	encryptedData := encryptedBlob[:l-8]
	counter := encryptedBlob[l-8 : l-4]
	mic := encryptedBlob[l-4:]

	// This could/should be used to check for replay attacks
	// counter seems to be unix time (?)
	// counterUint32 := binary.LittleEndian.Uint32(counterRaw)

	block, ok := p.cipherBlocks[strings.ToLower(address)]
	if !ok {
		return nil, fmt.Errorf("no decryption cipher block for address %s", address)
	}

	nonce := slices.Concat(addrBytes, BTHomeUUID[:], []byte{header}, counter)

	ccm, err := ccm.NewCCMWithSize(block, len(nonce), 4)
	if err != nil {
		return nil, fmt.Errorf("creating CCM: %w", err)
	}

	decryptedData, err := ccm.Open(nil, nonce, slices.Concat(encryptedData, mic), nil)
	if err != nil {
		return nil, fmt.Errorf("CCM open: %w", err)
	}

	return decryptedData, nil
}

func isBTHome(serviceDataUUID []byte) bool {
	return bytes.Equal(serviceDataUUID, BTHomeUUID[:])
}

func getBit(b byte, pos int) bool {
	return (b>>pos)&1 == 1
}
