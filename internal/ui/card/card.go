// Package card holds the rendering helpers the list panels built from
// fixed-height cards (Activity, Threads) share: the selection border and
// fill, the avatar column, the right-aligned meta on a line, date dividers,
// the short relative time, and the SGR recolouring that keeps a read card
// muted after the markdown renderer has set its own colours.
//
// Pure where possible so each piece can be unit-tested without a model.
package card

import (
	"slices"
	"strconv"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/gammons/slk/internal/ui/messages"
	"github.com/gammons/slk/internal/ui/styles"
	"github.com/muesli/reflow/truncate"
)

// AvatarColWidth is the avatar (4 columns) plus a one-column gap.
const AvatarColWidth = 5

// Muted is the style of a card's metadata and of every line of a read card.
func Muted() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(styles.TextMuted)
}

// Body is the style of an unread card's text.
func Body() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(styles.TextPrimary)
}

// Author renders the actor name of an unread card in bold so it anchors
// the card.
func Author() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(styles.TextPrimary).Bold(true)
}

// thickLeftBorder is the 1-column "▌" border that marks the selected card,
// as in the message pane.
var thickLeftBorder = lipgloss.Border{Left: "▌"}

// BorderInvis reserves the border column on an unselected card so the
// content stays aligned.
func BorderInvis() lipgloss.Style {
	return styles.InvisibleLeftBorder()
}

// BorderSelect is the selected card's border and fill: focused drives the
// border colour and the tint, the same "you're acting on this" language
// the message and thread panels use.
func BorderSelect(focused bool) lipgloss.Style {
	return lipgloss.NewStyle().
		BorderStyle(thickLeftBorder).BorderLeft(true).
		BorderForeground(styles.SelectionBorderColor(focused)).
		BorderBackground(styles.SelectionTintColor(focused)).
		Background(styles.SelectionTintColor(focused))
}

// BorderFill is the row fill of an unselected card.
func BorderFill() lipgloss.Style {
	return lipgloss.NewStyle().Background(styles.Background)
}

// AvatarWidthOf is the width the avatar column takes: zero without a
// renderer.
func AvatarWidthOf(fn messages.AvatarFunc) int {
	if fn == nil {
		return 0
	}
	return AvatarColWidth
}

// AvatarColumn returns rows lines of the avatar column for userID: the
// avatar's rows each followed by the gap, and blank AvatarColWidth-wide
// slots for the rows it has no line for (the third row of a card, or the
// whole column while the avatar loads or userID is empty), so cards stay
// aligned. Nil when fn is nil.
func AvatarColumn(fn messages.AvatarFunc, userID string, rows int) []string {
	if fn == nil {
		return nil
	}
	var lines []string
	if userID != "" {
		if av := fn(userID); av != "" {
			lines = strings.Split(av, "\n")
		}
	}
	out := make([]string, rows)
	for i := range out {
		if i < len(lines) {
			out[i] = lines[i] + " "
		} else {
			out[i] = strings.Repeat(" ", AvatarColWidth)
		}
	}
	return out
}

// WrapLine clips a rendered line to contentWidth and applies the
// selection border and background fill, restoring the row's background
// and text colour (fg, a full SGR escape as messages.FgANSI returns it)
// after every reset a styled span left behind, so the row is filled edge
// to edge and text after a span keeps the card's colour.
func WrapLine(line string, contentWidth int, selected, focused bool, fg string) string {
	line = ClipToWidth(line, contentWidth)
	bg := messages.BgANSI()
	if selected {
		bg = messages.SelectionTintBgANSI(focused)
	}
	line = messages.ReapplyBgAfterResets(line, bg+fg)
	borderStyle := BorderInvis()
	fill := BorderFill().Width(contentWidth)
	if selected {
		borderStyle = BorderSelect(focused)
		fill = lipgloss.NewStyle().
			Background(styles.SelectionTintColor(focused)).
			Width(contentWidth)
	}
	return borderStyle.Render(fill.Render(line))
}

// LayoutLine places right flush against the right edge of a width-column
// line with left on the left, padding between. When there's no room for
// both, right is dropped and left is clipped. Widths are measured with
// lipgloss so ANSI styling doesn't skew the math.
func LayoutLine(left, right string, width int) string {
	if right == "" {
		return ClipToWidth(left, width)
	}
	lw := lipgloss.Width(left)
	rw := lipgloss.Width(right)
	if lw+1+rw > width {
		return ClipToWidth(left, width)
	}
	return left + strings.Repeat(" ", width-lw-rw) + right
}

// DayLine returns a muted, centred "── label ──" divider exactly width
// columns wide.
func DayLine(label string, width int) string {
	return lipgloss.NewStyle().Width(width).Align(lipgloss.Center).
		Background(styles.Background).Foreground(styles.TextMuted).
		Render("── " + label + " ──")
}

// BlankLine returns an exactly width-column-wide empty line.
func BlankLine(width int) string {
	return lipgloss.NewStyle().Width(width).Render("")
}

// ClipToWidth truncates an already-styled string to at most width display
// columns (with a trailing ellipsis) but does not pad short strings:
// padding is left to the caller's width-bounded style so the background
// fills correctly.
func ClipToWidth(s string, width int) string {
	if width <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= width {
		return s
	}
	return truncate.StringWithTail(s, uint(width), "…")
}

// RelTime parses a Slack-style "1700000000.000000" timestamp and returns
// a coarse "now" / "Nm" / "Nh" / "Nd" string. Empty or unparseable input
// returns "".
func RelTime(ts string) string {
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

// RecolorSGR swaps one foreground colour for another in every SGR escape
// of s, including combined ones like "\x1b[1;38;2;r;g;bm". from and to
// are full escapes as messages.FgANSI returns them.
func RecolorSGR(s, from, to string) string {
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
