package model

import (
	"strings"
	"testing"
)

func TestCommitMessageModel_PlainEnterSubmits(t *testing.T) {
	m := NewCommitMessageDialog()
	m.input.SetValue("subject line")

	next, cmd := m.DialogUpdate(press("enter"))
	if next != nil {
		t.Error("enter: expected nil content (close signal)")
	}
	submit, ok := cmdMsg(cmd).(CommitMessageSubmitMsg)
	if !ok {
		t.Fatalf("expected CommitMessageSubmitMsg, got %T", cmdMsg(cmd))
	}
	if submit.Message != "subject line" {
		t.Errorf("Message = %q, want %q", submit.Message, "subject line")
	}
}

func TestCommitMessageModel_EnterOnEmptyStaysOpen(t *testing.T) {
	m := NewCommitMessageDialog()
	next, cmd := m.DialogUpdate(press("enter"))
	if next == nil {
		t.Error("enter with empty message should keep dialog open")
	}
	if cmd != nil {
		t.Error("enter with empty message should not submit")
	}
}

func TestCommitMessageModel_AltEnterInsertsNewlineInsteadOfSubmitting(t *testing.T) {
	m := NewCommitMessageDialog()
	m.input.SetValue("subject")

	next, cmd := m.DialogUpdate(press("alt+enter"))
	if next == nil {
		t.Fatal("alt+enter should not close the dialog")
	}
	if _, ok := cmdMsg(cmd).(CommitMessageSubmitMsg); ok {
		t.Error("alt+enter should not emit a submit message")
	}
	m = next.(*CommitMessageModel)
	if !strings.Contains(m.input.Value(), "\n") {
		t.Errorf("Value() = %q, want a newline inserted", m.input.Value())
	}
}

func TestCommitMessageModel_EscCloses(t *testing.T) {
	m := NewCommitMessageDialog()
	next, cmd := m.DialogUpdate(press("esc"))
	if next != nil {
		t.Error("esc: expected nil content (close signal)")
	}
	if cmd != nil {
		t.Error("esc: expected nil cmd")
	}
}
