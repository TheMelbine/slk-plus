// Package activityview is the UI model for the "Activity" panel: a vertical
// list of items the user was notified about — @mentions, thread replies,
// reactions to their messages, and DMs — sourced from core.ActivityItem.
//
// It mirrors internal/ui/threadsview so the App layer can wire it
// symmetrically: callers push a fresh page via SetItems and read
// SelectedItem to drive panel switching when the user activates a row.
//
// Each item renders as a two-line card: line 1 is author + activity context
// ("Mention/Thread/Reacted in #ch", or "DM") + relative time; line 2 is the
// hydrated message-body preview (empty until the messages.list hydration
// lands — see SetBodies).
package activityview

import (
	"maps"
	"slices"
	"strconv"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/gammons/slk/internal/core"
	"github.com/gammons/slk/internal/ui/messages"
	"github.com/gammons/slk/internal/ui/styles"
	"github.com/muesli/reflow/truncate"
)

// Each activity row is a two-line card (line 1: author + context + time;
// line 2: the hydrated message-body preview) followed by one blank
// separator line, except the last card. card i therefore starts at flat
// line i*cardStride. Same layout and math as threadsview.
const (
	cardContentLines = 2
	cardStride       = cardContentLines + 1
)

// Local styles, kept package-private and built from the shared color tokens
// so theme changes propagate via styles.Apply().
func mutedStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(styles.TextMuted)
}

func bodyStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(styles.TextPrimary)
}

// glyphStyle colours the leading type glyph. A read card keeps it muted
// with the rest of the card; an unread one gets the type's colour, so the
// column of glyphs shows at a glance what is new and what kind it is.
func glyphStyle(itemType string, unread bool) lipgloss.Style {
	if !unread {
		return mutedStyle()
	}
	c := styles.Primary // mentions and DMs: someone is talking to you
	switch itemType {
	case "thread_v2":
		c = styles.Accent
	case "message_reaction":
		c = styles.Warning
	}
	return lipgloss.NewStyle().Foreground(c).Bold(true)
}

// thickLeftBorder mirrors the messages/threadsview convention: a 1-column
// left border using "▌" that marks the selected row.
var thickLeftBorder = lipgloss.Border{Left: "▌"}

func borderInvisStyle() lipgloss.Style {
	return styles.InvisibleLeftBorder()
}

func borderSelectStyle(focused bool) lipgloss.Style {
	return lipgloss.NewStyle().
		BorderStyle(thickLeftBorder).BorderLeft(true).
		BorderForeground(styles.SelectionBorderColor(focused)).
		BorderBackground(styles.SelectionTintColor(focused)).
		Background(styles.SelectionTintColor(focused))
}

func borderFillStyle() lipgloss.Style {
	return lipgloss.NewStyle().Background(styles.Background)
}

// Model holds the activity-list state.
type Model struct {
	items        []core.ActivityItem
	userNames    map[string]string
	channelNames map[string]string
	channelTypes map[string]string
	selfUserID   string
	// dmUsers maps a 1:1 DM channel ID to the other person's user ID, so
	// a DM card can show their avatar.
	dmUsers map[string]string
	// avatarFn renders a user's 4x2 avatar; "" while it loads. Nil means
	// no avatar column at all.
	avatarFn messages.AvatarFunc

	// bodies holds hydrated message text/author for each item ref,
	// keyed by core.ActivityMsgKey(channelID, ts). Filled by SetBodies
	// after the messages.list second fetch; a row renders ref-only until
	// its body arrives.
	bodies map[string]core.ActivityMessage

	// unreadOnly tracks the read/unread filter flag. The view only tracks
	// it and shows an indicator; the App re-fetches with unread_only=true
	// when it flips (the server does the filtering).
	unreadOnly bool

	selected int
	yOffset  int
	focused  bool

	// snappedSelection lets View() avoid snapping yOffset back to the
	// selected row on every render (see threadsview for the rationale).
	snappedSelection int
	hasSnapped       bool

	version int64
}

// New creates an empty Model. userNames resolves actor IDs to display names;
// selfUserID renders the current user as "me".
func New(userNames map[string]string, selfUserID string) Model {
	if userNames == nil {
		userNames = map[string]string{}
	}
	return Model{
		userNames:    userNames,
		selfUserID:   selfUserID,
		channelNames: map[string]string{},
		channelTypes: map[string]string{},
		bodies:       map[string]core.ActivityMessage{},
	}
}

