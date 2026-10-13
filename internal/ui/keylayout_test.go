package ui

import (
	"testing"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

func TestPhysicalKey(t *testing.T) {
	keys := DefaultKeyMap()
	cases := []struct {
		name     string
		in       tea.KeyPressMsg
		want     key.Binding
		nonLatin bool
	}{
		{"table: о is j", tea.KeyPressMsg{Code: 'о', Text: "о"}, keys.Down, false},
		{"table: л is k", tea.KeyPressMsg{Code: 'л', Text: "л"}, keys.Up, false},
		{"table: legacy uppercase П is G", tea.KeyPressMsg{Code: 'П', Text: "П"}, keys.Bottom, false},
		{"table: shift+ж is :", tea.KeyPressMsg{Code: 'ж', Mod: tea.ModShift, Text: "Ж"}, keys.CommandMode, false},
		{"table: ctrl+е is ctrl+t", tea.KeyPressMsg{Code: 'е', Mod: tea.ModCtrl}, keys.FuzzyFinder, false},
		{"table: ctrl+ъ is ctrl+]", tea.KeyPressMsg{Code: 'ъ', Mod: tea.ModCtrl}, keys.ToggleThread, false},
		{"basecode wins", tea.KeyPressMsg{Code: 'ш', BaseCode: 'i', Text: "ш"}, keys.InsertMode, false},
		{"ru legacy: . is /", tea.KeyPressMsg{Code: '.', Text: "."}, keys.SearchMode, true},
		{"ru legacy: , is ?", tea.KeyPressMsg{Code: ',', Text: ","}, keys.Help, true},
		{"ru basecode: . is /", tea.KeyPressMsg{Code: '.', BaseCode: '/', Text: "."}, keys.SearchMode, true},
		{"ru basecode: shift+/ key is ?", tea.KeyPressMsg{Code: ',', BaseCode: '/', Mod: tea.ModShift, Text: ","}, keys.Help, true},
		{"basecode shifted", tea.KeyPressMsg{Code: 'й', BaseCode: 'q', Mod: tea.ModShift, Text: "Й"}, keys.QuitConfirm, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := physicalKey(tc.in, true, tc.nonLatin)
			if !key.Matches(got, tc.want) {
				t.Fatalf("physicalKey(%q) = %q, want a match for %v", tc.in.String(), got.String(), tc.want.Keys())
			}
		})
	}
}

func TestPhysicalKey_LeavesTextAlone(t *testing.T) {
	in := tea.KeyPressMsg{Code: 'о', BaseCode: 'j', Text: "о"}
	if got := physicalKey(in, false, true); got.String() != "о" {
		t.Fatalf("text-entry key rewritten to %q", got.String())
	}
	ascii := tea.KeyPressMsg{Code: 'j', Text: "j"}
	if got := physicalKey(ascii, true, true); got != tea.KeyMsg(ascii) {
		t.Fatalf("ASCII key changed: %#v", got)
	}
	// On a Latin layout `.` stays `.`.
	dot := tea.KeyPressMsg{Code: '.', BaseCode: '.', Text: "."}
	if got := physicalKey(dot, true, false); got.String() != "." {
		t.Fatalf("latin . rewritten to %q", got.String())
	}
	// Chords are rewritten even while typing: ctrl+г clears compose like ctrl+u.
	if got := physicalKey(tea.KeyPressMsg{Code: 'г', Mod: tea.ModCtrl}, false, true); got.String() != "ctrl+u" {
		t.Fatalf("ctrl+г = %q, want ctrl+u", got.String())
	}
}

func TestHandleKey_RussianLayoutEntersInsertMode(t *testing.T) {
	a := newTestAppWithMessages(t)
	a.SetMode(ModeNormal)
	_ = a.handleKey(tea.KeyPressMsg{Code: 'ш', Text: "ш"})
	if a.mode != ModeInsert {
		t.Fatalf("mode = %v, want insert", a.mode)
	}
	_ = a.handleKey(tea.KeyPressMsg{Code: 'ш', Text: "ш"})
	if v := a.compose.Value(); v != "ш" {
		t.Fatalf("compose = %q, want the Cyrillic letter typed as-is", v)
	}
}

