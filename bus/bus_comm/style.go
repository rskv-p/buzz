// file: buz/bus/bus_comm/style.go

package bus_comm

import (
	"io"

	"github.com/charmbracelet/lipgloss"
)

//---------------------
// Constants - Color Definitions
//---------------------

// Color constants for styling logs
const (
	ColorTeal40    = "#3ddbd9"
	ColorBlue40    = "#78a9ff"
	ColorBlue60    = "#4589ff"
	ColorBlue70    = "#0043ce"
	ColorBlueBase  = "#0f62fe"
	ColorRed60     = "#da1e28"
	ColorRedStrong = "#ff0000"
	ColorOrange40  = "#ff832b"
	ColorGray60    = "#8d8d8d"
	ColorGray10    = "#f4f4f4"
	ColorGray90    = "#262626"
	ColorGreen40   = "#42be65"
)

//---------------------
// Types - Logger Styling
//---------------------

// Styles represents the styles used in the logger for various log levels, keys, and values.
type Styles struct {
	Out               io.Writer                 // Reserved output writer
	Timestamp         lipgloss.Style            // Timestamp style
	Levels            map[string]lipgloss.Style // Level-specific styles
	Keys              map[string]lipgloss.Style // Field key styles
	Values            map[string]lipgloss.Style // Field value styles
	DefaultKeyStyle   lipgloss.Style            // Fallback for unknown keys
	DefaultValueStyle lipgloss.Style            // Fallback for unknown values
}

//---------------------
// Default Styles (Dark Theme)
//---------------------

// DefaultStylesDark returns a set of default styles for dark-themed logging.
func DefaultStylesDark() *Styles {
	return &Styles{
		// Timestamp style: gray color and fixed width
		Timestamp: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorGray60)).
			Width(16),

		// Default styles for keys and values
		DefaultKeyStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorBlue40)),

		DefaultValueStyle: lipgloss.NewStyle(),

		// Level-specific styles for log levels
		Levels: map[string]lipgloss.Style{
			"info":  lipgloss.NewStyle().Foreground(lipgloss.Color(ColorGreen40)),
			"debug": lipgloss.NewStyle().Foreground(lipgloss.Color(ColorGray60)),
			"warn":  lipgloss.NewStyle().Foreground(lipgloss.Color(ColorOrange40)),
			"error": lipgloss.NewStyle().Foreground(lipgloss.Color(ColorRed60)),
			"panic": lipgloss.NewStyle().Foreground(lipgloss.Color(ColorRedStrong)),
		},

		// Key-specific styles for common fields in logs
		Keys: map[string]lipgloss.Style{
			"user":   lipgloss.NewStyle().Foreground(lipgloss.Color(ColorBlue40)),
			"file":   lipgloss.NewStyle().Foreground(lipgloss.Color(ColorBlue40)),
			"ip":     lipgloss.NewStyle().Foreground(lipgloss.Color(ColorBlue40)),
			"step":   lipgloss.NewStyle().Foreground(lipgloss.Color(ColorBlue40)),
			"module": lipgloss.NewStyle().Foreground(lipgloss.Color(ColorBlue40)),
			"err":    lipgloss.NewStyle().Foreground(lipgloss.Color(ColorRed60)),
		},

		// Value-specific styles for formatting field values
		Values: map[string]lipgloss.Style{
			"user":   lipgloss.NewStyle().Italic(true),
			"file":   lipgloss.NewStyle().Italic(true),
			"step":   lipgloss.NewStyle().Bold(true),
			"err":    lipgloss.NewStyle().Bold(true),
			"module": lipgloss.NewStyle(),
		},
	}
}