// SetBodies installs hydrated message bodies (keyed by
// core.ActivityMsgKey) fetched via messages.list, and forces a re-render.
// Identical contents are a no-op: every feed load re-hydrates, usually to
// the same bodies.
func (m *Model) SetBodies(bodies map[string]core.ActivityMessage) {
	if bodies == nil {
		bodies = map[string]core.ActivityMessage{}
	}
	if maps.Equal(m.bodies, bodies) {
		return
	}
	m.bodies = bodies
	m.dirty()
}

// Version returns a counter that increments any time View() output could
// change. App's panel-output cache uses this to reuse rendered frames.
func (m *Model) Version() int64 { return m.version }

func (m *Model) dirty() { m.version++ }

// SetUserNames replaces the user id -> display name map. No-op (no version
// bump) when the new map is content-equal to the current one so the
// App-level panel cache can hit on idle re-renders.
func (m *Model) SetUserNames(names map[string]string) {
	if names == nil {
		names = map[string]string{}
	}
	if stringMapsEqual(m.userNames, names) {
		return
	}
	m.userNames = names
	m.dirty()
}

// PatchUserName records a display name resolved after SetUserNames and
// bumps the version when a card shows that user (as actor, hydrated
// author, or a <@id> mention in the body). The map is normally the App's
// shared one, which already holds the name, so the bump is decided by
// the cards rather than by a map change; unrelated names, which resolve
// in bursts, cost no re-render.
func (m *Model) PatchUserName(userID, displayName string) {
	if userID == "" {
		return
	}
	if m.userNames == nil {
		m.userNames = map[string]string{}
	}
	m.userNames[userID] = displayName
	mention := "<@" + userID
	for _, it := range m.items {
		body := m.bodies[core.ActivityMsgKey(it.ChannelID, it.TS)]
		if it.AuthorID == userID || body.UserID == userID || strings.Contains(body.Text, mention) {
			m.dirty()
			return
		}
	}
}

// SetChannelNames replaces the channel id -> name map. No-op when
// content-equal to the current one.
func (m *Model) SetChannelNames(names map[string]string) {
	if names == nil {
		names = map[string]string{}
	}
	if stringMapsEqual(m.channelNames, names) {
		return
	}
	m.channelNames = names
	m.dirty()
}

// SetAvatarFunc installs the avatar renderer. Cards get an avatar column
// once it is set.
func (m *Model) SetAvatarFunc(fn messages.AvatarFunc) {
	m.avatarFn = fn
	m.dirty()
}

// SetDMUsers replaces the DM channel id -> other user id map used for the
// avatar on DM cards. No-op when content-equal to the current one.
func (m *Model) SetDMUsers(users map[string]string) {
	if users == nil {
		users = map[string]string{}
	}
	if stringMapsEqual(m.dmUsers, users) {
		return
	}
	m.dmUsers = users
	m.dirty()
}

// HandleAvatarReady re-renders the feed once a lazily fetched avatar has
// landed. Empty userID is a no-op.
func (m *Model) HandleAvatarReady(userID string) {
	if userID == "" || m.avatarFn == nil {
		return
	}
	m.dirty()
}

// SetChannelTypes replaces the channel id -> type map ("channel",
// "private", "dm", "group_dm"), which decides how a conversation is named
// on a card. No-op when content-equal to the current one.
func (m *Model) SetChannelTypes(types map[string]string) {
	if types == nil {
		types = map[string]string{}
	}
	if stringMapsEqual(m.channelTypes, types) {
		return
	}
	m.channelTypes = types
	m.dirty()
}

// stringMapsEqual reports whether two map[string]string have identical
// contents. Hand-rolled (rather than reflect.DeepEqual) because it runs on
// the render hot path.
func stringMapsEqual(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, va := range a {
		if vb, ok := b[k]; !ok || vb != va {
			return false
		}
	}
	return true
}

// SetSelfUserID updates the current user's ID. Used to render "me" labels.
func (m *Model) SetSelfUserID(id string) {
	if m.selfUserID != id {
		m.selfUserID = id
		m.dirty()
	}
}

// SetFocused marks whether the panel currently has keyboard focus.
func (m *Model) SetFocused(f bool) {
	if m.focused != f {
		m.focused = f
		m.dirty()
	}
}

// Focused reports whether the panel currently has keyboard focus.
func (m *Model) Focused() bool { return m.focused }

