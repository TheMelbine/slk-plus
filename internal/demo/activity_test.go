package demo

import (
	"testing"

	"github.com/gammons/slk/internal/core"
)

func TestActivity_FeedFromFixtures(t *testing.T) {
	w := fixtureWorld()
	items := w.activity(teamLumen, false)
	seen := map[string]bool{}
	for i, it := range items {
		seen[it.Type] = true
		if i > 0 && items[i-1].FeedTS < it.FeedTS {
			t.Errorf("feed not newest first at %d", i)
		}
	}
	for _, typ := range []string{"at_user", "dm", "thread_v2", "message_reaction"} {
		if !seen[typ] {
			t.Errorf("feed has no %s item: %+v", typ, items)
		}
	}
	bodies := w.activityBodies(items)
	for _, it := range items {
		if bodies[core.ActivityMsgKey(it.ChannelID, it.TS)].Text == "" {
			t.Errorf("no body for %s %s/%s", it.Type, it.ChannelID, it.TS)
		}
	}
	for _, it := range w.activity(teamLumen, true) {
		if !it.IsUnread {
			t.Errorf("unread-only feed has a read item: %+v", it)
		}
	}
}
