package format

import (
	"fmt"
	"math"
)

// The measurements a page shows are all said the same way: one decimal place. The sensors are not better than a tenth of a degree
// or a tenth of a percent, and a number with more only looks as if they were. These take any number (a float32 reading, a float64
// from JSON, an int) so a template can use them whatever a plugin's data has, and say nothing for what is not a number.

// Temperature is a temperature in Celsius as shown, to a tenth: "21.1".
func Temperature(v any) string {
	f, ok := number(v)
	if !ok {
		return ""
	}

	return tenths(f)
}

// Fahrenheit is a temperature given in Celsius, shown in Fahrenheit to a tenth: Fahrenheit(21.1) is "70.0".
func Fahrenheit(v any) string {
	f, ok := number(v)
	if !ok {
		return ""
	}

	return tenths(f*9/5 + 32)
}

// Humidity is a relative humidity as shown, to a tenth: "36.6".
func Humidity(v any) string {
	f, ok := number(v)
	if !ok {
		return ""
	}

	return tenths(f)
}

// tenths is the number to one decimal place, never as "-0.0".
func tenths(f float64) string {
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return ""
	}

	text := fmt.Sprintf("%.1f", f)
	if text == "-0.0" {
		return "0.0"
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