// UnreadOnly reports whether the unread-only filter is currently on.
func (m *Model) UnreadOnly() bool { return m.unreadOnly }

// ToggleUnreadOnly flips the unread-only filter flag and returns its new
// value. The App observes the returned value and re-fetches the feed with
// unread_only set accordingly; the view merely tracks the flag.
func (m *Model) ToggleUnreadOnly() bool {
	m.unreadOnly = !m.unreadOnly
	m.dirty()
	return m.unreadOnly
}

// SetItems replaces the list of activity items. If the previously-selected
// item (by Key) is still present, the selection follows it to its new
// position; otherwise the selection resets to the top.
//
// Identical contents are a no-op (no version bump, no re-snap): every
// activation of the view refetches the feed, usually unchanged.
func (m *Model) SetItems(items []core.ActivityItem) {
	if slices.Equal(m.items, items) {
		return
	}
	prevKey, hadSel := m.selectedKey()
	m.items = items

	newSel := 0
	if hadSel {
		for i, it := range items {
			if it.Key == prevKey {
				newSel = i
				break
			}
		}
	}
	m.selected = newSel
	m.clampSelection()
	m.hasSnapped = false // force re-snap on next render
	m.dirty()
}

// Items returns the current list of activity items.
func (m *Model) Items() []core.ActivityItem { return m.items }

// SelectedIndex returns the selection cursor's position, or 0 when empty.
func (m *Model) SelectedIndex() int { return m.selected }

// SelectedItem returns the currently selected ActivityItem (ChannelID / TS /
// ThreadTS carry enough to open the target), with ok=false when the list is
// empty.
func (m *Model) SelectedItem() (core.ActivityItem, bool) {
	if len(m.items) == 0 || m.selected < 0 || m.selected >= len(m.items) {
		return core.ActivityItem{}, false
	}
	return m.items[m.selected], true
}

// selectedKey returns the Key of the currently selected item, with ok=false
// when empty. Used by SetItems to re-anchor selection across refreshes.
func (m *Model) selectedKey() (string, bool) {
	it, ok := m.SelectedItem()
	if !ok {
		return "", false
	}
	return it.Key, true
}

// MoveDown advances the cursor by one row, clamping at the bottom.
func (m *Model) MoveDown() {
	if m.selected < len(m.items)-1 {
		m.selected++
		m.dirty()
	}
}

// MoveUp moves the cursor up by one row, clamping at zero.
func (m *Model) MoveUp() {
	if m.selected > 0 {
		m.selected--
		m.dirty()
	}
}

// GoToTop jumps to the first row.
func (m *Model) GoToTop() {
	if m.selected != 0 {
		m.selected = 0
		m.dirty()
	}
}

// GoToBottom jumps to the last row.
func (m *Model) GoToBottom() {
	if n := len(m.items); n > 0 && m.selected != n-1 {
		m.selected = n - 1
		m.dirty()
	}
}

// ScrollUp moves the viewport up n lines without changing the selection.
func (m *Model) ScrollUp(n int) {
	if n <= 0 {
		return
	}
	m.yOffset -= n
	if m.yOffset < 0 {
		m.yOffset = 0
	}
	m.snappedSelection = m.selected
	m.hasSnapped = true
	m.dirty()
}

// ScrollDown moves the viewport down n lines without changing the selection.
// View() clamps yOffset against the actual content height.
func (m *Model) ScrollDown(n int) {
	if n <= 0 {
		return
	}
	m.yOffset += n
	m.snappedSelection = m.selected
	m.hasSnapped = true
	m.dirty()
}

// ViewportAtTop reports whether the viewport is scrolled to the very top.
func (m *Model) ViewportAtTop() bool {
	return m.yOffset == 0 && len(m.items) > 0
}

// ClickAt selects the activity row whose visual row contains rowY (the
// panel-local Y inside the bordered messages-pane content area). Returns
// true when a row was selected and the caller should follow up with the
// open command; false for separator rows, the blank-fill region past the
// last row, and negative rowY. A click on either line of a card maps to
// the same item via absLine / cardStride.
func (m *Model) ClickAt(rowY int) bool {
	if rowY < 0 {
		return false
	}
	absLine := m.yOffset + rowY
	if absLine%cardStride >= cardContentLines {
		return false
	}
	idx := absLine / cardStride
	if idx < 0 || idx >= len(m.items) {
		return false
	}
	if m.selected != idx {
		m.selected = idx
		m.dirty()
	}
	return true
}

