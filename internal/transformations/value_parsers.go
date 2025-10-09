package transformations

import (
	"fmt"
	"image/color"
	"math"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2/log"
)

func parseBackgroundColor(value string) (color.Color, error) {
	// URL decode the value first (handles %23 -> #, %28 -> (, etc.)
	decodedValue, err := url.QueryUnescape(value)
	if err != nil {
		return nil, fmt.Errorf("failed to decode background color value: %w", err)
	}

	// Handle hex colors (#RRGGBB or #RGB)
	if strings.HasPrefix(decodedValue, "#") {
		return parseHexColor(decodedValue)
	}

	// Handle CSS named colors
	if namedColor := parseNamedColor(decodedValue); namedColor != nil {
		return namedColor, nil
	}

	// Handle rgb() and rgba() functions
	if strings.HasPrefix(decodedValue, "rgb") {
		return parseRGBColor(decodedValue)
	}

	return nil, fmt.Errorf("unsupported color format: %s", decodedValue)
}

func parseHexColor(hex string) (color.Color, error) {
	hex = strings.TrimPrefix(hex, "#")

	var r, g, b, a uint8 = 0, 0, 0, 255

	switch len(hex) {
	case 3: // #RGB -> #RRGGBB
		if parsed, err := strconv.ParseUint(hex, 16, 12); err == nil {
			r = uint8(((parsed >> 8) & 0xF) * 17)
			g = uint8(((parsed >> 4) & 0xF) * 17)
			b = uint8((parsed & 0xF) * 17)
		} else {
			return nil, fmt.Errorf("invalid hex color: #%s", hex)
		}
	case 6: // #RRGGBB
		if parsed, err := strconv.ParseUint(hex, 16, 24); err == nil {
			r = uint8((parsed >> 16) & 0xFF)
			g = uint8((parsed >> 8) & 0xFF)
			b = uint8(parsed & 0xFF)
		} else {
			return nil, fmt.Errorf("invalid hex color: #%s", hex)
		}
	case 8: // #RRGGBBAA
		if parsed, err := strconv.ParseUint(hex, 16, 32); err == nil {
			r = uint8((parsed >> 24) & 0xFF)
			g = uint8((parsed >> 16) & 0xFF)
			b = uint8((parsed >> 8) & 0xFF)
			a = uint8(parsed & 0xFF)
		} else {
			return nil, fmt.Errorf("invalid hex color: #%s", hex)
		}
	default:
		return nil, fmt.Errorf("invalid hex color length: #%s", hex)
	}

	return color.RGBA{r, g, b, a}, nil
}

func parseNamedColor(name string) color.Color {
	colors := map[string]color.RGBA{
		"transparent": {0, 0, 0, 0},
		"black":       {0, 0, 0, 255},
		"white":       {255, 255, 255, 255},
		"red":         {255, 0, 0, 255},
		"green":       {0, 128, 0, 255},
		"blue":        {0, 0, 255, 255},
		"yellow":      {255, 255, 0, 255},
		"cyan":        {0, 255, 255, 255},
		"magenta":     {255, 0, 255, 255},
		"gray":        {128, 128, 128, 255},
		"orange":      {255, 165, 0, 255},
		"purple":      {128, 0, 128, 255},
		"pink":        {255, 192, 203, 255},
		"brown":       {165, 42, 42, 255},
	}

	if c, exists := colors[strings.ToLower(name)]; exists {
		return c
	}
	return nil
}

func parseRGBColor(rgbStr string) (color.Color, error) {
	// Match rgb(r g b) or rgba(r g b a) - CSS4 modern syntax with spaces
	// Also support legacy rgb(r,g,b) and rgba(r,g,b,a) with commas
	rgbRegex := regexp.MustCompile(`rgba?\(\s*(\d+)[\s,]+(\d+)[\s,]+(\d+)(?:[\s,]+(\d+(?:\.\d+)?))?\s*\)`)
	matches := rgbRegex.FindStringSubmatch(rgbStr)

	if len(matches) < 4 {
		return nil, fmt.Errorf("invalid rgb/rgba format: %s", rgbStr)
	}

	r, _ := strconv.Atoi(matches[1])
	g, _ := strconv.Atoi(matches[2])
	b, _ := strconv.Atoi(matches[3])
	a := 255

	if len(matches) > 4 && matches[4] != "" {
		if aFloat, err := strconv.ParseFloat(matches[4], 32); err == nil {
			if aFloat <= 1.0 {
				// Alpha as decimal (0.0-1.0)
				a = int(aFloat * 255)
			} else {
				// Alpha as integer (0-255)
				a = int(aFloat)
			}
		}
	}

	// Clamp values
	if r > 255 {
		r = 255
	}
	if g > 255 {
		g = 255
	}
	if b > 255 {
		b = 255
	}
	if a > 255 {
		a = 255
	}

	return color.RGBA{uint8(r), uint8(g), uint8(b), uint8(a)}, nil
}

