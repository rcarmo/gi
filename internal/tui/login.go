package tui

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"runtime"
	"strings"

	gotui "github.com/grindlemire/go-tui"
	"github.com/rcarmo/gi/internal/inference"
	"github.com/rcarmo/go-ai/oauth"
)

// /login and /logout are Pi's (interactive-mode.js handleLoginCommand,
// showOAuthSelector): an authentication-method menu, Pi's provider selector
// (OAuthSelectorComponent) and its login dialog (LoginDialogComponent),
// which runs go-ai's sign-in flow or takes an API key. Golden:
// scripts/golden-login.mjs. See docs/internal/tui-oauth-login.md.

var errLoginCancelled = errors.New("Login cancelled")

type authSelectorState struct {
	mode     string // "login" or "logout"
	all      []inference.LoginOption
	query    string
	selected int
	onSelect func(inference.LoginOption)
	onCancel func()
}

func (s *authSelectorState) filtered() []inference.LoginOption {
	if strings.TrimSpace(s.query) == "" {
		return s.all
	}
	items := make([]slashItem, len(s.all))
	for i, p := range s.all {
		items[i] = slashItem{value: fmt.Sprint(i), description: p.Name + " " + p.ID + " " + p.AuthType + " "}
	}
	var out []inference.LoginOption
	for _, it := range piFuzzyFilter(items, s.query, func(it slashItem) string { return it.description }) {
		var i int
		fmt.Sscan(it.value, &i)
		out = append(out, s.all[i])
	}
	return out
}

func (c *chatTUI) openAuthSelector(mode string, options []inference.LoginOption, query string, onSelect func(inference.LoginOption), onCancel func()) {
	c.authSelector = &authSelectorState{mode: mode, all: options, query: query, onSelect: onSelect, onCancel: onCancel}
	c.openMenuKind("auth-selector")
}

// openMenuKind shows a selector that replaces the editor.
func (c *chatTUI) openMenuKind(kind string) {
	c.modelMenuOpen = true
	c.modelMenuKind = kind
	c.modelMenuError = ""
	c.inputActive = false
	if c.app != nil {
		c.app.BlurFocused()
		c.app.MarkDirty()
	}
}

func (c *chatTUI) authSelectorKeys() gotui.KeyMap {
	s := c.authSelector
	move := func(delta int) {
		if n := len(s.filtered()); n > 0 {
			s.selected = max(0, min(n-1, s.selected+delta))
			c.markDirty()
		}
	}
	search := func(query string) {
		s.query = query
		s.selected = max(0, min(s.selected, len(s.filtered())-1))
		c.markDirty()
	}
	finish := func(choice *inference.LoginOption) {
		c.authSelector = nil
		c.closeModelMenu()
		if choice != nil {
			s.onSelect(*choice)
		} else if s.onCancel != nil {
			s.onCancel()
		}
		c.markDirty()
	}
	cancel := func(gotui.KeyEvent) { finish(nil) }
	page := func(direction int) {
		s.selected = selectPageIndex(s.selected, len(s.filtered()), 8, direction)
		c.markDirty()
	}
	return append(selectorPageKeys(page), gotui.KeyMap{
		gotui.OnPreemptStop(gotui.KeyUp, func(gotui.KeyEvent) { move(-1) }),
		gotui.OnPreemptStop(gotui.KeyDown, func(gotui.KeyEvent) { move(1) }),
		gotui.OnPreemptStop(gotui.KeyEnter, func(gotui.KeyEvent) {
			if list := s.filtered(); s.selected < len(list) {
				finish(&list[s.selected])
			}
		}),
		gotui.OnPreemptStop(gotui.KeyEscape, cancel),
		gotui.OnPreemptStop(gotui.KeyCtrlC, cancel),
		gotui.OnPreemptStop(gotui.KeyBackspace, func(gotui.KeyEvent) {
			if r := []rune(s.query); len(r) > 0 {
				search(string(r[:len(r)-1]))
			}
		}),
		gotui.OnFocused(gotui.AnyRune, func(ke gotui.KeyEvent) { search(s.query + string(ke.Rune)) }),
	}...)
}