// UnreadCount returns the number of items currently flagged unread.
func (m *Model) UnreadCount() int {
	n := 0
	for _, it := range m.items {
		if it.IsUnread {
			n++
		}
	}
	return n
}

func (m *Model) clampSelection() {
	if m.selected < 0 {
		m.selected = 0
	}
	if n := len(m.items); n == 0 {
		m.selected = 0
	} else if m.selected >= n {
		m.selected = n - 1
	}
}

// View renders the activity list to `height` lines, each `width` columns
// wide. Argument order matches sidebar.View / thread.View (height first).
func (m *Model) View(height, width int) string {
	if width < 1 {
		width = 1
	}
	if height < 1 {
		height = 1
	}

	if len(m.items) == 0 {
		msg := "no activity"
		if m.unreadOnly {
			msg = "no unread activity"
		}
		empty := mutedStyle().Render(msg)
		return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, empty)
	}

	// Cards have a fixed stride, so the scroll math needs only the line
	// count; render just the cards in the viewport.
	totalLines := len(m.items)*cardStride - 1
	if !m.hasSnapped || m.snappedSelection != m.selected {
		m.snapToSelected(height, totalLines)
		m.snappedSelection = m.selected
		m.hasSnapped = true
	}
	maxOffset := totalLines - height
	if maxOffset < 0 {
		maxOffset = 0
	}
	if m.yOffset > maxOffset {
		m.yOffset = maxOffset
	}
	if m.yOffset < 0 {
		m.yOffset = 0
	}
	end := m.yOffset + height
	if end > totalLines {
		end = totalLines
	}
	visible := m.renderLines(m.yOffset, end, width)
	if pad := height - len(visible); pad > 0 {
		filler := blankLine(width)
		out := make([]string, 0, height)
		out = append(out, visible...)
		for i := 0; i < pad; i++ {
			out = append(out, filler)
		}
		visible = out
	}
	return strings.Join(visible, "\n")
}

// snapToSelected adjusts yOffset so the selected card's content lines are
// fully inside the viewport.
func (m *Model) snapToSelected(height, totalLines int) {
	start := m.selected * cardStride
	end := start + cardContentLines

	if end > m.yOffset+height {
		m.yOffset = end - height
	}
	if start < m.yOffset {
		m.yOffset = start
	}
	if m.yOffset < 0 {
		m.yOffset = 0
	}
	maxOffset := totalLines - height
	if maxOffset < 0 {
		maxOffset = 0
	}
	if m.yOffset > maxOffset {
		m.yOffset = maxOffset
	}
}

// renderLines returns lines [start, end) of the flat line list -- each
// item's two-line card, with a blank separator between adjacent cards, so
// the windowing / snap / click math stays stride-based -- rendering only
// the cards that overlap that range.
func (m *Model) renderLines(start, end, width int) []string {
	if end <= start {
		return nil
	}
	out := make([]string, 0, end-start)
	var separator string
	for i := start / cardStride; i < len(m.items) && i*cardStride < end; i++ {
		base := i * cardStride
		l1, l2 := m.renderCard(m.items[i], width, i == m.selected)
		for j, line := range [cardContentLines]string{l1, l2} {
			if abs := base + j; abs >= start && abs < end {
				out = append(out, line)
			}
		}
		if sep := base + cardContentLines; i < len(m.items)-1 && sep >= start && sep < end {
			// The gap before the first card of an older day carries its
			// date, so the feed reads in days without changing the
			// stride the windowing and click math rely on.
			if day := messages.DateFromTS(m.items[i+1].FeedTS); day != "" && day != messages.DateFromTS(m.items[i].FeedTS) {
				out = append(out, dayLine(messages.FormatDateSeparator(day), width))
				continue
			}
			if separator == "" {
				separator = blankLine(width)
			}
			out = append(out, separator)
		}
	}
	return out
}

// dayLine returns a muted, centred "── label ──" divider exactly `width`
// columns wide.
func dayLine(label string, width int) string {
	return lipgloss.NewStyle().Width(width).Align(lipgloss.Center).
		Background(styles.Background).Foreground(styles.TextMuted).
		Render("── " + label + " ──")
}

// blankLine returns an exactly `width`-column-wide empty line.
func blankLine(width int) string {
	return lipgloss.NewStyle().Width(width).Render("")
}