func parseFloatValue(value string, min, max, defaultVal float32) float32 {
	if value == "" {
		return defaultVal
	}

	parsed64, err := strconv.ParseFloat(value, 32)
	parsed32 := float32(parsed64)
	if err != nil || parsed32 < min || parsed32 > max {
		return defaultVal
	}
	return parsed32
}

func parseIntValue(value string, min, max int) (int, error) {
	if value == "" {
		return 0, fmt.Errorf("value cannot be empty")
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("value must be a valid integer: %s", value)
	}

	if parsed < min || parsed > max {
		return 0, fmt.Errorf("value must be between %d and %d, got %d", min, max, parsed)
	}

	return parsed, nil
}

func parseRotateAngle(value string) (float32, error) {
	// Handle common rotation shortcuts
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "90", "cw", "right":
		return 90, nil
	case "180", "flip":
		return 180, nil
	case "270", "-90", "ccw", "left":
		return 270, nil
	case "0":
		return 0, nil
	}

	// Parse the string as a float or integer for degrees
	floatVal, err := strconv.ParseFloat(value, 32)
	if err != nil {
		return 0, fmt.Errorf("rotate angle must be a valid number or shortcut (90, 180, 270, cw, ccw, left, right, flip): %s", value)
	}

	// Normalize angle to 0-360 range
	angle := math.Mod(floatVal, 360)
	if angle < 0 {
		angle += 360
	}

	// Optional: warn about non-standard angles that might cause quality loss
	standardAngles := []float64{0, 45, 90, 135, 180, 225, 270, 315}
	isStandard := false
	for _, stdAngle := range standardAngles {
		if math.Abs(angle-stdAngle) < 0.1 {
			isStandard = true
			break
		}
	}

	if !isStandard {
		log.Warnf("Warning: Non-standard rotation angle %f degrees may result in quality loss", angle)
	}

	return float32(angle), nil
}

func parseFormatValue(value string) (string, error) {
	format := strings.ToLower(strings.TrimSpace(value))

	switch format {
	case "jpg", "jpeg":
		return "jpeg", nil
	case "png":
		return "png", nil
	case "webp":
		return "webp", nil
	default:
		return "", fmt.Errorf("unsupported format: %s (supported formats: jpeg, jpg, png, webp)", value)
	}
}

// BorderRadiusValue represents a border radius value that can be in pixels or percentage
type BorderRadiusValue struct {
	Value     float32
	IsPercent bool
}

// parseBorderRadiusValue parses border radius values like "10", "20px", "15%"
func parseBorderRadiusValue(value string) (*BorderRadiusValue, error) {
	if value == "" {
		return nil, fmt.Errorf("border radius value cannot be empty")
	}

	var radiusValue BorderRadiusValue

	if strings.HasSuffix(value, "%") {
		// Percentage value
		percentStr := strings.TrimSuffix(value, "%")
		parsed, parseErr := strconv.ParseFloat(percentStr, 32)
		if parseErr != nil || parsed < 0 || parsed > 50 {
			return nil, fmt.Errorf("percentage value must be between 0%% and 50%%, got %s%%", percentStr)
		}
		radiusValue = BorderRadiusValue{Value: float32(parsed), IsPercent: true}
	} else {
		// Pixel value (with or without "px" suffix)
		pixelStr := strings.TrimSuffix(value, "px")
		parsed, parseErr := strconv.ParseFloat(pixelStr, 32)
		if parseErr != nil || parsed < 0 {
			return nil, fmt.Errorf("pixel value must be a positive number, got %s", value)
		}
		radiusValue = BorderRadiusValue{Value: float32(parsed), IsPercent: false}
	}

	return &radiusValue, fmt.Errorf("unexpected number of radius values")
}
