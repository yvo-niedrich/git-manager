package model

import "testing"

// ── buildFileTree ─────────────────────────────────────────────────────────────

func TestBuildFileTree_NestedPaths(t *testing.T) {
	lines := buildFileTree([]string{"internal/git/client.go", "internal/model/app.go", "README.md"})
	want := []string{
		"├── internal/",
		"│   ├── git/",
		"│   │   └── client.go",
		"│   └── model/",
		"│       └── app.go",
		"└── README.md",
	}
	if len(lines) != len(want) {
		t.Fatalf("got %d lines, want %d\n got: %v\nwant: %v", len(lines), len(want), lines, want)
	}
	for i := range want {
		if lines[i] != want[i] {
			t.Errorf("line %d = %q, want %q", i, lines[i], want[i])
		}
	}
}

// ── CommitPreviewModel ────────────────────────────────────────────────────────

func TestCommitPreviewModel_ProceedPreselected(t *testing.T) {
	m := NewCommitPreviewDialog("title", []string{"a.txt"}, 40)
	if !m.proceed {
		t.Error("proceed should be pre-selected")
	}
}

func TestCommitPreviewModel_LeftSelectsCancel(t *testing.T) {
	m := NewCommitPreviewDialog("title", []string{"a.txt"}, 40)
	next, _ := m.DialogUpdate(press("left"))
	m = next.(*CommitPreviewModel)
	if m.proceed {
		t.Error("left should select cancel")
	}

	next, cmd := m.DialogUpdate(press("enter"))
	if next != nil {
		t.Error("enter on cancel should close the dialog (nil content)")
	}
	if cmd != nil {
		t.Error("enter on cancel should not emit a submit message")
	}
}

func TestCommitPreviewModel_EnterOnProceedEmitsSubmit(t *testing.T) {
	m := NewCommitPreviewDialog("title", []string{"a.txt"}, 40)
	next, cmd := m.DialogUpdate(press("enter"))
	if next != nil {
		t.Error("enter: expected nil content (close signal)")
	}
	if _, ok := cmdMsg(cmd).(CommitPreviewSubmitMsg); !ok {
		t.Fatalf("expected CommitPreviewSubmitMsg, got %T", cmdMsg(cmd))
	}
}

func TestCommitPreviewModel_RightReselectsProceed(t *testing.T) {
	m := NewCommitPreviewDialog("title", []string{"a.txt"}, 40)
	next, _ := m.DialogUpdate(press("left"))
	m = next.(*CommitPreviewModel)
	next, _ = m.DialogUpdate(press("right"))
	m = next.(*CommitPreviewModel)
	if !m.proceed {
		t.Error("right should reselect proceed")
	}
}

func TestCommitPreviewModel_EscCloses(t *testing.T) {
	m := NewCommitPreviewDialog("title", []string{"a.txt"}, 40)
	next, cmd := m.DialogUpdate(press("esc"))
	if next != nil {
		t.Error("esc: expected nil content (close signal)")
	}
	if cmd != nil {
		t.Error("esc: expected nil cmd")
	}
}

func TestCommitPreviewModel_ScrollClampsAtBounds(t *testing.T) {
	paths := make([]string, 30)
	for i := range paths {
		paths[i] = string(rune('a'+i)) + ".txt"
	}
	m := NewCommitPreviewDialog("title", paths, 40)

	next, _ := m.DialogUpdate(press("up"))
	if next.(*CommitPreviewModel).offset != 0 {
		t.Error("scrolling up at offset 0 should stay at 0")
	}

	for i := 0; i < 100; i++ {
		next, _ = m.DialogUpdate(press("down"))
		m = next.(*CommitPreviewModel)
	}
	if m.offset != m.maxOffset() {
		t.Errorf("offset = %d after over-scrolling down, want clamped to maxOffset %d", m.offset, m.maxOffset())
	}
}