func TestHandleKey_RussianDotOpensSearch(t *testing.T) {
	a := newTestAppWithMessages(t)
	a.SetMode(ModeInsert)
	_ = a.handleKey(tea.KeyPressMsg{Code: 'п', Text: "п"})
	_ = a.handleKey(tea.KeyPressMsg{Code: tea.KeyEscape})
	_ = a.handleKey(tea.KeyPressMsg{Code: '.', Text: "."})
	if a.mode != ModeSearch {
		t.Fatalf("mode = %v, want search", a.mode)
	}
}

func TestHandleKey_RussianLayoutScrollsHelp(t *testing.T) {
	a := newTestAppWithMessages(t)
	a.SetMode(ModeNormal)
	_ = a.handleKey(tea.KeyPressMsg{Code: '?', Text: "?"})
	if a.mode != ModeHelp {
		t.Fatalf("mode = %v, want help", a.mode)
	}
	_ = a.handleKey(tea.KeyPressMsg{Code: 'о', Text: "о"})
	_ = a.handleKey(tea.KeyPressMsg{Code: 'о', Text: "о"})
	_ = a.handleKey(tea.KeyPressMsg{Code: 'л', Text: "л"})
	if got := a.help.Selected(); got != 1 {
		t.Fatalf("selected = %d after о о л, want 1", got)
	}
	// Search inside the cheatsheet takes the letter as typed.
	_ = a.handleKey(tea.KeyPressMsg{Code: '/', Text: "/"})
	_ = a.handleKey(tea.KeyPressMsg{Code: 'о', Text: "о"})
	if !a.help.IsSearching() || a.help.Selected() != 0 {
		t.Fatalf("searching = %v selected = %d, want a Cyrillic query", a.help.IsSearching(), a.help.Selected())
	}
}

// The marks overlay has no text input: on a Russian layout the physical
// A key (ф) jumps to mark a, as the M chord recorded it.
func TestHandleKey_RussianLayoutJumpsFromMarksOverlay(t *testing.T) {
	app := jumpMarkTestApp(t)
	seedMark(t, app, "a", Location{TeamID: "T1", ChannelID: "C2", MessageTS: "10.0"})
	app.openMarksOverlay()
	app.SetMode(ModeMarks)

	_, cmd := app.Update(tea.KeyPressMsg{Code: 'ф', Text: "ф"})
	if cmd == nil {
		t.Fatal("ф in the marks overlay must jump to mark a")
	}
	cs, ok := cmd().(ChannelSelectedMsg)
	if !ok || cs.ID != "C2" {
		t.Fatalf("ф jumped to %T %+v, want C2", cs, cs)
	}
	if app.mode != ModeNormal || app.marksOverlay.IsVisible() {
		t.Fatal("the jump must close the overlay")
	}
}

// э is the ' key on a Russian layout: pressed twice it takes the
// back-jump, the second press reaching the open overlay.
func TestHandleKey_RussianLayoutBackJumpFromMarksOverlay(t *testing.T) {
	app := jumpMarkTestApp(t)
	seedMark(t, app, "a", Location{TeamID: "T1", ChannelID: "C2", MessageTS: "10.0"})
	app.Update(ChannelSelectedMsg{ID: "C1", Name: "one", Type: "channel"})
	app.messagepane.SelectByTS("1.0")
	driveJump(app, pressJumpChord(app, 'a'))
	if app.activeChannelID != "C2" {
		t.Fatalf("setup: jump landed in %q, want C2", app.activeChannelID)
	}

	app.Update(tea.KeyPressMsg{Code: 'э', Text: "э"})
	if app.mode != ModeMarks {
		t.Fatalf("э must open the marks overlay, mode=%v", app.mode)
	}
	_, cmd := app.Update(tea.KeyPressMsg{Code: 'э', Text: "э"})
	if app.mode != ModeNormal || app.marksOverlay.IsVisible() {
		t.Fatal("the back-jump must close the overlay")
	}
	driveJump(app, cmd)
	if app.activeChannelID != "C1" {
		t.Fatalf("back-jump landed in %q, want C1", app.activeChannelID)
	}
}
