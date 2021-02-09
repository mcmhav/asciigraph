package asciigraph

import (
	"bytes"
	"fmt"
	"math"
	"strings"
)

var black = "\x1b[30m"
var red = "\x1b[31m"
var green = "\x1b[32m"
var yellow = "\x1b[33m"
var blue = "\x1b[34m"
var magenta = "\x1b[35m"
var cyan = "\x1b[36m"
var lightgray = "\x1b[37m"
var defaultColor = "\x1b[39m"
var darkgray = "\x1b[90m"
var lightred = "\x1b[91m"
var lightgreen = "\x1b[92m"
var lightyellow = "\x1b[93m"
var lightblue = "\x1b[94m"
var lightmagenta = "\x1b[95m"
var lightcyan = "\x1b[96m"
var white = "\x1b[97m"
var reset = "\x1b[0m"

// colorChar
func colorChar(character string, color string) string {
  if color != "" {
    return color + character + reset
  }
  return character
}

// Plot returns ascii graph for a series.
func Plot(series []float64, options ...Option) string {
	var logMaximum float64
	config := configure(config{
		Offset: 3,
	}, options)

	if config.Width > 0 {
		series = interpolateArray(series, config.Width)
	}

	minimum, maximum := minMaxFloat64Slice(series)
	interval := math.Abs(maximum - minimum)

	if config.Height <= 0 {
		if int(interval) <= 0 {
			config.Height = int(interval * math.Pow10(int(math.Ceil(-math.Log10(interval)))))
		} else {
			config.Height = int(interval)
		}
	}

	if config.Offset <= 0 {
		config.Offset = 3
	}

	var ratio float64
	if interval != 0 {
		ratio = float64(config.Height) / interval
	} else {
		ratio = 1
	}
	min2 := round(minimum * ratio)
	max2 := round(maximum * ratio)

	intmin2 := int(min2)
	intmax2 := int(max2)

	rows := int(math.Abs(float64(intmax2 - intmin2)))
	width := len(series) + config.Offset

	plot := make([][]string, rows+1)

	// initialise empty 2D grid
	for i := 0; i < rows+1; i++ {
		line := make([]string, width)
		for j := 0; j < width; j++ {
			line[j] = " "
		}
		plot[i] = line
	}

	precision := 2
	logMaximum = math.Log10(math.Max(math.Abs(maximum), math.Abs(minimum))) //to find number of zeros after decimal
	if minimum == float64(0) && maximum == float64(0) {
		logMaximum = float64(-1)
	}

	if logMaximum < 0 {
		// negative log
		if math.Mod(logMaximum, 1) != 0 {
			// non-zero digits after decimal
			precision += int(math.Abs(logMaximum))
		} else {
			precision += int(math.Abs(logMaximum) - 1.0)
		}
	} else if logMaximum > 2 {
		precision = 0
	}

	maxNumLength := len(fmt.Sprintf("%0.*f", precision, maximum))
	minNumLength := len(fmt.Sprintf("%0.*f", precision, minimum))
	maxWidth := int(math.Max(float64(maxNumLength), float64(minNumLength)))

	// axis and labels
	for y := intmin2; y < intmax2+1; y++ {
		var magnitude float64
		if rows > 0 {
			magnitude = maximum - (float64(y-intmin2) * interval / float64(rows))
		} else {
			magnitude = float64(y)
		}

		label := fmt.Sprintf("%*.*f", maxWidth+1, precision, magnitude)
		w := y - intmin2
		h := int(math.Max(float64(config.Offset)-float64(len(label)), 0))

		plot[w][h] = label
		if y == 0 {
			plot[w][config.Offset-1] = "┼"
		} else {
			plot[w][config.Offset-1] = "┤"
		}
	}

	y0 := int(round(series[0]*ratio) - min2)
	var y1 int

	plot[rows-y0][config.Offset-1] = "┼" // first value

  var color = blue
	for x := 0; x < len(series)-1; x++ { // plot the line

		d0 := series[x]
		d1 := series[x+1]

		if math.IsNaN(d0) && math.IsNaN(d1) {
			continue
		}

		if math.IsNaN(d1) && !math.IsNaN(d0) {
			y0 = int(round(d0*ratio) - float64(intmin2))
			plot[rows-y0][x+config.Offset] = "╴"
			continue
		}

		if math.IsNaN(d0) && !math.IsNaN(d1) {
			y1 = int(round(d1*ratio) - float64(intmin2))
			plot[rows-y1][x+config.Offset] = "╶"
			continue
		}

		y0 = int(round(d0*ratio) - float64(intmin2))
		y1 = int(round(d1*ratio) - float64(intmin2))

		if y0 == y1 {
			plot[rows-y0][x+config.Offset] = colorChar("─", color)
		} else {
			if y0 > y1 {
				plot[rows-y1][x+config.Offset] = colorChar("╰", color)
				plot[rows-y0][x+config.Offset] = colorChar("╮", color)
			} else {
				plot[rows-y1][x+config.Offset] = colorChar("╭", color)
				plot[rows-y0][x+config.Offset] = colorChar("╯", color)
			}

			start := int(math.Min(float64(y0), float64(y1))) + 1
			end := int(math.Max(float64(y0), float64(y1)))
			for y := start; y < end; y++ {
				plot[rows-y][x+config.Offset] = colorChar("│", color)
			}
		}
	}

	// join columns
	var lines bytes.Buffer
	for h, horizontal := range plot {
		if h != 0 {
			lines.WriteRune('\n')
		}

		// remove trailing spaces
		lastCharIndex := 0
		for i := width - 1; i >= 0; i-- {
			if horizontal[i] != " " {
				lastCharIndex = i
				break
			}
		}

		for _, v := range horizontal[:lastCharIndex+1] {
			lines.WriteString(v)
		}
	}

	// add caption if not empty
	if config.Caption != "" {
		lines.WriteRune('\n')
		lines.WriteString(strings.Repeat(" ", config.Offset+maxWidth))
		if len(config.Caption) < len(series) {
			lines.WriteString(strings.Repeat(" ", (len(series)-len(config.Caption))/2))
		}
		lines.WriteString(config.Caption)
	}

	return lines.String()
}