// renderCard returns the two lines for one activity item:
//
//	<glyph> <author> · <channel>                        <relTime>
//	  <body preview>            (reactions prefix the :emoji:)
//
// The glyph carries the item type (see activityGlyph), so line 1 has no
// "Mention in" / "Thread in" words. An unread card is drawn in the full
// text colour with a coloured glyph and a bold author; a read card is
// muted as a whole, so the eye lands on what is new. Line 2 is the
// hydrated message body (empty until the messages.list fetch lands), and
// only a mention of you keeps the mention highlight in it. Selection is
// a green left border (▌) on both lines.
func (m *Model) renderCard(it core.ActivityItem, width int, selected bool) (string, string) {
	contentWidth := width - 1
	if contentWidth < 1 {
		contentWidth = 1
	}

	body := m.bodies[core.ActivityMsgKey(it.ChannelID, it.TS)]

	// Author = the actor (reactor for reactions, message author
	// otherwise), falling back to the hydrated message's user when the
	// ref carried no author (thread_v2 / dm).
	authorID := it.AuthorID
	if authorID == "" {
		authorID = body.UserID
	}
	author := m.resolveUser(authorID)
	// A DM row is about the conversation, so it is headed by the other
	// person (the DM's name) rather than by whoever wrote last, which is
	// often you.
	if isDMItem(it.Type) {
		if name := m.channelNames[it.ChannelID]; name != "" {
			author = name
		}
	}

	// The avatar column: the actor's face, or for a DM the other
	// person's. A loading avatar keeps its blank slot so cards stay
	// aligned.
	avatarUserID := authorID
	if isDMItem(it.Type) {
		if uid := m.dmUsers[it.ChannelID]; uid != "" {
			avatarUserID = uid
		}
	}
	var avatarCol [cardContentLines]string
	if m.avatarFn != nil {
		contentWidth -= avatarColWidth
		lines := []string{}
		if avatarUserID != "" {
			if av := m.avatarFn(avatarUserID); av != "" {
				lines = strings.Split(av, "\n")
			}
		}
		for i := range avatarCol {
			if i < len(lines) {
				avatarCol[i] = lines[i] + " "
			} else {
				avatarCol[i] = strings.Repeat(" ", avatarColWidth)
			}
		}
	}

	textStyle, nameStyle, fg := mutedStyle(), mutedStyle(), messages.MutedFgANSI()
	if it.IsUnread {
		textStyle, nameStyle, fg = bodyStyle(), authorStyle(), messages.FgANSI()
	}

	left := glyphStyle(it.Type, it.IsUnread).Render(activityGlyph(it.Type)) + " "
	if author != "" {
		left += nameStyle.Render(author)
	}
	if ctx := m.contextLabel(it); ctx != "" {
		if author != "" {
			left += mutedStyle().Render(" · ")
		}
		left += ctx
	}
	right := ""
	if rel := formatRelTime(it.FeedTS); rel != "" {
		right = mutedStyle().Render(rel)
	}
	line1 := layoutLine(left, right, contentWidth)

	// Render the body the way the Threads view and message pane do, then
	// fold it onto one line: a card has exactly cardContentLines lines,
	// so a newline in the message would desync the windowing/click math.
	opts := messages.RenderSlackMarkdownOpts{UserNames: m.userNames, ChannelNames: m.channelNames, SelfUserID: m.selfUserID}
	preview := ""
	if body.Text != "" {
		preview = messages.RenderSlackMarkdownWith(body.Text, opts)
	}
	if detail := activityDetail(it); detail != "" {
		detail = messages.RenderSlackMarkdownWith(detail, opts)
		if preview != "" {
			preview = detail + "  " + preview
		} else {
			preview = detail
		}
	}
	// The renderer leaves plain text unstyled, so set the base text colour
	// here (the message pane does the same via styles.MessageText). Only
	// the foreground: a background would paint over the selection tint.
	preview = strings.ReplaceAll(preview, "\n", " ")
	if !it.IsUnread {
		// The renderer re-applies the theme text colour after each span
		// and gives bold/italic an explicit one; a read card wants all of
		// that plain text muted.
		preview = recolorSGR(preview, messages.FgANSI(), messages.MutedFgANSI())
	}
	if selected {
		// The renderer restores the theme background after each styled
		// span; on the selected card that must be the selection tint.
		preview = messages.RepaintBgToSelectionTint(preview, m.focused)
	}
	line2 := "  "
	if isDMItem(it.Type) && preview != "" && m.selfUserID != "" && body.UserID == m.selfUserID {
		// Outside textStyle: its closing reset would drop the body's
		// text colour.
		line2 += mutedStyle().Render("You: ")
	}
	line2 += textStyle.Render(preview)

	line2 = clipToWidth(line2, contentWidth)
	return m.wrapLine(avatarCol[0]+line1, contentWidth+avatarWidthOf(m.avatarFn), selected, fg),
		m.wrapLine(avatarCol[1]+line2, contentWidth+avatarWidthOf(m.avatarFn), selected, fg)
}

