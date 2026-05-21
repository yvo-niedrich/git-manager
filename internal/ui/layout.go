package ui

// Widths returns [branches, commits, detail] column widths for a given terminal
// width and focused panel index (0=branches, 1=commits, 2=detail).
// The focused panel gets extra width redistributed from the other two.
func Widths(termW, focus int) [3]int {
	base := [3]float64{0.22, 0.35, 0.43}
	boost := [3]float64{0.06, 0.12, 0.10}

	if focus < 0 || focus > 2 {
		focus = 1
	}

	// Add boost to focused; redistribute by distance so the farther column
	// yields proportionally more than the adjacent one (1:2 for edge panels,
	// 1:1 for the centre panel where both neighbours are equidistant).
	adjusted := base
	adjusted[focus] += boost[focus]
	excess := boost[focus]
	totalDist := 0
	for i := 0; i < 3; i++ {
		if i != focus {
			d := i - focus
			if d < 0 {
				d = -d
			}
			totalDist += d
		}
	}
	for i := 0; i < 3; i++ {
		if i == focus {
			continue
		}
		d := i - focus
		if d < 0 {
			d = -d
		}
		adjusted[i] -= excess * float64(d) / float64(totalDist)
	}

	var result [3]int
	remaining := termW
	for i := 0; i < 2; i++ {
		result[i] = int(float64(termW) * adjusted[i])
		remaining -= result[i]
	}
	result[2] = remaining
	return result
}

// InnerWidth returns usable inner width after lipgloss panel border+padding (2 border + 2 padding).
func InnerWidth(w int) int {
	if w > 4 {
		return w - 4
	}
	return 1
}
