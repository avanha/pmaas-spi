package format

import (
	"math"
	"testing"
)

func TestTemperature(t *testing.T) {
	for name, c := range map[string]struct {
		in   any
		want string
	}{
		"float32":          {float32(21.1), "21.10"},
		"too many places":  {float32(21.123456), "21.12"},
		"float64":          {21.149999999, "21.15"},
		"rounds":           {21.166, "21.17"},
		"whole":            {float32(22), "22.00"},
		"negative":         {-5.04, "-5.04"},
		"negative zero":    {-0.004, "0.00"},
		"an int":           {22, "22.00"},
		"inkbird":          {float32(18.94), "18.94"},
		"not a number":     {"warm", ""},
		"nothing":          {nil, ""},
		"not a number yet": {math.NaN(), ""},
		"infinite":         {math.Inf(1), ""},
	} {
		if got := Temperature(c.in); got != c.want {
			t.Errorf("%s: got %q, want %q", name, got, c.want)
		}
	}
}

func TestFahrenheit(t *testing.T) {
	for in, want := range map[float32]string{0: "32.00", 100: "212.00", -40: "-40.00", 20: "68.00", 37: "98.60", 21.1: "69.98", 5.36: "41.65"} {
		if got := Fahrenheit(in); got != want {
			t.Errorf("Fahrenheit(%v) = %q, want %q", in, got, want)
		}
	}

	if Fahrenheit("x") != "" {
		t.Error("a string was a temperature")
	}
}

func TestHumidity(t *testing.T) {
	for name, c := range map[string]struct {
		in   any
		want string
	}{
		"govee":         {float32(36.6), "36.6"},
		"inkbird":       {float32(39.91), "39.9"},
		"nest":          {45.0, "45.0"},
		"json noise":    {45.123456789, "45.1"},
		"float32 noise": {float32(41.299999), "41.3"},
	} {
		if got := Humidity(c.in); got != c.want {
			t.Errorf("%s: got %q, want %q", name, got, c.want)
		}
	}
}