// avatarColWidth is the avatar (4 columns) plus a one-column gap.
const avatarColWidth = 5

func avatarWidthOf(fn messages.AvatarFunc) int {
	if fn == nil {
		return 0
	}
	return avatarColWidth
}

// recolorSGR swaps one foreground colour for another in every SGR escape
// of s, including combined ones like "\x1b[1;38;2;r;g;bm". from and to are
// full escapes as messages.FgANSI returns them.
func recolorSGR(s, from, to string) string {
	fromP, toP := sgrParams(from), sgrParams(to)
	if len(fromP) == 0 || len(toP) == 0 || from == to {
		return s
	}
	var b strings.Builder
	for {
		i := strings.Index(s, "\x1b[")
		if i < 0 {
			b.WriteString(s)
			return b.String()
		}
		j := strings.IndexByte(s[i+2:], 'm')
		if j < 0 {
			b.WriteString(s)
			return b.String()
		}
		j += i + 2
		params := strings.Split(s[i+2:j], ";")
		b.WriteString(s[:i])
		b.WriteString("\x1b[")
		b.WriteString(strings.Join(replaceRun(params, fromP, toP), ";"))
		b.WriteByte('m')
		s = s[j+1:]
	}
}

func sgrParams(esc string) []string {
	if !strings.HasPrefix(esc, "\x1b[") || !strings.HasSuffix(esc, "m") {
		return nil
	}
	return strings.Split(esc[2:len(esc)-1], ";")
}

// replaceRun replaces each run of params equal to from with to.
func replaceRun(params, from, to []string) []string {
	out := make([]string, 0, len(params))
	for i := 0; i < len(params); {
		if i+len(from) <= len(params) && slices.Equal(params[i:i+len(from)], from) {
			out = append(out, to...)
			i += len(from)
			continue
		}
		out = append(out, params[i])
		i++
	}
	return out
}

// contextLabel names the conversation for line 1, muted: "#ch" for a
// public channel, the slk glyph plus name for private channels and group
// DMs. Empty for a DM item and for anything in a 1:1 DM (the other person
// is already on the card), and for unknown item types.
func (m *Model) contextLabel(it core.ActivityItem) string {
	if contextVerb(it.Type) == "" || isDMItem(it.Type) {
		return ""
	}
	ch := m.resolveChannel(it.ChannelID)
	if ch == "" {
		return ""
	}
	switch chType := m.channelTypes[it.ChannelID]; chType {
	case "dm":
		return ""
	case "private", "group_dm":
		return mutedStyle().Render(messages.ChannelGlyph(chType) + " " + ch)
	default:
		return mutedStyle().Render("#" + ch)
	}
}

// isDMItem reports whether an activity type is a DM row, which is headed
// by the conversation rather than by the latest message's sender.
func isDMItem(itemType string) bool {
	return itemType == "dm" || itemType == "bot_dm_bundle"
}

// contextVerb maps an item type to its Activity-feed verb.
func contextVerb(itemType string) string {
	switch itemType {
	case "at_user", "at_user_group", "at_channel", "at_everyone",
		"keyword", "list_user_mentioned", "unjoined_channel_mention", "channel":
		return "Mention"
	case "thread_v2":
		return "Thread"
	case "message_reaction":
		return "Reacted"
	case "dm", "bot_dm_bundle":
		return "DM"
	default:
		return ""
	}
}

// authorStyle renders the actor name of an unread card in bold so it
// anchors the card.
func authorStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(styles.TextPrimary).Bold(true)
}

