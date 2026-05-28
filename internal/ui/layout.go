package ui

// Widths returns [branches, commits, detail] column widths for a given terminal
// width and focused panel index (0=branches, 1=commits, 2=detail).
// The focused panel gets extra width taken from the furthest panel (or split
// equally between equidistant panels when commits is focused).
func Widths(termW, focus int) [3]int {
	base := [3]float64{0.22, 0.35, 0.43}
	boost := [3]float64{0.10, 0.12, 0.10}

	if focus < 0 || focus > 2 {
		focus = 1
	}

	adjusted := base
	adjusted[focus] += boost[focus]
	excess := boost[focus]

	// Find the maximum distance from the focused panel.
	maxDist := 0
	for i := 0; i < 3; i++ {
		if i == focus {
			continue
		}
		if d := abs(i - focus); d > maxDist {
			maxDist = d
		}
	}
	// Count how many panels share that maximum distance (2 for center focus).
	countAtMax := 0
	for i := 0; i < 3; i++ {
		if i != focus && abs(i-focus) == maxDist {
			countAtMax++
		}
	}
	// Deduct the boost only from the furthest panel(s).
	for i := 0; i < 3; i++ {
		if i != focus && abs(i-focus) == maxDist {
			adjusted[i] -= excess / float64(countAtMax)
		}
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

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// InnerWidth returns usable inner width after lipgloss panel border+padding (2 border + 2 padding).
func InnerWidth(w int) int {
	if w > 4 {
		return w - 4
	}
	return 1
}
