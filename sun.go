package openinghours

import (
	"math"
	"time"
)

// Default times when no coordinates are set
const (
	defaultSunrise = 6 * 60      // 06:00
	defaultSunset  = 18 * 60     // 18:00
	defaultDawn    = 5*60 + 30   // 05:30
	defaultDusk    = 18*60 + 30  // 18:30
)

// calculateSunrise returns minutes from midnight for sunrise in UTC
// Uses a simplified astronomical algorithm
func calculateSunrise(t time.Time, lat, lon float64) int {
	// Simplified sunrise calculation
	// For a real implementation, use a proper astronomy library like go-sunrise
	// This uses a basic approximation

	dayOfYear := t.YearDay()

	// Calculate solar declination (simplified)
	// Uses the formula: δ = 23.45° * sin(2π * (284 + N) / 365)
	declination := 23.45 * math.Sin(2*math.Pi*(284+float64(dayOfYear))/365)

	// Calculate hour angle for sunrise
	latRad := lat * math.Pi / 180
	decRad := declination * math.Pi / 180

	// cos(hour angle) = -tan(latitude) * tan(declination)
	cosHourAngle := -math.Tan(latRad) * math.Tan(decRad)

	// Handle polar day/night
	if cosHourAngle < -1 {
		// Sun never sets (midnight sun)
		return 0
	}
	if cosHourAngle > 1 {
		// Sun never rises (polar night) - use noon as fallback
		return 720
	}

	// Calculate hour angle in degrees
	hourAngle := math.Acos(cosHourAngle) * 180 / math.Pi

	// Calculate approximate equation of time (in minutes)
	// This accounts for Earth's elliptical orbit
	B := 2 * math.Pi * (float64(dayOfYear) - 81) / 365
	eqTime := 9.87*math.Sin(2*B) - 7.53*math.Cos(B) - 1.5*math.Sin(B)

	// Solar noon at this longitude (in UTC)
	// Solar noon at prime meridian is at 12:00 UTC
	// For each degree east, solar noon is 4 minutes earlier
	solarNoon := 12*60 - lon*4 - eqTime

	// Sunrise is solar noon minus half the day length
	// Day length in hours = 2 * hourAngle / 15
	dayLengthMinutes := 2 * hourAngle * 4 // hourAngle in degrees * 4 min/degree
	sunriseMinutes := int(solarNoon - dayLengthMinutes/2)

	// Normalize to 0-1440 range
	for sunriseMinutes < 0 {
		sunriseMinutes += 1440
	}
	for sunriseMinutes >= 1440 {
		sunriseMinutes -= 1440
	}

	return sunriseMinutes
}

// calculateSunset returns minutes from midnight for sunset in UTC
func calculateSunset(t time.Time, lat, lon float64) int {
	dayOfYear := t.YearDay()

	// Calculate solar declination (simplified)
	declination := 23.45 * math.Sin(2*math.Pi*(284+float64(dayOfYear))/365)

	// Calculate hour angle for sunset
	latRad := lat * math.Pi / 180
	decRad := declination * math.Pi / 180

	cosHourAngle := -math.Tan(latRad) * math.Tan(decRad)

	// Handle polar day/night
	if cosHourAngle < -1 {
		// Sun never sets (midnight sun)
		return 1440
	}
	if cosHourAngle > 1 {
		// Sun never rises (polar night) - use noon as fallback
		return 720
	}

	hourAngle := math.Acos(cosHourAngle) * 180 / math.Pi

	// Calculate approximate equation of time (in minutes)
	B := 2 * math.Pi * (float64(dayOfYear) - 81) / 365
	eqTime := 9.87*math.Sin(2*B) - 7.53*math.Cos(B) - 1.5*math.Sin(B)

	// Solar noon at this longitude (in UTC)
	solarNoon := 12*60 - lon*4 - eqTime

	// Sunset is solar noon plus half the day length
	dayLengthMinutes := 2 * hourAngle * 4 // hourAngle in degrees * 4 min/degree
	sunsetMinutes := int(solarNoon + dayLengthMinutes/2)

	// Normalize to 0-1440 range
	for sunsetMinutes < 0 {
		sunsetMinutes += 1440
	}
	for sunsetMinutes >= 1440 {
		sunsetMinutes -= 1440
	}

	return sunsetMinutes
}

// calculateDawn returns minutes from midnight for civil dawn
// Civil dawn is when the sun is 6° below the horizon
func calculateDawn(t time.Time, lat, lon float64) int {
	// Simplified: dawn is approximately 30 minutes before sunrise
	sunrise := calculateSunrise(t, lat, lon)
	dawn := sunrise - 30
	if dawn < 0 {
		dawn += 1440
	}
	return dawn
}

// calculateDusk returns minutes from midnight for civil dusk
// Civil dusk is when the sun is 6° below the horizon
func calculateDusk(t time.Time, lat, lon float64) int {
	// Simplified: dusk is approximately 30 minutes after sunset
	sunset := calculateSunset(t, lat, lon)
	dusk := sunset + 30
	if dusk >= 1440 {
		dusk -= 1440
	}
	return dusk
}
