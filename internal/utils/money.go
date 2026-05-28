package utils

import "math"

// MicrosPerUnit is the Google Ads micros scale: 1 currency unit == 1,000,000 micros.
const MicrosPerUnit = 1_000_000

// MicrosToUnits converts a *micros value to a *float64 currency amount rounded
// to two decimal places. nil in, nil out.
func MicrosToUnits(micros *int64) *float64 {
	if micros == nil {
		return nil
	}
	v := float64(*micros) / float64(MicrosPerUnit)
	v = math.Round(v*100) / 100
	return &v
}

// UnitsToMicros converts a currency amount to micros.
func UnitsToMicros(units float64) int64 {
	return int64(math.Round(units * float64(MicrosPerUnit)))
}
