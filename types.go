package bthome

import "fmt"

type Trigger bool

const (
	TriggerByButton   Trigger = true
	TriggerByInterval Trigger = false
)

func (t Trigger) String() string {
	if t == TriggerByButton {
		return "Button"
	}
	return "Interval"
}

type Button byte

const (
	ButtonNone            Button = 0x00
	ButtonPress           Button = 0x01
	ButtonDoublePress     Button = 0x02
	ButtonTriplePress     Button = 0x03
	ButtonLongPress       Button = 0x04
	ButtonLongDoublePress Button = 0x05
	ButtonLongTriplePress Button = 0x06
	ButtonHold            Button = 0x80
)

func (b Button) String() string {
	switch b {
	case ButtonNone:
		return "None"
	case ButtonPress:
		return "Press"
	case ButtonDoublePress:
		return "DoublePress"
	case ButtonTriplePress:
		return "TriplePress"
	case ButtonLongPress:
		return "LongPress"
	case ButtonLongDoublePress:
		return "LongDoublePress"
	case ButtonLongTriplePress:
		return "LongTriplePress"
	case ButtonHold:
		return "Hold"
	default:
		return fmt.Sprintf("Unknown (%b)", b)
	}
}
