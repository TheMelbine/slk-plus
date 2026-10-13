package threadsview

import (
	"image/color"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/gammons/slk/internal/cache"
	"github.com/gammons/slk/internal/ui/styles"
)

var sgrFgRe = regexp.MustCompile(`\x1b\[[0-9;]*?38;2;(\d+;\d+;\d+)[0-9;]*m`)

// fgOf returns the truecolor foreground lipgloss emits for c.
func fgOf(t *testing.T, c color.Color) string {
	t.Helper()
	m := sgrFgRe.FindStringSubmatch(lipgloss.NewStyle().Foreground(c).Render("x"))
	if m == nil {
		t.Fatalf("style for %v emitted no truecolor foreground", c)
	}
	return m[1]
}

// fgBefore returns the last truecolor foreground set before word in line.
func fgBefore(t *testing.T, line, word string) string {
	t.Helper()
	i := strings.Index(line, word)
	if i < 0 {
		t.Fatalf("%q not found in %q", word, ansi.Strip(line))
	}
	all := sgrFgRe.FindAllStringSubmatch(line[:i], -1)
	if len(all) == 0 {
		return ""
	}
	return all[len(all)-1][1]
}

func summary(unread bool) cache.ThreadSummary {
	return cache.ThreadSummary{
		ChannelID: "C1", ChannelName: "general", ChannelType: "channel",
		ThreadTS: "1.000000", ParentUserID: "U1", ParentText: "hello *world* <@U2> <@UME>",
		ParentTS: "1.000000", ReplyCount: 3, LastReplyTS: "5.000000", LastReplyBy: "U2",
		Unread: unread,
	}
}

// Line 1 is the flag, the author and the channel with the time of the
// last reply on the right; line 3 is the reply count. An unread card has
// a coloured flag and a bold author in the text colour.
func TestRenderCard_UnreadLayout(t *testing.T) {
	m := New(map[string]string{"U1": "alice", "U2": "bob", "UME": "Dan"}, "UME")
	s := summary(true)
	s.LastReplyTS = strconv.FormatInt(time.Now().Add(-7*time.Minute).Unix(), 10) + ".0"
	lines := m.renderCard(s, 60, false)
	if len(lines) != cardContentLines {
		t.Fatalf("card has %d lines, want %d", len(lines), cardContentLines)
	}
	l1 := ansi.Strip(lines[0])
	if !strings.HasPrefix(l1, "▌⚑ alice · #general") || !strings.HasSuffix(l1, "7m") {
		t.Errorf("line 1 = %q, want '⚑ alice · #general … 7m'", l1)
	}
	if l3 := ansi.Strip(lines[2]); !strings.Contains(l3, "3 replies · last by bob") {
		t.Errorf("line 3 = %q", l3)
	}
	if got, want := fgBefore(t, lines[0], "⚑"), fgOf(t, styles.Accent); got != want {
		t.Errorf("unread flag colour = %s, want Accent %s", got, want)
	}
	if got, want := fgBefore(t, lines[0], "alice"), fgOf(t, styles.TextPrimary); got != want {
		t.Errorf("unread author colour = %s, want text %s", got, want)
	}
	if !strings.Contains(lines[0][:strings.Index(lines[0], "alice")], "\x1b[1;") {
		t.Errorf("unread author is not bold: %q", lines[0])
	}
	for _, line := range lines {
		if w := lipgloss.Width(line); w != 60 {
			t.Errorf("line width = %d, want 60: %q", w, ansi.Strip(line))
		}
	}
}

// A read card is muted as a whole: the flag, the author and the preview,
// including bold text and text after a mention, which the renderer would
// otherwise give the full text colour.
func TestRenderCard_ReadCardMuted(t *testing.T) {
	m := New(map[string]string{"U1": "alice", "U2": "bob", "UME": "Dan"}, "UME")
	lines := m.renderCard(summary(false), 80, false)
	muted := fgOf(t, styles.TextMuted)
	for _, word := range []string{"⚑", "alice"} {
		if got := fgBefore(t, lines[0], word); got != muted {
			t.Errorf("%q colour = %s, want muted %s", word, got, muted)
		}
	}
	for _, word := range []string{"hello", "world", "@bob"} {
		if got := fgBefore(t, lines[1], word); got != muted {
			t.Errorf("%q colour = %s, want muted %s", word, got, muted)
		}
	}
}

