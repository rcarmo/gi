package tui

// /hotkeys: Pi's handleHotkeysCommand text for Pi's default keybindings
// (golden: scripts/golden-hotkeys.mjs), then the keys gi adds.

// piHotkeysTemplate is Pi's text; piKeys.hotkeysMarkdown fills in the keys
// that depend on the platform.
const piHotkeysTemplate = "**Navigation**\n" +
	"| Key | Action |\n|-----|--------|\n" +
	"| `Up` / `Down` / `Left/Ctrl+B` / `Right/Ctrl+F` | Move cursor / browse history |\n" +
	"| `Alt+Left/Ctrl+Left/Alt+B` / `Alt+Right/Ctrl+Right/Alt+F` | Move by word |\n" +
	"| `Home/Ctrl+A` | Start of line |\n" +
	"| `End/Ctrl+E` | End of line |\n" +
	"| `Ctrl+]` | Jump forward to character |\n" +
	"| `Ctrl+Alt+]` | Jump backward to character |\n" +
	"| `PageUp/Ctrl+PageUp` / `PageDown/Ctrl+PageDown` | Scroll by page |\n" +
	"\n**Editing**\n" +
	"| Key | Action |\n|-----|--------|\n" +
	"| `Enter` | Send message |\n" +
	"| `Shift+Enter/Ctrl+J` | {newLine} |\n" +
	"| `Ctrl+W/Alt+Backspace` | Delete word backwards |\n" +
	"| `Alt+D/Alt+Delete` | Delete word forwards |\n" +
	"| `Ctrl+U` | Delete to start of line |\n" +
	"| `Ctrl+K` | Delete to end of line |\n" +
	"| `Ctrl+Y` | Paste the most-recently-deleted text |\n" +
	"| `Alt+Y` | Cycle through the deleted text after pasting |\n" +
	"| `{undo}` | Undo |\n" +
	"\n**Other**\n" +
	"| Key | Action |\n|-----|--------|\n" +
	"| `Tab` | Path completion / accept autocomplete |\n" +
	"| `Escape` | Cancel autocomplete / abort streaming |\n" +
	"| `Ctrl+C` | Clear editor (first) / exit (second) |\n" +
	"| `Ctrl+D` | Exit (when editor is empty) |\n" +
	"| `{suspend}` | Suspend to background |\n" +
	"| `Shift+Tab` | Cycle thinking level |\n" +
	"| `Ctrl+P` / `{cycleBack}` | Cycle models |\n" +
	"| `Ctrl+L` | Open model selector |\n" +
	"| `Ctrl+O` | Toggle tool output expansion |\n" +
	"| `Ctrl+T` | Toggle thinking block visibility |\n" +
	"| `Ctrl+G` | Edit message in external editor |\n" +
	"| `Ctrl+X` | Copy selection or last assistant message |\n" +
	"| `{followUp}` | Queue follow-up message |\n" +
	"| `{dequeue}` | Restore queued messages |\n" +
	"| `{paste}` | Paste files on macOS, images, or text from clipboard |\n" +
	"| `/` | Slash commands |\n" +
	"| `!` | Run bash command |\n" +
	"| `!!` | Run bash command (excluded from context) |\n"

// giHotkeysTemplate lists gi's keys beyond Pi's /hotkeys: Pi's fullscreen
// viewport keys and gi's own, on keys Pi leaves free.
const giHotkeysTemplate = "\n**gi**\n" +
	"| Key | Action |\n|-----|--------|\n" +
	"| `Ctrl+Home` / `Ctrl+End` | Scroll to top / bottom |\n" +
	"| `Ctrl+Up` / `Ctrl+Down` | Previous / next prompt |\n" +
	"| `{search}` | Search the transcript (`Enter`/`Ctrl+G` next, `Shift+Enter`/`Ctrl+Shift+G` previous) |\n" +
	"| `F6` / `F7` | Select the previous / next transcript block |\n" +
	"| `F8` | Expand or collapse the selected block |\n" +
	"| `Ctrl+R` | Search prompt history |\n" +
	"| `F2` / `F3` | Previous / next prompt from history |\n" +
	"| `Alt+L` | Cycle models backwards |\n" +
	"| `Alt+T` | Cycle thinking level backwards |\n" +
	"| `Alt+M` | Open model selector |\n" +
	"| `Alt+S` | Sessions |\n" +
	"| `Alt+C` | Compact the session |\n" +
	"| `Alt+I` | Workspace index |\n" +
	"| `@path` | Reference a file |\n"

func (c *chatTUI) hotkeyLines() []string {
	if c.regularMode {
		return []string{"hotkeys · regular mode", "Terminal wheel/selection/copy owns printed history; Home/End edit the draft", "Enter send · Shift+Enter newline · Alt-S sessions · Alt-M models · Alt-C compact · Alt-I index", "Printed output is immutable; use -tui-mode fullscreen for in-app paging and tool toggles"}
	}
	return append([]string{"Keyboard Shortcuts"}, renderChatMarkdown("assistant", "", defaultPiKeys.hotkeysMarkdown()+defaultPiKeys.giHotkeysMarkdown(), c.transcriptRenderWidth())...)
}