// layoutLine places `right` flush against the right edge of a width-column
// line with `left` on the left, padding between. When there's no room for
// both, the right meta is dropped and the left is clipped. Widths are
// measured with lipgloss so ANSI styling doesn't skew the math.
func layoutLine(left, right string, width int) string {
	if right == "" {
		return clipToWidth(left, width)
	}
	lw := lipgloss.Width(left)
	rw := lipgloss.Width(right)
	if lw+1+rw > width {
		return clipToWidth(left, width)
	}
	return left + strings.Repeat(" ", width-lw-rw) + right
}

// wrapLine clips a rendered line to contentWidth and applies the selection
// border + background fill (shared by both card lines so selection spans
// the whole card).
func (m *Model) wrapLine(line string, contentWidth int, selected bool, fg string) string {
	line = clipToWidth(line, contentWidth)
	// Each styled span ends in a reset that drops the row's background
	// and text colour; restore both so the row is filled edge to edge and
	// text after a span keeps the card's colour.
	bg := messages.BgANSI()
	if selected {
		bg = messages.SelectionTintBgANSI(m.focused)
	}
	line = messages.ReapplyBgAfterResets(line, bg+fg)
	borderStyle := borderInvisStyle()
	fill := borderFillStyle().Width(contentWidth)
	if selected {
		borderStyle = borderSelectStyle(m.focused)
		fill = lipgloss.NewStyle().
			Background(styles.SelectionTintColor(m.focused)).
			Width(contentWidth)
	}
	return borderStyle.Render(fill.Render(line))
}

// activityGlyph returns the leading glyph for a row keyed on the item type:
// "@" for mentions, the sidebar's thread flag for thread replies, a star
// for reactions, and an arrow for DMs (not a dot: the dot means presence
// and unread elsewhere). Each is one column wide
// and has no colour-emoji form, so terminals agree on its width. Pure (no
// styling) so it can be unit-tested.
func activityGlyph(itemType string) string {
	switch itemType {
	case "at_user", "at_user_group", "at_channel", "at_everyone",
		"keyword", "list_user_mentioned", "unjoined_channel_mention", "channel":
		return "@"
	case "thread_v2":
		return "⚑"
	case "message_reaction":
		return "✦"
	case "dm", "bot_dm_bundle":
		return "→"
	default:
		return "•"
	}
}

// activityDetail returns a short detail (currently the reaction emoji
// ":name:") that renderCard prefixes to the body preview for
// message_reaction items. Pure so it can be unit-tested.
func activityDetail(it core.ActivityItem) string {
	if it.Type == "message_reaction" && it.Reaction != "" {
		return ":" + it.Reaction + ":"
	}
	return ""
}

// resolveChannel maps a channel ID to its cached name (falling back to the
// raw ID when uncached — acceptable for MVP; a mention can reference a
// channel slk hasn't cached).
func (m *Model) resolveChannel(id string) string {
	if id == "" {
		return ""
	}
	if name, ok := m.channelNames[id]; ok && name != "" {
		return name
	}
	return id
}

// resolveUser maps a Slack user ID to a display label: "me" for the current
// user, the cached display name when known, and the raw ID otherwise.
func (m *Model) resolveUser(uid string) string {
	if uid == "" {
		return ""
	}
	if uid == m.selfUserID {
		return "me"
	}
	if name, ok := m.userNames[uid]; ok && name != "" {
		return name
	}
	return uid
}

// formatRelTime parses a Slack-style "1700000000.000000" timestamp and
// returns a coarse "Nm" / "Nh" / "Nd" string. Empty /
// unparseable inputs return "".
func formatRelTime(ts string) string {
	if ts == "" {
		return ""
	}
	secStr := ts
	if dot := strings.IndexByte(ts, '.'); dot >= 0 {
		secStr = ts[:dot]
	}
	sec, err := strconv.ParseInt(secStr, 10, 64)
	if err != nil {
		return ""
	}
	d := time.Since(time.Unix(sec, 0))
	switch {
	case d < time.Minute:
		return "now"
	case d < time.Hour:
		return strconv.Itoa(int(d/time.Minute)) + "m"
	case d < 24*time.Hour:
		return strconv.Itoa(int(d/time.Hour)) + "h"
	default:
		return strconv.Itoa(int(d/(24*time.Hour))) + "d"
	}
}

// clipToWidth truncates an already-styled string to at most `width` display
// columns (trailing ellipsis) without padding short strings.
func clipToWidth(s string, width int) string {
	if width <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= width {
		return s
	}
	return truncate.StringWithTail(s, uint(width), "…")
}
