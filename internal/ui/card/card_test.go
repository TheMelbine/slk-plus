package card

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

func TestRelTime(t *testing.T) {
	if got := RelTime(""); got != "" {
		t.Errorf("RelTime(empty) = %q, want empty", got)
	}
	if got := RelTime("not-a-ts"); got != "" {
		t.Errorf("RelTime(garbage) = %q, want empty", got)
	}
	ts := strconv.FormatInt(time.Now().Add(-7*time.Minute).Unix(), 10) + ".0"
	if got := RelTime(ts); got != "7m" {
		t.Errorf("RelTime(7m ago) = %q, want 7m", got)
	}
}

func TestRecolorSGR(t *testing.T) {
	from, to := "\x1b[38;2;1;2;3m", "\x1b[38;2;9;9;9m"
	in := "\x1b[38;2;1;2;3ma\x1b[1;38;2;1;2;3mb\x1b[38;2;4;5;6mc\x1b[m"
	want := "\x1b[38;2;9;9;9ma\x1b[1;38;2;9;9;9mb\x1b[38;2;4;5;6mc\x1b[m"
	if got := RecolorSGR(in, from, to); got != want {
		t.Errorf("RecolorSGR = %q, want %q", got, want)
	}
}

func TestLayoutLine(t *testing.T) {
	if got := LayoutLine("ab", "7m", 10); got != "ab      7m" {
		t.Errorf("LayoutLine = %q", got)
	}
	// No room for both: the meta is dropped and the left clipped.
	if got := LayoutLine("abcdefghij", "7m", 10); got != "abcdefghij" {
		t.Errorf("LayoutLine(no room) = %q", got)
	}
}

// The avatar column keeps a blank slot while the avatar loads and for
// the rows the avatar has no line for.
func TestAvatarColumn(t *testing.T) {
	fn := func(uid string) string {
		if uid == "U3" {
			return ""
		}
		return "AAAA\nbbbb"
	}
	if got := AvatarColumn(nil, "U1", 2); got != nil {
		t.Errorf("AvatarColumn(nil fn) = %v, want nil", got)
	}
	got := AvatarColumn(fn, "U1", 3)
	want := []string{"AAAA ", "bbbb ", "     "}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Errorf("AvatarColumn = %q, want %q", got, want)
	}
	for _, line := range AvatarColumn(fn, "U3", 2) {
		if line != "     " {
			t.Errorf("loading avatar row = %q, want a blank slot", line)
		}
	}
}

func TestWrapLine_Width(t *testing.T) {
	for _, selected := range []bool{false, true} {
		line := WrapLine("hello", 20, selected, true, "")
		if w := lipgloss.Width(line); w != 21 {
			t.Errorf("WrapLine(selected=%v) width = %d, want 21 (border + 20)", selected, w)
		}
		if !strings.Contains(ansi.Strip(line), "hello") {
			t.Errorf("WrapLine lost the text: %q", ansi.Strip(line))
		}
	}
}
