# Go Idiomatic Improvements for go-bthome

This document outlines suggested improvements to make the go-bthome codebase more idiomatic and follow Go best practices. Each suggestion includes reasoning, diffs, and impact assessment.

## Priority Level: HIGH

### 1. Add Go Documentation Comments

**Issue**: Missing documentation comments on exported types, functions, and methods.

**Go Convention**: All exported identifiers should have doc comments starting with the identifier name.

**Files Affected**: All files

#### bthome.go improvements:

```diff
+// Parser is a BTHome packet parser that can decode BTHome v2 format advertisements.
+// It supports both encrypted and unencrypted packets and maintains state for
+// duplicate packet detection.
 type Parser struct {
 	objectParsers map[byte]ObjectParserFunc
 	lastPacket    map[string]uint8
 	ccms          map[string]aes_ccm.CCM
 }

+// BTHomeUUID is the Service Data UUID used to identify BTHome packets in BLE advertisements.
 var BTHomeUUID = [...]byte{0xD2, 0xFC}
```

#### types.go improvements:

```diff
+// Trigger indicates whether a BTHome packet was triggered by a button press or interval.
 type Trigger bool

 const (
+	// TriggerByButton indicates the packet was triggered by a user button press.
 	TriggerByButton   Trigger = true
+	// TriggerByInterval indicates the packet was triggered by a periodic interval.
 	TriggerByInterval Trigger = false
 )

+// Button represents the type of button event that occurred.
 type Button byte

 const (
+	// ButtonNone indicates no button event.
 	ButtonNone            Button = 0x00
+	// ButtonPress indicates a single button press.
 	ButtonPress           Button = 0x01
 	// ... (continue for all button constants)
```

**Reasoning**: Go doc comments are essential for package documentation and IDE support. They make the API self-documenting and help users understand the purpose of each exported identifier.

---

### 2. Use Functional Options Pattern for Parser Configuration

**Issue**: The `Parser` struct is created with `NewParser()` but doesn't allow for configuration options.

**Current Code**:
```go
func NewParser() *Parser {
	parser := &Parser{
		lastPacket:    map[string]uint8{},
		objectParsers: map[byte]ObjectParserFunc{},
		ccms:          map[string]aes_ccm.CCM{},
	}
	// ...
}
```

**Improved Code**:
```diff
+// ParserOption is a function that configures a Parser.
+type ParserOption func(*Parser) error

+// WithEncryptionKey configures the parser with an encryption key for a specific device.
+func WithEncryptionKey(address, key string) ParserOption {
+	return func(p *Parser) error {
+		return p.AddEncryptionKey(address, key)
+	}
+}

+// WithCustomObjectParser adds a custom object parser to the parser.
+func WithCustomObjectParser(op ObjectParser) ParserOption {
+	return func(p *Parser) error {
+		p.RegisterObjectParser(op)
+		return nil
+	}
+}

+// NewParser creates a new BTHome parser with the given options.
+func NewParser(opts ...ParserOption) (*Parser, error) {
 	parser := &Parser{
 		lastPacket:    make(map[string]uint8),
 		objectParsers: make(map[byte]ObjectParserFunc),
 		ccms:          make(map[string]aes_ccm.CCM),
 	}

 	for _, p := range objectParsers {
 		parser.RegisterObjectParser(p)
 	}

+	for _, opt := range opts {
+		if err := opt(parser); err != nil {
+			return nil, fmt.Errorf("applying parser option: %w", err)
+		}
+	}

+	return parser, nil
 }
```

**Reasoning**: The functional options pattern is a Go idiom for configurable constructors. It provides flexibility while maintaining backward compatibility and makes the API more extensible.

---

### 3. Improve Error Handling and Messages

**Issue**: Error messages don't follow Go conventions and some errors could be wrapped better.

**Current Code**:
```go
return nil, fmt.Errorf("serviceData must be pairs of UUID and service data")
```

**Improved Code**:
```diff
+// Common errors that can be returned by the parser.
+var (
+	ErrInvalidServiceData = errors.New("service data must be pairs of UUID and data")
+	ErrNoEncryptionKey    = errors.New("no encryption key configured for device")
+	ErrDuplicatePacket    = errors.New("duplicate packet detected")
+	ErrUnknownObjectID    = errors.New("unknown object ID")
+)

 func (p *Parser) Parse(address string, serviceData ...[]byte) ([]Packet, error) {
 	l := len(serviceData)
 	if l%2 != 0 {
-		return nil, fmt.Errorf("serviceData must be pairs of UUID and service data")
+		return nil, ErrInvalidServiceData
 	}
```

**Reasoning**: Go convention favors predefined error variables for common errors. This makes error handling more consistent and allows callers to use `errors.Is()` for error checking.

---

### 4. Use More Specific Types Instead of Slices

**Issue**: The `Packet` struct uses slices for values that should be singular or have a known cardinality.

**Current Code**:
```go
type Packet struct {
	// ...
	Battery     []uint8
	Humidity    []uint8
	Temperature []float32
	Raw         [][]byte
	Button      []Button
}
```

**Improved Code**:
```diff
+// SensorReading represents a timestamped sensor value.
+type SensorReading[T any] struct {
+	Value     T
+	Timestamp time.Time // if timing info is available
+}

 type Packet struct {
 	Encrypted     bool
 	Trigger       Trigger
 	BTHomeVersion uint8
 	ID            uint8

 	FirmwareVersion *string
 	DeviceTypeID    *int16

-	Battery     []uint8
-	Humidity    []uint8
-	Temperature []float32
+	Battery     *uint8        // Usually singular
+	Humidity    *uint8        // Usually singular  
+	Temperature *float32      // Usually singular
 	Raw         [][]byte      // Can be multiple
-	Button      []Button
+	ButtonEvents []Button     // Renamed for clarity
 }
```

