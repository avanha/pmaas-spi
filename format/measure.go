package format

import (
	"fmt"
	"math"
	"strings"
)

// The measurements a page shows are all said the same way: a temperature to two decimal places and a humidity to one. A number with
// more only looks more exact than a sensor is. These take any number (a float32 reading, a float64 from JSON, an int) so a template
// can use them whatever a plugin's data has, and say nothing for what is not a number.

// Temperature is a temperature in Celsius as shown, to a hundredth: "21.10".
func Temperature(v any) string {
	f, ok := number(v)
	if !ok {
		return ""
	}

	return fixed(f, 2)
}

// Fahrenheit is a temperature given in Celsius, shown in Fahrenheit to a hundredth: Fahrenheit(21.1) is "69.98".
func Fahrenheit(v any) string {
	f, ok := number(v)
	if !ok {
		return ""
	}

	return fixed(f*9/5+32, 2)
}

// Humidity is a relative humidity as shown, to a tenth: "36.6".
func Humidity(v any) string {
	f, ok := number(v)
	if !ok {
		return ""
	}

	return fixed(f, 1)
}

// fixed is the number to that many decimal places, never as "-0.0".
func fixed(f float64, places int) string {
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return ""
	}

	text := fmt.Sprintf("%.*f", places, f)
	if strings.Trim(text, "-0.") == "" {
		return strings.TrimPrefix(text, "-")
	}

	return text
}

func number(v any) (float64, bool) {
	switch n := v.(type) {
	case float32:
		// Through the shortest decimal form, so 21.1 is 21.1 and not 21.100000381.
		var f float64

		if _, err := fmt.Sscanf(fmt.Sprint(n), "%g", &f); err != nil {
			return float64(n), true
		}

		return f, true
	case float64:
		return n, true
	case int:
		return float64(n), true
	case int32:
		return float64(n), true
	case int64:
		return float64(n), true
	default:
		return 0, false
	}
}