// authTypeLabel is Pi's formatAuthSelectorProviderType.
func authTypeLabel(authType string, subscription bool) string {
	switch {
	case authType == "api_key":
		return "API key"
	case subscription:
		return "subscription"
	}
	return "account"
}

var envNamesPattern = regexp.MustCompile(`^[A-Z][A-Z0-9_]*(?:, [A-Z][A-Z0-9_]*)*$`)

// authStatusSpans is Pi's formatAuthSelectorProviderStatus.
func authStatusSpans(p inference.LoginOption) []gotui.TextSpan {
	switch {
	case p.StatusType == "":
		return []gotui.TextSpan{{Text: " • not configured", Style: piFg(piMuted)}}
	case p.StatusType != p.AuthType:
		return []gotui.TextSpan{{Text: " • ", Style: piFg(piMuted)}, {Text: authTypeLabel(p.StatusType, p.Subscription) + " configured", Style: piFg(piWarning)}}
	case p.StatusSource == "" || p.StatusSource == "OAuth" || p.StatusSource == "stored credential":
		return []gotui.TextSpan{{Text: " ✓ configured", Style: piFg(piSuccess)}}
	}
	source := p.StatusSource
	if envNamesPattern.MatchString(source) {
		source = "env: " + source
	}
	return []gotui.TextSpan{{Text: " ✓ " + source, Style: piFg(piSuccess)}}
}

// piTruncatedText is pi-tui's TruncatedText with paddingX 1: one line,
// cut with "..." to the width less the padding.
func piTruncatedText(width int, spans ...gotui.TextSpan) []gotui.TextSpan {
	return append([]gotui.TextSpan{{Text: " "}}, piTruncate(spans, max(1, width-2))...)
}

// piTruncate is pi-tui's truncateToWidth: spans wider than width are cut
// and end in "...".
func piTruncate(spans []gotui.TextSpan, width int) []gotui.TextSpan {
	total := 0
	for _, s := range spans {
		total += gotui.StringWidth(s.Text)
	}
	if total <= width {
		return spans
	}
	return append(clipSpans(spans, max(0, width-3)), gotui.TextSpan{Text: "..."})
}

func (c *chatTUI) piAuthSelectorRows(width int) spanRows {
	s := c.authSelector
	title := "Select provider to logout:"
	if s.mode == "login" {
		title = "Select provider to configure:"
	}
	rows := spanRows{piRule(width, piBorder), nil, piTruncatedText(width, gotui.TextSpan{Text: title, Style: piFg(piAccent).Bold()}), nil, piSearchRow(s.query, width), nil}
	types := map[string]bool{}
	for _, p := range s.all {
		types[p.AuthType] = true
	}
	list := s.filtered()
	start := max(0, min(s.selected-4, len(list)-8))
	end := min(start+8, len(list))
	for i := start; i < end; i++ {
		p := list[i]
		line := []gotui.TextSpan{{Text: "  "}, {Text: p.Name, Style: piFg(piText)}}
		if i == s.selected {
			line = []gotui.TextSpan{{Text: "→ ", Style: piFg(piAccent)}, {Text: p.Name, Style: piFg(piAccent)}}
		}
		if len(types) > 1 {
			line = append(line, gotui.TextSpan{Text: " [" + authTypeLabel(p.AuthType, p.Subscription) + "]", Style: piFg(piMuted)})
		}
		rows = append(rows, piTruncatedText(width, append(line, authStatusSpans(p)...)...))
	}
	if start > 0 || end < len(list) {
		rows = append(rows, piTruncatedText(width, gotui.TextSpan{Text: fmt.Sprintf("  (%d/%d)", s.selected+1, len(list)), Style: piFg(piMuted)}))
	}
	if len(list) == 0 {
		message := "No matching providers"
		switch {
		case len(s.all) == 0 && s.mode == "login":
			message = "No providers available"
		case len(s.all) == 0:
			message = "No providers logged in. Use /login first."
		}
		rows = append(rows, piTruncatedText(width, gotui.TextSpan{Text: "  " + message, Style: piFg(piMuted)}))
	}
	return append(rows, nil, piRule(width, piBorder))
}

