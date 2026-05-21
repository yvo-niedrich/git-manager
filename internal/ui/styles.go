package ui

import "github.com/charmbracelet/lipgloss"

var (
	// Foreground / semantic colours
	ColorAccent    = lipgloss.Color("#7C3AED") // violet — focused borders, key hints
	ColorDim       = lipgloss.Color("#4B5563") // inactive borders
	ColorSuccess   = lipgloss.Color("#10B981")
	ColorError     = lipgloss.Color("#EF4444")
	ColorWarn      = lipgloss.Color("#F59E0B")
	ColorSubtle    = lipgloss.Color("#6B7280") // section headers, hint descriptions
	ColorText      = lipgloss.Color("#F9FAFB") // primary text
	ColorTextMuted = lipgloss.Color("#94A3B8") // dimmed text on highlighted rows
	ColorHead      = lipgloss.Color("#F59E0B") // HEAD commit / current-branch marker
	ColorRemote    = lipgloss.Color("#06B6D4") // remote branch names

	// Background colours
	ColorSelected    = lipgloss.Color("#1D4ED8") // single-cursor highlight
	ColorMultiSelect = lipgloss.Color("#1E3A5F") // multi-select highlight
	ColorBarBg       = lipgloss.Color("#111827") // status-bar background
	ColorMenuBg      = lipgloss.Color("#1F2937") // context-menu background
	ColorBackdrop    = lipgloss.Color("#10141b") // context-menu backdrop
)

func PanelStyle(focused bool) lipgloss.Style {
	border := lipgloss.RoundedBorder()
	s := lipgloss.NewStyle().
		Border(border).
		BorderForeground(ColorDim).
		Padding(0, 1)
	if focused {
		s = s.BorderForeground(ColorAccent)
	}
	return s
}

func TitleStyle(focused bool) lipgloss.Style {
	s := lipgloss.NewStyle().Bold(true).Foreground(ColorDim)
	if focused {
		s = s.Foreground(ColorAccent)
	}
	return s
}

var (
	SelectedItemStyle    = lipgloss.NewStyle().Background(ColorSelected).Foreground(ColorText)
	SelectedItemDimStyle = lipgloss.NewStyle().Background(ColorSelected).Foreground(ColorTextMuted)
	NormalItemStyle      = lipgloss.NewStyle().Foreground(ColorText)
	DimItemStyle         = lipgloss.NewStyle().Foreground(ColorSubtle)
	HeadStyle            = lipgloss.NewStyle().Foreground(ColorHead).Bold(true)
	TagStyle             = lipgloss.NewStyle().Foreground(ColorWarn).Bold(true)
	RemoteStyle          = lipgloss.NewStyle().Foreground(ColorRemote)
	SectionStyle         = lipgloss.NewStyle().Foreground(ColorSubtle).Bold(true)

	StatusOKStyle  = lipgloss.NewStyle().Foreground(ColorSuccess)
	StatusErrStyle = lipgloss.NewStyle().Foreground(ColorError)
	StatusBarStyle = lipgloss.NewStyle().Background(ColorBarBg).Foreground(ColorSubtle)

	KeyHintStyle  = lipgloss.NewStyle().Foreground(ColorAccent).Bold(true)
	DescHintStyle = lipgloss.NewStyle().Foreground(ColorSubtle)

	MenuBorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorAccent).
			BorderBackground(ColorMenuBg).
			Background(ColorMenuBg).
			Padding(0, 2)
)
