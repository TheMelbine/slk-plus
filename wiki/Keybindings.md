# Keybindings

| Key | Mode | Action |
|---|---|---|
| `j` / `k` | Normal | Move down/up in channel list or messages |
| `h` / `l` | Normal | Switch focus between panels |
| `Tab` / `Shift+Tab` | Normal | Cycle focus |
| `Enter` | Normal (sidebar) | Open selected channel, or toggle a section header |
| `Space` | Normal (sidebar) | Toggle the selected section header (collapse/expand). The state is saved per workspace and survives restarts; Slack sections are keyed by ID, so renaming one in Slack keeps it, while a `[sections.*]` section renamed in the config falls back to its default. Jumping into a collapsed section (`Ctrl+T`, next unread, a notification click) opens it only until the next start |
| `Enter` | Normal (message) | Open thread |
| `i` | Normal | Enter insert mode |
| `Esc` | Insert / Command | Return to normal mode |
| `Enter` | Insert | Send message |
| `Shift+Enter` | Insert | Newline |
| `/name args` then `Enter` | Insert (start of message) | Run a slash command instead of posting it: `/open`, `/join`, `/msg`, `/away` and `/status` run inside slk, anything else goes to Slack. If a command fails, its text comes back to the composer; start with a space to send `/...` as text. Responses only you see are marked "only visible to you"; `b` on one with buttons responds to them |
| `Ctrl+V` | Insert | Smart paste — image / file path / text (use `Ctrl+V`, not the terminal's `Ctrl+Shift+V`) |
| `Ctrl+U` | Insert | Clear compose (text + pending attachments) |
| `Ctrl+E` | Insert | Edit the draft in `$VISUAL` / `$EDITOR` (or `compose.editor`) |
| `Ctrl+O` | Insert (thread) | Toggle "also send to channel" for the next thread reply |
| `Alt+Enter` | Insert (thread) | Send thread reply and broadcast to channel (one-shot) |
| `Ctrl+U` / `Ctrl+D` | Normal | Half-page up / down |
| `Up` | Insert | Previous line; on the first line, jump to start of message |
| `Down` | Insert | Next line; on the last line, jump to end of message |
| `gg` / `G` | Normal | Jump to top / bottom |
| `/` | Normal | Search in channel (vim-style; searches cached history of the current channel) |
| `n` / `N` | Normal | Next / previous search match (wraps) |
| `a` / `A` | Normal | Jump to next / previous unread channel (wraps). Muted channels and channels set to mentions only or nothing are skipped |
| `Ctrl+h` / `Ctrl+l` | Normal | Navigate back / forward through visited channels, returning to the message you were reading when you left (`Ctrl+k` also works) |
| `Esc` | Normal (search active) | Clear active search |
| `M{a-z}` | Normal (message) | Set a session mark on the selected message — lost when slk exits |
| `M{A-Z}` | Normal (message) | Set a persistent mark — survives a restart |
| `'{letter}` / `` `{letter} `` | Normal | Jump to a mark: opens the channel, loads history if needed, opens the thread and selects the reply |
| `''` / ` `` ` | Normal | Back-jump to where you jumped from; press again to return |
| `:marks` | Normal | List every mark in the workspace with its channel and message preview |
| `:delmarks abc` | Normal | Delete marks by letter (`:delmarks a b c` works too) |
| `Backspace` | Marks overlay | Delete the selected mark |
| `Ctrl+f` | Any | Search workspace (Slack server-side; supports modifiers like `from:@user`, `in:#channel`, `before:YYYY-MM-DD`) |
| `Ctrl+b` | Any | Toggle sidebar |
| `Ctrl+]` | Any | Toggle thread panel |
| `Ctrl+a` | Normal | Toggle the Activity view (mentions, thread replies, reactions, DMs) |
| `:activity` | Normal | Open the Activity view |
| `u` | Normal (activity) | Toggle the Activity view's unread-only filter |
| `U` | Normal (activity) | Toggle the selected Activity item read/unread |
| `Ctrl+t` / `Ctrl+p` | Any | Fuzzy channel finder |
| `:ws` | Normal | Workspace picker |
| `1`–`9` | Normal | Jump to workspace N |
| `r` | Normal (message) | Open reaction picker |
| `R` | Normal (message) | Quick-toggle existing reactions |
| `E` | Normal (message) | Edit your own message |
| `D` | Normal (message) | Delete your own message (with confirmation) |
| `U` | Normal (message) | Mark selected message and everything newer as unread |
| `s` | Normal | Star or unstar a channel: the one under the sidebar cursor, or the open one |
| `m` | Normal | Cycle a channel's Slack notification level: all new messages → mentions only → nothing (same channel rule as `s`) |
| `f` | Normal (message) | Follow or unfollow the thread of the selected message (Slack's "Get notified about new replies") |
| `S` | Normal (thread) | Save thread to markdown file (`~/.local/share/slk/exports/` or `$XDG_DATA_HOME/slk/exports/`) |
| `y` | Normal (message) | Copy message text |
| `Y` / `C` | Normal (message) | Copy message permalink |
| `F` | Normal (message / thread) | Forward selected message using the channel finder; Enter sends, Esc cancels |
| `O` / `v` | Normal (message) | Open full-screen image preview |
| `Esc` / `q` | Preview | Close preview |
| `Enter` | Preview | Open in system image viewer |
| `h` / `←` | Preview | Previous image (when message has multiple) |
| `l` / `→` | Preview | Next image (when message has multiple) |
| Click | Any (on image) | Open full-screen preview |
| `Ctrl+y` | Any | Switch theme |
| `Alt+y` / `Alt+Shift+y` | Normal | Next / previous theme (wraps; saved per workspace) |
| `Ctrl+s` | Any | Set status (Active / Away / DND snooze) |
| `q` | Normal | Quit (with confirmation) |
| `Q` | Normal | Quit immediately |
| `Ctrl+c` | Any | Quit (with confirmation) |
| `:q` | Normal | Close the focused window; in the last one, quit without confirmation. `:q!` / `:qa` always quit. On a Russian layout `:й` works too |

Custom keybinding overrides are on the roadmap — see [[Tradeoffs and Non-Goals|Tradeoffs-and-Non-Goals]].