// loginDialogState is Pi's LoginDialogComponent: a title, content that the
// sign-in flow adds to, and an input while it waits for an answer.
type loginDialogState struct {
	title     string
	content   []loginLine
	input     string
	answer    chan string // set while a prompt waits
	cancelled chan struct{}
	onCancel  func()
	authURL   *authURL // the shown sign-in URL, which Ctrl+X copies (Pi)
}

type loginLine struct {
	spans []gotui.TextSpan
	input bool // the live input
	auth  bool // the sign-in URL
}

func (d *loginDialogState) add(spans ...gotui.TextSpan) {
	d.content = append(d.content, loginLine{spans: spans})
}
func (d *loginDialogState) spacer() { d.content = append(d.content, loginLine{}) }

func clickHint() string {
	if runtime.GOOS == "darwin" {
		return "Cmd+click to open"
	}
	return "Ctrl+click to open"
}

// showAuth is Pi's: the URL with its click and copy hints (AuthUrlComponent)
// and any instructions.
func (d *loginDialogState) showAuth(url, instructions string) {
	d.content = nil
	d.spacer()
	d.authURL = newAuthURL(url)
	d.content = append(d.content, loginLine{auth: true})
	if instructions != "" {
		d.spacer()
		d.add(gotui.TextSpan{Text: instructions, Style: piFg(piWarning)})
	}
}

// showDeviceCode is Pi's: the verification URL with its click hint (no copy
// key) and the code to enter. go-ai's device flows report through OnAuth
// with "Enter code: …" instructions until it has device-code events.
func (d *loginDialogState) showDeviceCode(url, code string) {
	d.authURL = nil
	d.content = nil
	d.spacer()
	d.add(gotui.TextSpan{Text: url, Style: piFg(piAccent), Link: url})
	d.add(gotui.TextSpan{Text: clickHint(), Style: piFg(piDim), Link: url})
	d.spacer()
	d.add(gotui.TextSpan{Text: "Enter code: " + code, Style: piFg(piWarning)})
}

// showPrompt is Pi's: the message, an example, the input and its keys.
func (d *loginDialogState) showPrompt(message, placeholder string) {
	d.spacer()
	d.add(gotui.TextSpan{Text: message, Style: piFg(piText)})
	if placeholder != "" {
		d.add(gotui.TextSpan{Text: "e.g., " + placeholder, Style: piFg(piDim)})
	}
	d.input = ""
	d.content = append(d.content, loginLine{input: true})
	hint := append([]gotui.TextSpan{{Text: "("}}, keyHintSpans("escape/ctrl+c", "to cancel,")...)
	hint = append(append(hint, gotui.TextSpan{Text: " "}), keyHintSpans("enter", "to submit")...)
	d.add(append(hint, gotui.TextSpan{Text: ")"})...)
	d.answer = make(chan string, 1)
}

// showWaiting is Pi's, for polling flows.
func (d *loginDialogState) showWaiting(message string) {
	d.spacer()
	d.add(gotui.TextSpan{Text: message, Style: piFg(piDim)})
	d.add(append(append([]gotui.TextSpan{{Text: "("}}, keyHintSpans("escape/ctrl+c", "to cancel")...), gotui.TextSpan{Text: ")"})...)
}

func (d *loginDialogState) showProgress(message string) {
	d.add(gotui.TextSpan{Text: message, Style: piFg(piDim)})
}

