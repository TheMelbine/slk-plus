// internal/ui/mode.go
package ui

import "github.com/gammons/slk/internal/locale"

type Mode int

const (
	ModeNormal Mode = iota
	ModeInsert
	ModeCommand
	ModeSearch
	ModeChannelFinder
	ModeReactionPicker
	ModeWorkspaceFinder
	ModeThemeSwitcher
	ModePresenceMenu
	ModePresenceCustomSnooze
	ModeConfirm
	ModeHelp
	ModeNewMessage
	ModeReactionsView
	ModeLinkPicker
	ModeWorkspaceSearch
	ModeUserProfile
	ModeMarks
)

// IsModalOverlay reports whether the mode is a full-screen modal
// overlay that captures input (channel finder, pickers, help,
// confirm, etc.), as opposed to the inline modes (normal, insert,
// command, search) where the main tab stays interactive.
//
// Used by the mouse-wheel router to decide whether a wheel notch
// scrolls the items inside the modal rather than the panel under
// the cursor on the main tab behind it.
func (m Mode) IsModalOverlay() bool {
	switch m {
	case ModeChannelFinder,
		ModeReactionPicker,
		ModeWorkspaceFinder,
		ModeThemeSwitcher,
		ModePresenceMenu,
		ModePresenceCustomSnooze,
		ModeConfirm,
		ModeHelp,
		ModeNewMessage,
		ModeReactionsView,
		ModeLinkPicker,
		ModeWorkspaceSearch,
		ModeUserProfile,
		ModeMarks:
		return true
	default:
		return false
	}
}

func (m Mode) String() string {
	switch m {
	case ModeNormal:
		return locale.T("NORMAL")
	case ModeInsert:
		return locale.T("INSERT")
	case ModeCommand:
		return locale.T("COMMAND")
	case ModeSearch:
		return locale.T("SEARCH")
	case ModeChannelFinder:
		return locale.T("FIND")
	case ModeReactionPicker:
		return locale.T("REACT")
	case ModeWorkspaceFinder:
		return locale.T("WORKSPACE")
	case ModeThemeSwitcher:
		return locale.T("THEME")
	case ModePresenceMenu:
		return locale.T("STATUS")
	case ModePresenceCustomSnooze:
		return locale.T("STATUS-INPUT")
	case ModeConfirm:
		return locale.T("CONFIRM")
	case ModeHelp:
		return locale.T("HELP")
	case ModeNewMessage:
		return locale.T("NEW MSG")
	case ModeReactionsView:
		return locale.T("REACTIONS")
	case ModeLinkPicker:
		return locale.T("LINKS")
	case ModeWorkspaceSearch:
		return locale.T("WS-SEARCH")
	case ModeUserProfile:
		return locale.T("PROFILE")
	case ModeMarks:
		return locale.T("MARKS")
	default:
		return locale.T("UNKNOWN")
	}
}
