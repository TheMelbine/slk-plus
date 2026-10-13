package demo

import (
	"sort"
	"strings"

	"github.com/gammons/slk/internal/core"
)

// activity builds the team's Activity feed from its conversations, the
// way Slack derives it: mentions of you, the latest message of each 1:1
// DM, threads you replied in, and reactions to your messages. Newest
// first.
func (w *World) activity(teamID string, unreadOnly bool) []core.ActivityItem {
	w.mu.Lock()
	defer w.mu.Unlock()
	t := w.teamByID(teamID)
	if t == nil {
		return nil
	}
	self := t.selfID
	var items []core.ActivityItem
	add := func(it core.ActivityItem) {
		it.Key = it.Type + "/" + it.ChannelID + "/" + it.TS
		if unreadOnly && !it.IsUnread {
			return
		}
		items = append(items, it)
	}
	for _, c := range t.channels {
		if c.typ == "dm" {
			if n := len(c.messages); n > 0 {
				last := c.messages[n-1]
				add(core.ActivityItem{Type: "dm", ChannelID: c.id, TS: last.TS, FeedTS: last.TS,
					IsUnread: last.TS > c.lastRead})
			}
		}
		for _, m := range c.messages {
			if m.UserID != self && strings.Contains(m.Text, "<@"+self+">") {
				add(core.ActivityItem{Type: "at_user", ChannelID: c.id, TS: m.TS, FeedTS: m.TS,
					AuthorID: m.UserID, IsUnread: m.TS > c.lastRead})
			}
			if m.UserID == self {
				for _, r := range m.Reactions {
					if by := otherThan(r.UserIDs, self); by != "" {
						add(core.ActivityItem{Type: "message_reaction", ChannelID: c.id, TS: m.TS, FeedTS: m.TS,
							AuthorID: by, Reaction: r.Emoji})
						break
					}
				}
			}
			replies := c.replies[m.TS]
			if n := len(replies); n > 0 && repliedBy(replies, self) {
				last := replies[n-1]
				add(core.ActivityItem{Type: "thread_v2", ChannelID: c.id, TS: last.TS, ThreadTS: m.TS,
					FeedTS: last.TS})
			}
		}
	}
	sort.SliceStable(items, func(i, j int) bool { return items[i].FeedTS > items[j].FeedTS })
	return items
}

// activityBodies returns the message behind each activity item, keyed by
// core.ActivityMsgKey.
func (w *World) activityBodies(items []core.ActivityItem) map[string]core.ActivityMessage {
	w.mu.Lock()
	defer w.mu.Unlock()
	out := map[string]core.ActivityMessage{}
	for _, it := range items {
		c := w.byID[it.ChannelID]
		if c == nil {
			continue
		}
		msgs := c.messages
		if it.ThreadTS != "" {
			msgs = c.replies[it.ThreadTS]
		}
		for _, m := range msgs {
			if m.TS == it.TS {
				out[core.ActivityMsgKey(it.ChannelID, it.TS)] = core.ActivityMessage{Text: m.Text, UserID: m.UserID}
				break
			}
		}
	}
	return out
}

func otherThan(ids []string, self string) string {
	for _, id := range ids {
		if id != self {
			return id
		}
	}
	return ""
}

func repliedBy(replies []core.MessageItem, self string) bool {
	for _, r := range replies {
		if r.UserID == self {
			return true
		}
	}
	return false
}
