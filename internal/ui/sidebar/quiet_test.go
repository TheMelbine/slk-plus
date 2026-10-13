package sidebar

import (
	"strings"
	"testing"

	"github.com/gammons/slk/internal/cache"
)

// rowWith renders the sidebar and returns the line holding name.
func rowWith(t *testing.T, m *Model, name string) string {
	t.Helper()
	m.ToggleCollapse("Channels")
	view := m.View(10, 30)
	for _, l := range strings.Split(view, "\n") {
		if strings.Contains(l, name) {
			return l
		}
	}
	t.Fatalf("%s row not rendered:\n%s", name, view)
	return ""
}

// A channel set to "mentions only" or "nothing" in Slack shows unread
// traffic without the dot and without bold, but brighter than a read row.
func TestQuietChannel_UnreadHasNoDotOrBold(t *testing.T) {
	unread := New([]ChannelItem{{ID: "C1", Name: "quiet", Type: "channel", IsQuiet: true}})
	unread.SetReadStateReader(func() map[string]cache.ReadState {
		return map[string]cache.ReadState{"C1": {HasUnread: true}}
	})
	read := New([]ChannelItem{{ID: "C1", Name: "quiet", Type: "channel", IsQuiet: true}})
	read.SetReadStateReader(func() map[string]cache.ReadState { return nil })

	line := rowWith(t, &unread, "quiet")
	if strings.Contains(line, "●") {
		t.Errorf("quiet channel rendered an unread dot:\n%q", line)
	}
	if strings.Contains(line, "\x1b[1m") || strings.Contains(line, "\x1b[1;") {
		t.Errorf("quiet unread channel rendered bold:\n%q", line)
	}
	if line == rowWith(t, &read, "quiet") {
		t.Errorf("quiet unread row looks the same as a read row:\n%q", line)
	}
}

// Mentions still badge in a quiet channel: that is the point of
// "mentions only".
func TestQuietChannel_MentionStillBadges(t *testing.T) {
	m := New([]ChannelItem{{ID: "C1", Name: "quiet", Type: "channel", IsQuiet: true}})
	m.SetReadStateReader(func() map[string]cache.ReadState {
		return map[string]cache.ReadState{"C1": {HasUnread: true, MentionCount: 2}}
	})
	if line := rowWith(t, &m, "quiet"); !strings.Contains(line, "2") {
		t.Errorf("quiet channel lost its mention badge:\n%q", line)
	}
}

func TestUnreadChannelCount_FiltersQuiet(t *testing.T) {
	m := New([]ChannelItem{
		{ID: "C1", Name: "a", Type: "channel"},
		{ID: "C2", Name: "b", Type: "channel", IsQuiet: true},
	})
	m.SetReadStateReader(func() map[string]cache.ReadState {
		return map[string]cache.ReadState{"C1": {HasUnread: true}, "C2": {HasUnread: true}}
	})
	if got := m.UnreadChannelCount(); got != 1 {
		t.Errorf("UnreadChannelCount = %d want 1 (C1)", got)
	}
}