**Reasoning**: Using specific types makes the API clearer and prevents confusion about whether multiple values are expected. It also reduces memory allocation and improves type safety.

---

### 5. Implement String() Methods More Efficiently

**Issue**: The `Packet.String()` method builds strings inefficiently using concatenation.

**Current Code**:
```go
func (p Packet) String() string {
	s := "BTHomePacket{"
	s += fmt.Sprintf("Encrypted: %t, Trigger: %s, BTHomeVersion: %d, ID: %d", p.Encrypted, p.Trigger, p.BTHomeVersion, p.ID)
	// ... more concatenations
	s += "}"
	return s
}
```

**Improved Code**:
```diff
 func (p Packet) String() string {
-	s := "BTHomePacket{"
-	s += fmt.Sprintf("Encrypted: %t, Trigger: %s, BTHomeVersion: %d, ID: %d", p.Encrypted, p.Trigger, p.BTHomeVersion, p.ID)
+	var b strings.Builder
+	b.WriteString("BTHomePacket{")
+	fmt.Fprintf(&b, "Encrypted: %t, Trigger: %s, BTHomeVersion: %d, ID: %d", 
+		p.Encrypted, p.Trigger, p.BTHomeVersion, p.ID)

 	if p.FirmwareVersion != nil {
-		s += fmt.Sprintf(", FirmwareVersion: %s", *p.FirmwareVersion)
+		fmt.Fprintf(&b, ", FirmwareVersion: %s", *p.FirmwareVersion)
 	}
 	// ... continue pattern for other fields
-	s += "}"
-	return s
+	b.WriteString("}")
+	return b.String()
 }
```

**Reasoning**: Using `strings.Builder` is more efficient than string concatenation as it avoids multiple memory allocations. This is especially important for String() methods that might be called frequently.

---

## Priority Level: MEDIUM

### 6. Add Interfaces for Better Testability

**Issue**: The code is tightly coupled and difficult to test due to lack of interfaces.

**Suggested Interfaces**:
```go
// PacketParser defines the interface for parsing BTHome packets.
type PacketParser interface {
	Parse(address string, serviceData ...[]byte) ([]Packet, error)
	AddEncryptionKey(address, key string) error
	RegisterObjectParser(op ObjectParser)
}

// Ensure Parser implements PacketParser.
var _ PacketParser = (*Parser)(nil)
```

**Reasoning**: Interfaces enable dependency injection, mocking for tests, and follow the Go principle of "accept interfaces, return structs."

---

### 7. Better Package Organization

**Issue**: All types are in the same file, making the package harder to navigate.

**Suggested Structure**:
```
bthome/
├── parser.go          // Parser struct and main parsing logic
├── packet.go          // Packet struct and related types
├── types.go           // Basic types (Trigger, Button)
├── parsers.go         // Object parsers
├── encryption.go      // Encryption-related functionality
├── errors.go          // Error definitions
└── doc.go             // Package documentation
```

**Reasoning**: Separating concerns into different files makes the codebase more maintainable and easier to understand.

---

### 8. Use Context for Cancellation

**Issue**: Long-running operations don't support cancellation.

**Example for the parsing functions**:
```diff
-func (p *Parser) Parse(address string, serviceData ...[]byte) ([]Packet, error) {
+func (p *Parser) ParseWithContext(ctx context.Context, address string, serviceData ...[]byte) ([]Packet, error) {
+	// Check for cancellation periodically during parsing
+	select {
+	case <-ctx.Done():
+		return nil, ctx.Err()
+	default:
+	}
 	// ... rest of parsing logic
 }
```

**Reasoning**: Context support is a Go best practice for operations that might need cancellation or have deadlines.

---

## Priority Level: LOW

### 9. Use Constants for Magic Numbers

**Issue**: Magic numbers in the code make it less readable.

```diff
+const (
+	// BTHome packet structure constants
+	headerByteIndex     = 0
+	encryptionBitPos    = 0
+	triggerBitPos       = 2
+	versionShiftBits    = 5
+	
+	// Encryption constants
+	nonceSize = 13
+	tagSize   = 4
+	micSize   = 4
+	counterSize = 4
+)
```

### 10. Add Validation Methods

**Issue**: No validation for input data.

```go
// IsValid checks if the packet contains valid data.
func (p Packet) IsValid() error {
	if p.BTHomeVersion < 1 || p.BTHomeVersion > 2 {
		return fmt.Errorf("unsupported BTHome version: %d", p.BTHomeVersion)
	}
	// Add more validation as needed
	return nil
}
```

---

## Summary of Improvements

### High Priority (Critical for Go Idioms):
1. **Add documentation comments** - Essential for Go packages
2. **Use functional options pattern** - Standard Go configuration pattern  
3. **Improve error handling** - Use predefined errors and proper wrapping
4. **Fix type design** - Use appropriate types instead of slices everywhere
5. **Optimize String() methods** - Use strings.Builder for efficiency

### Medium Priority (Good Practices):
6. **Add interfaces** - Better testability and modularity
7. **Organize package structure** - Separate concerns into different files
8. **Add context support** - Standard for long-running operations

### Low Priority (Nice to Have):
9. **Use constants for magic numbers** - Improve readability
10. **Add validation methods** - Better error detection

These improvements would make the code more idiomatic Go while maintaining backward compatibility and improving maintainability, testability, and performance.