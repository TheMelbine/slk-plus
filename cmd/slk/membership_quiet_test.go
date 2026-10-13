package main

import (
	"testing"

	"github.com/gammons/slk/internal/service"
	"github.com/gammons/slk/internal/ui/sidebar"
)

// A notification-level change from Slack must reach the sidebar items, so
// a channel switched to "mentions only" loses its unread dot right away.
func TestOnPrefChange_RefreshesQuietFlag(t *testing.T) {
	wctx := &WorkspaceContext{
		MuteStore: service.NewMuteStore(),
		Channels:  []sidebar.ChannelItem{{ID: "C1", Name: "a"}, {ID: "C2", Name: "b"}},
	}
	h := &rtmEventHandler{wsCtx: wctx, isActive: func() bool { return true }}

	h.OnPrefChange("all_notifications_prefs", `{"channels":{"C1":{"desktop":"mentions_dms"}}}`)

	if !wctx.Channels[0].IsQuiet {
		t.Errorf("C1 should be quiet after mentions_dms")
	}
	if wctx.Channels[1].IsQuiet {
		t.Errorf("C2 has no level and should not be quiet")
	}

	h.OnPrefChange("all_notifications_prefs", `{"channels":{"C1":{"desktop":"everything"}}}`)
	if wctx.Channels[0].IsQuiet {
		t.Errorf("C1 should not be quiet after everything")
	}
}
