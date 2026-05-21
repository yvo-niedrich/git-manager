package ui

import (
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"
)

// PlaceOverlay draws fg on top of bg with its top-left corner at (x, y)
// in visible (ANSI-aware) character coordinates.
func PlaceOverlay(x, y int, fg, bg string) string {
	fgLines := strings.Split(fg, "\n")
	bgLines := strings.Split(bg, "\n")

	out := make([]string, len(bgLines))
	for i, bgLine := range bgLines {
		fgIdx := i - y
		if fgIdx < 0 || fgIdx >= len(fgLines) {
			out[i] = bgLine
			continue
		}
		out[i] = overlayLine(x, fgLines[fgIdx], bgLine)
	}
	return strings.Join(out, "\n")
}

func overlayLine(x int, fg, bg string) string {
	bgW := lipgloss.Width(bg)
	fgW := lipgloss.Width(fg)

	leftContent := ansiTruncate(bg, x)
	left := leftContent + "\x1b[0m"
	if pad := x - lipgloss.Width(leftContent); pad > 0 {
		left += strings.Repeat(" ", pad)
	}

	right := ""
	if end := x + fgW; end < bgW {
		right = ansiSkip(bg, end)
	}

	return left + fg + right
}

// ansiTruncate returns s truncated to `width` visible characters,
// preserving CSI escape sequences.
func ansiTruncate(s string, width int) string {
	if width <= 0 {
		return ""
	}
	var out strings.Builder
	vis := 0
	i := 0
	for i < len(s) {
		if s[i] == '\x1b' && i+1 < len(s) && s[i+1] == '[' {
			j := i + 2
			for j < len(s) && (s[j] < 0x40 || s[j] > 0x7e) {
				j++
			}
			if j < len(s) {
				j++
			}
			out.WriteString(s[i:j])
			i = j
			continue
		}
		if vis >= width {
			break
		}
		r, size := utf8.DecodeRuneInString(s[i:])
		out.WriteRune(r)
		vis++
		i += size
	}
	return out.String()
}

// ansiSkip skips `skip` visible characters from s, re-emitting any CSI
// sequences encountered so the colour state is restored at the cut point.
func ansiSkip(s string, skip int) string {
	var carry strings.Builder
	vis := 0
	i := 0
	for i < len(s) && vis < skip {
		if s[i] == '\x1b' && i+1 < len(s) && s[i+1] == '[' {
			j := i + 2
			for j < len(s) && (s[j] < 0x40 || s[j] > 0x7e) {
				j++
			}
			if j < len(s) {
				j++
			}
			carry.WriteString(s[i:j])
			i = j
			continue
		}
		_, size := utf8.DecodeRuneInString(s[i:])
		vis++
		i += size
	}
	return carry.String() + s[i:]
}