func (c *chatTUI) piLoginDialogRows(width int) spanRows {
	d := c.loginDialog
	text := func(spans ...gotui.TextSpan) spanRows {
		var rows spanRows
		for _, line := range piWrapLine(spans, max(1, width-2)) {
			rows = append(rows, append([]gotui.TextSpan{{Text: " "}}, line...))
		}
		return rows
	}
	rows := spanRows{piRule(width, piBorder)}
	rows = append(rows, text(gotui.TextSpan{Text: d.title, Style: piFg(piAccent).Bold()})...)
	for _, line := range d.content {
		switch {
		case line.input && d.answer != nil:
			rows = append(rows, piSearchRow(d.input, width))
		case line.input:
			rows = append(rows, []gotui.TextSpan{{Text: "> " + d.input}})
		case line.auth && d.authURL != nil:
			for _, spans := range d.authURL.spans() {
				rows = append(rows, text(spans...)...)
			}
		case line.spans == nil:
			rows = append(rows, nil)
		default:
			rows = append(rows, text(line.spans...)...)
		}
	}
	return append(rows, piRule(width, piBorder))
}

func (c *chatTUI) loginDialogKeys() gotui.KeyMap {
	d := c.loginDialog
	edit := func(input string) {
		if d.answer != nil {
			d.input = input
			c.markDirty()
		}
	}
	return gotui.KeyMap{
		gotui.OnPreemptStop(gotui.KeyEscape, func(gotui.KeyEvent) { c.cancelLoginDialog() }),
		gotui.OnPreemptStop(gotui.KeyCtrlC, func(gotui.KeyEvent) { c.cancelLoginDialog() }),
		gotui.OnPreemptStop(gotui.KeyCtrlX, func(gotui.KeyEvent) {
			if d.authURL != nil {
				c.copyAuthURL(d.authURL)
			}
		}),
		gotui.OnPreemptStop(gotui.KeyEnter, func(gotui.KeyEvent) {
			if d.answer != nil {
				answer := d.answer
				d.answer = nil // the input line now shows the submitted text
				answer <- d.input
				c.markDirty()
			}
		}),
		gotui.OnPreemptStop(gotui.KeyBackspace, func(gotui.KeyEvent) {
			if r := []rune(d.input); len(r) > 0 {
				edit(string(r[:len(r)-1]))
			}
		}),
		gotui.OnFocused(gotui.AnyRune, func(ke gotui.KeyEvent) { edit(d.input + string(ke.Rune)) }),
	}
}

func (c *chatTUI) openLoginDialog(title string, onCancel func()) *loginDialogState {
	c.loginDialog = &loginDialogState{title: title, cancelled: make(chan struct{}), onCancel: onCancel}
	c.openMenuKind("login-dialog")
	return c.loginDialog
}

// cancelLoginDialog is Pi's cancel: the flow stops waiting and the editor
// returns (Pi then reopens the menu the login started from).
func (c *chatTUI) cancelLoginDialog() {
	d := c.loginDialog
	if d == nil {
		return
	}
	c.closeLoginDialog()
	close(d.cancelled)
	if d.onCancel != nil {
		d.onCancel()
	}
}

func (c *chatTUI) closeLoginDialog() {
	c.loginDialog = nil
	if c.modelMenuKind == "login-dialog" {
		c.closeModelMenu()
	}
	c.markDirty()
}

// prompt waits for the dialog's answer from the sign-in goroutine.
func (c *chatTUI) loginPrompt(d *loginDialogState, show func()) (string, error) {
	return c.loginPromptContext(context.Background(), d, show)
}

