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
		"float32":          {float32(21.1), "21.1"},
		"too many places":  {float32(21.123456), "21.1"},
		"float64":          {21.149999999, "21.1"},
		"rounds":           {21.16, "21.2"},
		"whole":            {float32(22), "22.0"},
		"negative":         {-5.04, "-5.0"},
		"negative zero":    {-0.04, "0.0"},
		"an int":           {22, "22.0"},
		"a float32 sum":    {float32(18.94) + float32(0.02), "19.0"},
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
	for in, want := range map[float32]string{0: "32.0", 100: "212.0", -40: "-40.0", 20: "68.0", 37: "98.6", 21.1: "70.0", 5.36: "41.6"} {
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