// In the preview only a mention of you keeps the highlight; other
// people's mentions read as plain text.
func TestRenderCard_OnlySelfMentionHighlighted(t *testing.T) {
	m := New(map[string]string{"U1": "alice", "U2": "bob", "UME": "Dan"}, "UME")
	lines := m.renderCard(summary(true), 80, false)
	if got, want := fgBefore(t, lines[1], "@bob"), fgOf(t, styles.TextPrimary); got != want {
		t.Errorf("@bob colour = %s, want plain text %s", got, want)
	}
	if got, want := fgBefore(t, lines[1], "@Dan"), fgOf(t, styles.Primary); got != want {
		t.Errorf("@Dan colour = %s, want the mention accent %s", got, want)
	}
}

// With an avatar renderer each card leads with the parent author's
// avatar; a still-loading avatar keeps its blank slot so cards stay
// aligned, and the footer row under the avatar is blank too.
func TestRenderCard_AvatarColumn(t *testing.T) {
	m := New(map[string]string{"U1": "alice"}, "UME")
	m.SetAvatarFunc(func(uid string) string {
		if uid == "U3" {
			return ""
		}
		return strings.ToUpper(uid[:2]) + "AA\n" + "bbbb"
	})
	lines := m.renderCard(summary(true), 60, false)
	if got := ansi.Strip(lines[0]); !strings.Contains(got, "U1AA ⚑ alice") {
		t.Errorf("line 1 = %q, want the avatar before the flag", got)
	}
	if got := ansi.Strip(lines[1]); !strings.HasPrefix(got, "▌bbbb ") {
		t.Errorf("line 2 = %q, want the avatar's second row", got)
	}
	if got := ansi.Strip(lines[2]); !strings.HasPrefix(got, "▌       3 replies") {
		t.Errorf("line 3 = %q, want a blank slot under the avatar", got)
	}
	s := summary(true)
	s.ParentUserID = "U3"
	lines = m.renderCard(s, 60, false)
	if got := ansi.Strip(lines[0]); !strings.HasPrefix(got, "▌     ⚑ U3") {
		t.Errorf("loading avatar line 1 = %q, want a blank 5-column slot", got)
	}
	for _, line := range lines {
		if w := lipgloss.Width(line); w != 60 {
			t.Errorf("card line width = %d, want 60", w)
		}
	}
}

// Private channels and DMs carry the sidebar glyph instead of "#".
func TestChannelLabel(t *testing.T) {
	cases := map[string]string{"channel": "#general", "private": "◆ general", "dm": "● general", "group_dm": "● general"}
	for typ, want := range cases {
		if got := channelLabel(cache.ThreadSummary{ChannelName: "general", ChannelType: typ}); got != want {
			t.Errorf("channelLabel(%s) = %q, want %q", typ, got, want)
		}
	}
}

// The gap before the first card of an older day is a date divider.
func TestView_DayDividerBetweenDays(t *testing.T) {
	now := time.Now()
	ts := func(t time.Time) string { return strconv.FormatInt(t.Unix(), 10) + ".0" }
	m := New(nil, "")
	m.SetSummaries([]cache.ThreadSummary{
		{ChannelID: "C1", ThreadTS: "1.1", ParentText: "a", LastReplyTS: ts(now)},
		{ChannelID: "C1", ThreadTS: "1.2", ParentText: "b", LastReplyTS: ts(now)},
		{ChannelID: "C1", ThreadTS: "1.3", ParentText: "c", LastReplyTS: ts(now.AddDate(0, 0, -3))},
	})
	lines := strings.Split(ansi.Strip(m.View(20, 60)), "\n")
	if strings.Contains(lines[3], "──") {
		t.Errorf("same-day gap shows a divider: %q", lines[3])
	}
	if !strings.Contains(lines[7], "──") {
		t.Errorf("gap before an older day has no divider: %q", lines[7])
	}
}