// loginPromptContext is loginPrompt that ctx can end: when the browser's
// callback wins over a pasted code (go-ai's OnPromptContext, Pi's manual
// input race), the input stops taking an answer.
func (c *chatTUI) loginPromptContext(ctx context.Context, d *loginDialogState, show func()) (string, error) {
	ready := make(chan chan string, 1)
	c.runOnUI(func() {
		if c.loginDialog != d {
			ready <- nil
			return
		}
		show()
		ready <- d.answer
		c.markDirty()
	})
	select {
	case answer := <-ready:
		if answer == nil {
			return "", errLoginCancelled
		}
		select {
		case value := <-answer:
			return value, nil
		case <-d.cancelled:
			return "", errLoginCancelled
		case <-ctx.Done():
			c.runOnUI(func() {
				if d.answer == answer {
					d.answer = nil
					c.markDirty()
				}
			})
			return "", ctx.Err()
		}
	case <-d.cancelled:
		return "", errLoginCancelled
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

// handleLoginCommand is Pi's: no argument opens the method menu; a provider
// signs in directly, or picks among its methods.
func (c *chatTUI) handleLoginCommand(provider string) {
	if provider == "" {
		c.showLoginAuthTypeSelector(nil)
		return
	}
	want := strings.ToLower(provider)
	var matches []inference.LoginOption
	for _, p := range inference.LoginOptions("") {
		if strings.ToLower(p.ID) == want || strings.ToLower(p.Name) == want {
			matches = append(matches, p)
		}
	}
	switch {
	case len(matches) == 1:
		c.startProviderLogin(matches[0], nil)
	case len(matches) > 1:
		c.showLoginAuthTypeSelector(matches)
	default:
		c.showLoginProviderSelector("", provider)
	}
}

const (
	loginAccountLabel = "Sign in with an account"
	loginAPIKeyLabel  = "Sign in with an API key"
)

// showLoginAuthTypeSelector is Pi's, over all providers or one provider's
// methods.
func (c *chatTUI) showLoginAuthTypeSelector(options []inference.LoginOption) {
	types := map[string]bool{"oauth": options == nil, "api_key": options == nil}
	for _, p := range options {
		types[p.AuthType] = true
	}
	var labels []string
	if types["oauth"] {
		labels = append(labels, loginAccountLabel)
	}
	if types["api_key"] {
		labels = append(labels, loginAPIKeyLabel)
	}
	if options != nil && len(labels) == 1 {
		c.startProviderLogin(options[0], nil)
		return
	}
	title := "Select authentication method:"
	if len(options) > 0 {
		title = fmt.Sprintf("Select authentication method for %s:", options[0].Name)
	}
	c.openSelect(title, labels, func(label string) {
		authType := "api_key"
		if label == loginAccountLabel {
			authType = "oauth"
		}
		if options == nil {
			c.showLoginProviderSelector(authType, "")
			return
		}
		for _, p := range options {
			if p.AuthType == authType {
				c.startProviderLogin(p, func() { c.showLoginAuthTypeSelector(options) })
			}
		}
	}, nil)
}

// showLoginProviderSelector is Pi's: the providers for one method.
func (c *chatTUI) showLoginProviderSelector(authType, query string) {
	options := inference.LoginOptions(authType)
	if len(options) == 0 {
		message := "No login providers available."
		switch authType {
		case "oauth":
			message = "No account providers available."
		case "api_key":
			message = "No API key providers available."
		}
		c.appendTranscript(message)
		return
	}
	c.openAuthSelector("login", options, query, func(p inference.LoginOption) {
		c.startProviderLogin(p, func() { c.showLoginProviderSelector(authType, query) })
	}, func() {
		if authType != "" {
			c.showLoginAuthTypeSelector(nil)
		}
	})
}

// startProviderLogin is Pi's: the OAuth flow or an API key prompt; onBack
// reopens the menu the login started from when it is cancelled.
func (c *chatTUI) startProviderLogin(p inference.LoginOption, onBack func()) {
	d := c.openLoginDialog("Login to "+p.Name, onBack)
	if p.AuthType == "api_key" {
		go func() {
			key, err := c.loginPrompt(d, func() { d.showPrompt("Enter API key:", "") })
			if err == nil {
				err = inference.SaveAPIKeyLogin(p.ID, key)
			}
			c.runOnUI(func() { c.finishProviderLogin(d, p, err) })
		}()
		return
	}
	callbacks := oauth.LoginCallbacks{
		OnAuth: func(info oauth.AuthInfo) {
			c.runOnUI(func() {
				if c.loginDialog == d {
					d.showAuth(info.URL, info.Instructions)
					c.markDirty()
				}
			})
			openBrowser(info.URL)
		},
		OnPrompt: func(prompt oauth.Prompt) (string, error) {
			return c.loginPrompt(d, func() { d.showPrompt(prompt.Message, prompt.Placeholder) })
		},
		OnPromptContext: func(ctx context.Context, prompt oauth.Prompt) (string, error) {
			return c.loginPromptContext(ctx, d, func() { d.showPrompt(prompt.Message, prompt.Placeholder) })
		},
		OnSelect: func(prompt oauth.SelectPrompt) (string, error) {
			return c.loginSelect(d, prompt)
		},
		OnProgress: func(message string) {
			c.runOnUI(func() {
				if c.loginDialog == d {
					d.showProgress(message)
					c.markDirty()
				}
			})
		},
	}
	go func() {
		creds, err := inference.OAuthLogin(p.ID, callbacks)
		if err == nil {
			err = inference.SaveOAuthLogin(p.ID, creds)
		}
		c.runOnUI(func() { c.finishProviderLogin(d, p, err) })
	}()
}

// loginSelect is Pi's showAuthSelect: a menu over the dialog.
func (c *chatTUI) loginSelect(d *loginDialogState, prompt oauth.SelectPrompt) (string, error) {
	result := make(chan string, 1)
	c.runOnUI(func() {
		if c.loginDialog != d {
			result <- ""
			return
		}
		labels := make([]string, len(prompt.Options))
		for i, o := range prompt.Options {
			labels[i] = o.Label
		}
		restore := func() { c.loginDialog = d; c.openMenuKind("login-dialog") }
		c.openSelect(prompt.Message, labels, func(label string) {
			restore()
			for _, o := range prompt.Options {
				if o.Label == label {
					result <- o.Value
					return
				}
			}
			result <- ""
		}, func() {
			restore()
			c.cancelLoginDialog()
			result <- ""
		})
	})
	select {
	case value := <-result:
		if value == "" {
			return "", errLoginCancelled
		}
		return value, nil
	case <-d.cancelled:
		return "", errLoginCancelled
	}
}

// finishProviderLogin is Pi's completion: the dialog closes with a status,
// or an error; a cancelled login has already closed.
func (c *chatTUI) finishProviderLogin(d *loginDialogState, p inference.LoginOption, err error) {
	select {
	case <-d.cancelled:
		return
	default:
	}
	c.closeLoginDialog()
	action := "Logged in to " + p.Name
	if p.AuthType == "api_key" {
		action = "Saved API key for " + p.Name
	}
	switch {
	case err == nil:
		c.appendTranscript(fmt.Sprintf("%s. Credentials saved to %s", action, inference.AuthFilePath()))
	case p.AuthType == "api_key":
		c.appendTranscript(fmt.Sprintf("error: Failed to save API key for %s: %v", p.Name, err))
	default:
		c.appendTranscript(fmt.Sprintf("error: Failed to login to %s: %v", p.Name, err))
	}
}

// handleLogoutCommand is Pi's showOAuthSelector("logout").
func (c *chatTUI) handleLogoutCommand() {
	options, err := inference.LogoutOptions()
	if err != nil {
		c.appendTranscript(fmt.Sprintf("error: Could not read stored credentials: %v", err))
		return
	}
	if len(options) == 0 {
		c.appendTranscript("No stored credentials to remove. /logout only removes credentials saved by /login; environment variables and models.json config are unchanged.")
		return
	}
	c.openAuthSelector("logout", options, "", func(p inference.LoginOption) {
		if _, err := inference.RemoveAuthEntry(p.ID); err != nil {
			c.appendTranscript(fmt.Sprintf("error: Logout failed: %v", err))
			return
		}
		if p.AuthType == "oauth" {
			c.appendTranscript("Logged out of " + p.Name)
		} else {
			c.appendTranscript(fmt.Sprintf("Removed stored API key for %s. Environment variables and models.json config are unchanged.", p.Name))
		}
	}, nil)
}
