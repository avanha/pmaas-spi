package device

import (
	"reflect"
	"time"
)

// ButtonPress is how a button was pressed. The zero value, PressUnknown, is not any of them, so a Button that has never been
// pressed can't be mistaken for one that was.
type ButtonPress int

const (
	PressUnknown ButtonPress = iota
	PressSingle
	PressDouble
	PressTriple
	PressLong
)

func (p ButtonPress) String() string {
	switch p {
	case PressSingle:
		return "single"
	case PressDouble:
		return "double"
	case PressTriple:
		return "triple"
	case PressLong:
		return "long"
	default:
		return "unknown"
	}
}

// Button is the generic, producer-facing representation of a button that is pressed: a wireless one, say, that is used as
// a switch for something else. What a press does is not the button's business: it announces the press (see
// ButtonPressedEvent) and whatever is listening decides.
type Button struct {
	Name string

	// Presses is how many presses have been seen since the plugin started, and LastPress and LastPressTime the latest.
	// Presses is zero, LastPress is PressUnknown and LastPressTime is zero until the button has been pressed.
	Presses       uint64
	LastPress     ButtonPress
	LastPressTime time.Time
}

// IButton is the interface a plugin implements for a button.
type IButton interface {
	GetButtonData() Button
}

// IButtonType is the entity type to register a button under.
var IButtonType = reflect.TypeFor[IButton]()
