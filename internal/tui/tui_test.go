package tui

import (
	"os"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/cggithub333/deepmd/internal/config"
	"github.com/cggithub333/deepmd/internal/finder"
	"github.com/cggithub333/deepmd/internal/model"
)

func TestNewApp(t *testing.T) {
	files := []model.FileInfo{
		{RelPath: "README.md", Path: "/tmp/README.md"},
		{RelPath: "docs/guide.md", Path: "/tmp/docs/guide.md"},
	}

	app := NewApp(files, "/tmp")
	if len(app.Files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(app.Files))
	}
	if app.State != StateDualPane {
		t.Fatalf("expected StateDualPane, got %v", app.State)
	}
}

func TestApp_NavigationAndReaderTransition(t *testing.T) {
	files := []model.FileInfo{
		{RelPath: "README.md", Path: "/tmp/README.md"},
	}

	app := NewApp(files, "/tmp")

	// Trigger Enter key
	updated, _ := app.Update(tea.KeyMsg{Type: tea.KeyEnter})
	appModel := updated.(App)

	if appModel.State != StateReader {
		t.Fatalf("expected transition to StateReader on Enter, got %v", appModel.State)
	}

	// Trigger Esc key to return
	updated2, _ := appModel.Update(tea.KeyMsg{Type: tea.KeyEsc})
	appModel2 := updated2.(App)

	if appModel2.State != StateDualPane {
		t.Fatalf("expected return to StateDualPane on Esc, got %v", appModel2.State)
	}
}

func TestApp_WindowResizeAndTooSmallGuard(t *testing.T) {
	app := NewApp(nil, ".")

	// Very small terminal
	updated, _ := app.Update(tea.WindowSizeMsg{Width: 40, Height: 10})
	smallApp := updated.(App)

	if !smallApp.Layout.TooSmall {
		t.Fatal("expected TooSmall guard to be true for 40x10 terminal")
	}

	// Standard terminal
	updated2, _ := app.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	stdApp := updated2.(App)

	if stdApp.Layout.TooSmall {
		t.Fatal("expected TooSmall guard to be false for 100x30 terminal")
	}
}

func TestTruncateString(t *testing.T) {
	cases := []struct {
		input    string
		maxW     int
		expected string
	}{
		{"hello", 0, ""},
		{"hello", -5, ""},
		{"hello", 2, "he"},
		{"hello", 3, "hel"},
		{"hello", 4, "h..."},
		{"hello world", 8, "hello..."},
		{"hello", 10, "hello"},
	}

	for _, c := range cases {
		res := TruncateString(c.input, c.maxW)
		if res != c.expected {
			t.Errorf("TruncateString(%q, %d) = %q; want %q", c.input, c.maxW, res, c.expected)
		}
	}
}

func TestListPane_UninitializedAndSmallWidth(t *testing.T) {
	files := []model.FileInfo{
		{RelPath: "very/long/nested/path/to/some/documentation/file.md", Path: "/tmp/file.md"},
	}

	// 1. Test completely uninitialized width (0) - must not panic!
	lp := NewListPane(files)
	view0 := lp.View()
	if view0 == "" {
		t.Error("expected non-empty view for uninitialized ListPane")
	}

	// 2. Test very small widths (5, 8, 10)
	widths := []int{1, 5, 8, 10, 15}
	for _, w := range widths {
		lp.SetDimensions(w, 20)
		view := lp.View()
		if view == "" {
			t.Errorf("expected non-empty view for width %d", w)
		}
	}

	// 3. Test content search mode with uninitialized and small widths
	lp.Mode = SearchModeContent
	lp.ContentMatches = []finder.ContentMatch{
		{File: files[0], LineNum: 42, LineText: "some deeply nested content match line text here"},
	}
	for _, w := range []int{0, 5, 10, 20} {
		lp.SetDimensions(w, 20)
		view := lp.View()
		if view == "" {
			t.Errorf("expected non-empty content view for width %d", w)
		}
	}
}

func TestApp_InitialRenderWithoutWindowSizeMsg(t *testing.T) {
	files := []model.FileInfo{
		{RelPath: "README.md", Path: "/tmp/README.md"},
		{RelPath: "docs/architecture.md", Path: "/tmp/docs/architecture.md"},
	}

	app := NewApp(files, "/tmp")

	// Render immediately without sending tea.WindowSizeMsg (simulates first frame)
	view := app.View()
	if view == "" {
		t.Fatal("expected non-empty initial view")
	}
}

func TestApp_LeaderModeViewSwitchAndResize(t *testing.T) {
	files := []model.FileInfo{
		{RelPath: "README.md", Path: "/tmp/README.md"},
	}

	app := NewApp(files, "/tmp")
	if app.Focus != FocusList {
		t.Fatalf("expected initial FocusList, got %v", app.Focus)
	}

	// 1. Tab must NOT toggle focus anymore (unbound per user requirement)
	tabUpdated, _ := app.Update(tea.KeyMsg{Type: tea.KeyTab})
	tabApp := tabUpdated.(App)
	if tabApp.Focus != FocusList {
		t.Fatalf("expected FocusList because Tab is unbound, got %v", tabApp.Focus)
	}

	// 2. Press Alt+Q (Leader) then Right/l to switch view to Preview
	leader1, _ := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}, Alt: true})
	leaderApp1 := leader1.(App)
	if !leaderApp1.LeaderMode {
		t.Fatal("expected LeaderMode to be true after Alt+Q")
	}

	switchedToPreview, _ := leaderApp1.Update(tea.KeyMsg{Type: tea.KeyRight})
	previewApp := switchedToPreview.(App)
	if previewApp.Focus != FocusPreview {
		t.Fatalf("expected FocusPreview after Leader + Right, got %v", previewApp.Focus)
	}
	if previewApp.LeaderMode {
		t.Fatal("expected LeaderMode to be false after switching view")
	}

	// 3. Press Alt+Q (Leader) then Left/h to switch view back to List
	leader2, _ := previewApp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}, Alt: true})
	leaderApp2 := leader2.(App)
	switchedToList, _ := leaderApp2.Update(tea.KeyMsg{Type: tea.KeyLeft})
	listApp := switchedToList.(App)
	if listApp.Focus != FocusList {
		t.Fatalf("expected FocusList after Leader + Left, got %v", listApp.Focus)
	}

	// 4. In Preview, Esc returns focus to List
	toPrev2, _ := listApp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}, Alt: true})
	toPrevApp2, _ := toPrev2.(App).Update(tea.KeyMsg{Type: tea.KeyRight})
	prevFocusApp := toPrevApp2.(App)
	escBack, _ := prevFocusApp.Update(tea.KeyMsg{Type: tea.KeyEsc})
	escApp := escBack.(App)
	if escApp.Focus != FocusList {
		t.Fatalf("expected FocusList after Esc from preview, got %v", escApp.Focus)
	}
}

func TestApp_MouseWheelNeverScrollsList(t *testing.T) {
	files := []model.FileInfo{
		{RelPath: "file1.md", Path: "/tmp/file1.md"},
		{RelPath: "file2.md", Path: "/tmp/file2.md"},
		{RelPath: "file3.md", Path: "/tmp/file3.md"},
	}

	cfg := config.Config{
		MouseWheelEnabled: true,
		MouseWheelDelta:   2,
	}
	app := NewAppWithConfig(files, "/tmp", cfg)
	app.Preview.SetContent(strings.Repeat("line in document\n", 50))
	app.Preview.SetDimensions(40, 10)

	dividerX := app.Layout.LeftWidth + 2

	// 1. Send WheelDown over List Pane (X < dividerX) - cursor MUST NOT change
	wheelDownList := tea.MouseMsg{
		X:      dividerX - 5,
		Button: tea.MouseButtonWheelDown,
	}
	updated1, _ := app.Update(wheelDownList)
	appResult1 := updated1.(App)
	if appResult1.List.Cursor != 0 {
		t.Fatalf("expected cursor to remain 0 when scrolling over list, got %d", appResult1.List.Cursor)
	}

	// 2. Send WheelDown over Preview Pane (X >= dividerX) - preview viewport MUST scroll down
	wheelDownPreview := tea.MouseMsg{
		X:      dividerX + 5,
		Button: tea.MouseButtonWheelDown,
	}
	updated2, _ := appResult1.Update(wheelDownPreview)
	appResult2 := updated2.(App)
	if appResult2.Preview.Viewport.YOffset <= 0 {
		t.Fatalf("expected preview viewport YOffset to increase on WheelDown, got %d", appResult2.Preview.Viewport.YOffset)
	}

	// 3. Send WheelUp over Preview Pane - preview viewport MUST scroll back up
	wheelUpPreview := tea.MouseMsg{
		X:      dividerX + 5,
		Button: tea.MouseButtonWheelUp,
	}
	updated3, _ := appResult2.Update(wheelUpPreview)
	appResult3 := updated3.(App)
	if appResult3.Preview.Viewport.YOffset != 0 {
		t.Fatalf("expected preview viewport YOffset to return to 0 on WheelUp, got %d", appResult3.Preview.Viewport.YOffset)
	}
}

func TestApp_DividerMouseDrag(t *testing.T) {
	files := []model.FileInfo{
		{RelPath: "file1.md", Path: "/tmp/file1.md"},
	}
	app := NewApp(files, "/tmp")

	dividerX := app.Layout.LeftWidth + 2

	// 1. Mouse press on divider
	pressMsg := tea.MouseMsg{
		X:      dividerX,
		Y:      5,
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
	}
	updated, _ := app.Update(pressMsg)
	appResult := updated.(App)
	if !appResult.DraggingDivider {
		t.Fatal("expected DraggingDivider to be true after clicking on divider")
	}

	// 2. Drag mouse to the right (e.g. +10 cols)
	dragMsg := tea.MouseMsg{
		X:      dividerX + 10,
		Y:      5,
		Action: tea.MouseActionMotion,
		Button: tea.MouseButtonLeft,
	}
	updated2, _ := appResult.Update(dragMsg)
	appResult2 := updated2.(App)
	if appResult2.CustomLeftWidth <= app.Layout.LeftWidth {
		t.Fatalf("expected CustomLeftWidth to increase after dragging right, got %d", appResult2.CustomLeftWidth)
	}

	// 3. Mouse release
	releaseMsg := tea.MouseMsg{
		X:      dividerX + 10,
		Y:      5,
		Action: tea.MouseActionRelease,
		Button: tea.MouseButtonNone,
	}
	updated3, _ := appResult2.Update(releaseMsg)
	appResult3 := updated3.(App)
	if appResult3.DraggingDivider {
		t.Fatal("expected DraggingDivider to be false after mouse release")
	}
}

func TestApp_TmuxResize(t *testing.T) {
	files := []model.FileInfo{
		{RelPath: "file1.md", Path: "/tmp/file1.md"},
	}
	app := NewApp(files, "/tmp")
	initialLeft := app.Layout.LeftWidth

	// 1. Direct Alt+Right increases width
	updated, _ := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("alt+right")})
	appR := updated.(App)
	// Or msg with string "alt+right"
	if appR.Layout.LeftWidth < initialLeft {
		t.Fatalf("expected width to increase or stay equal")
	}

	// 2. Press Alt+Q to toggle LeaderMode
	updated2, _ := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}, Alt: true})
	appResize := updated2.(App)
	if !appResize.LeaderMode {
		t.Fatal("expected LeaderMode to be true after Alt+Q")
	}

	// Test AdjustSplit directly
	app.AdjustSplit(4)
	if app.Layout.LeftWidth != initialLeft+4 {
		t.Fatalf("expected left width to be %d, got %d", initialLeft+4, app.Layout.LeftWidth)
	}
	app.AdjustSplit(-4)
	if app.Layout.LeftWidth != initialLeft {
		t.Fatalf("expected left width to return to %d, got %d", initialLeft, app.Layout.LeftWidth)
	}
}

func TestApp_InPreviewGrep(t *testing.T) {
	files := []model.FileInfo{
		{RelPath: "file1.md", Path: "/tmp/file1.md"},
	}
	app := NewApp(files, "/tmp")
	app.Preview.SetContent("Line 1: Introduction\nLine 2: Important Topic\nLine 3: Details here\nLine 4: Another topic mention")

	// 1. Press Ctrl+F to activate in-preview search
	updated, _ := app.Update(tea.KeyMsg{Type: tea.KeyCtrlF})
	appSearch := updated.(App)

	if appSearch.Focus != FocusPreview {
		t.Fatalf("expected FocusPreview on Ctrl+F, got %v", appSearch.Focus)
	}
	if !appSearch.Preview.Searching {
		t.Fatal("expected Preview.Searching to be true after Ctrl+F")
	}

	// 2. Type query "topic"
	appSearch.Preview.SearchInput.SetValue("topic")
	appSearch.Preview.applySearch("topic")

	if len(appSearch.Preview.MatchLines) != 2 {
		t.Fatalf("expected 2 matches for 'topic', got %d", len(appSearch.Preview.MatchLines))
	}

	// 3. Next match navigation
	appSearch.Preview.NextMatch()
	if appSearch.Preview.CurrentMatch != 1 {
		t.Fatalf("expected CurrentMatch to be 1, got %d", appSearch.Preview.CurrentMatch)
	}

	// 4. View contains search box
	view := appSearch.Preview.View()
	if !strings.Contains(view, "grep") {
		t.Fatal("expected Preview View to contain grep search box")
	}

	// 5. Esc closes search
	updated2, _ := appSearch.Update(tea.KeyMsg{Type: tea.KeyEsc})
	appClosed := updated2.(App)
	if appClosed.Preview.Searching {
		t.Fatal("expected Preview.Searching to be false after Esc")
	}
}

func TestApp_ClipboardCopy(t *testing.T) {
	files := []model.FileInfo{
		{RelPath: "test.md", Path: "/tmp/deepmd-test.md"},
	}
	_ = os.WriteFile("/tmp/deepmd-test.md", []byte("# Header\nContent line"), 0644)
	app := NewApp(files, "/tmp")

	// 1. Ctrl+P in DualPane copies full absolute path
	updatedP, _ := app.Update(tea.KeyMsg{Type: tea.KeyCtrlP})
	appP := updatedP.(App)
	if !strings.Contains(appP.Notification, "/tmp/deepmd-test.md") {
		t.Fatalf("expected notification to contain full path '/tmp/deepmd-test.md', got %q", appP.Notification)
	}

	// Also verify Ctrl+Y copies full path
	updatedY, _ := app.Update(tea.KeyMsg{Type: tea.KeyCtrlY})
	appY := updatedY.(App)
	if !strings.Contains(appY.Notification, "/tmp/deepmd-test.md") {
		t.Fatalf("expected notification to contain full path '/tmp/deepmd-test.md', got %q", appY.Notification)
	}

	// 2. Ctrl+A copies content
	updatedA, _ := app.Update(tea.KeyMsg{Type: tea.KeyCtrlA})
	appA := updatedA.(App)
	if !strings.Contains(appA.Notification, "Copied content") {
		t.Fatalf("expected notification to contain 'Copied content', got %q", appA.Notification)
	}
}

func TestApp_ExitShortcuts(t *testing.T) {
	files := []model.FileInfo{
		{RelPath: "test.md", Path: "/tmp/test.md"},
	}
	app := NewApp(files, "/tmp")

	// 1. Ctrl+C exits immediately
	_, cmdCtrlC := app.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmdCtrlC == nil {
		t.Fatal("expected tea.Quit command on Ctrl+C")
	}

	// 2. 'q' does NOT exit in DualPane (it types into search input)
	updQ, _ := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	appQ := updQ.(App)
	if appQ.List.TextInput.Value() != "q" {
		t.Fatalf("expected 'q' to be entered into search input, got %q", appQ.List.TextInput.Value())
	}

	// 3. Typing 'exit' in search input and pressing Enter exits
	app.List.TextInput.SetValue("exit")
	_, cmdEnterExit := app.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmdEnterExit == nil {
		t.Fatal("expected tea.Quit command on Enter with 'exit'")
	}
}

func TestCalculateLayoutWithRatio(t *testing.T) {
	// Standard width 100
	l100 := CalculateLayout(100, 30)
	if l100.LeftWidth != 60 {
		t.Fatalf("expected LeftWidth to be 60 for width 100, got %d", l100.LeftWidth)
	}
	if l100.RightWidth != 36 {
		t.Fatalf("expected RightWidth to be 36 for width 100, got %d", l100.RightWidth)
	}

	// Widescreen 200
	l200 := CalculateLayout(200, 50)
	if l200.LeftWidth != 120 {
		t.Fatalf("expected LeftWidth to be 120 (60%%) for width 200, got %d", l200.LeftWidth)
	}
	if l200.RightWidth != 76 {
		t.Fatalf("expected RightWidth to be 76 for width 200, got %d", l200.RightWidth)
	}

	// Guard too small
	lSmall := CalculateLayout(50, 10)
	if !lSmall.TooSmall {
		t.Fatal("expected TooSmall to be true for 50x10")
	}

	// Explicit ratio
	l70 := CalculateLayoutWithRatio(100, 30, 0.70)
	if l70.LeftWidth != 70 {
		t.Fatalf("expected LeftWidth to be 70 for ratio 0.70, got %d", l70.LeftWidth)
	}
}

func TestSmartTruncatePath(t *testing.T) {
	// Base file exceeds max width
	s1 := SmartTruncatePath("longfilename.md", 8)
	if s1 != "longf..." {
		t.Fatalf("expected 'longf...', got %q", s1)
	}

	// Deeply nested file preserves base filename instead of directory prefix
	path := "bug-reports/bug-01-fix-crash.md"
	s2 := SmartTruncatePath(path, 26)
	if !strings.HasPrefix(s2, ".../") || !strings.HasSuffix(s2, "bug-01-fix-crash.md") {
		t.Fatalf("expected '.../bug-01-fix-crash.md', got %q", s2)
	}

	// Exact length or shorter path
	s3 := SmartTruncatePath("short/path.md", 20)
	if s3 != "short/path.md" {
		t.Fatalf("expected 'short/path.md', got %q", s3)
	}

	// Small max width edge cases
	s4 := SmartTruncatePath("test.md", 0)
	if s4 != "" {
		t.Fatalf("expected empty for maxWidth 0, got %q", s4)
	}
}

func TestListPane_ResponsiveWidth(t *testing.T) {
	files := []model.FileInfo{
		{RelPath: "sub/test.md", Path: "/tmp/test.md"},
	}
	lp := NewListPane(files)

	// Narrow width < 35
	lp.SetDimensions(25, 20)
	if lp.TextInput.Width != 19 {
		t.Fatalf("expected TextInput.Width 19, got %d", lp.TextInput.Width)
	}
	if lp.TextInput.Placeholder != "Search..." {
		t.Fatalf("expected placeholder 'Search...', got %q", lp.TextInput.Placeholder)
	}

	// Medium width 40
	lp.SetDimensions(40, 20)
	if lp.TextInput.Placeholder != "Search files..." {
		t.Fatalf("expected placeholder 'Search files...', got %q", lp.TextInput.Placeholder)
	}

	// Wide width 70
	lp.SetDimensions(70, 20)
	if lp.TextInput.Placeholder != "Type to search files... (Ctrl+F for grep)" {
		t.Fatalf("expected placeholder 'Type to search files... (Ctrl+F for grep)', got %q", lp.TextInput.Placeholder)
	}
}

func TestApp_DefaultSplitRatioOnWindowSizeMsg(t *testing.T) {
	files := []model.FileInfo{
		{RelPath: "file1.md", Path: "/tmp/file1.md"},
	}
	app := NewApp(files, "/tmp")

	// Bubble Tea sends WindowSizeMsg on startup
	updated, _ := app.Update(tea.WindowSizeMsg{Width: 200, Height: 50})
	appRes := updated.(App)

	if appRes.Layout.LeftWidth != 120 {
		t.Fatalf("expected LeftWidth to be 120 (60%% of 200), got %d", appRes.Layout.LeftWidth)
	}
	if appRes.Layout.RightWidth != 76 {
		t.Fatalf("expected RightWidth to be 76, got %d", appRes.Layout.RightWidth)
	}
	if appRes.CustomLeftWidth != 120 {
		t.Fatalf("expected CustomLeftWidth to be 120, got %d", appRes.CustomLeftWidth)
	}
}

func TestApp_MouseClickFocusSwitch(t *testing.T) {
	files := []model.FileInfo{
		{RelPath: "file1.md", Path: "/tmp/file1.md"},
		{RelPath: "file2.md", Path: "/tmp/file2.md"},
	}
	app := NewApp(files, "/tmp")
	app.Focus = FocusList

	dividerX := app.Layout.LeftWidth + 2

	// 1. Click right pane -> FocusPreview
	clickRight := tea.MouseMsg{
		X:      dividerX + 10,
		Y:      5,
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
	}
	upd1, _ := app.Update(clickRight)
	app1 := upd1.(App)
	if app1.Focus != FocusPreview {
		t.Fatalf("expected FocusPreview after clicking right pane, got %v", app1.Focus)
	}

	// 2. Click left pane -> FocusList
	clickLeft := tea.MouseMsg{
		X:      5,
		Y:      2,
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
	}
	upd2, _ := app1.Update(clickLeft)
	app2 := upd2.(App)
	if app2.Focus != FocusList {
		t.Fatalf("expected FocusList after clicking left pane, got %v", app2.Focus)
	}

	// 3. Click second file row in list -> updates cursor to 1
	clickRow2 := tea.MouseMsg{
		X:      5,
		Y:      5, // row 4 is idx 0, row 5 is idx 1
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
	}
	upd3, _ := app2.Update(clickRow2)
	app3 := upd3.(App)
	if app3.List.Cursor != 1 {
		t.Fatalf("expected List.Cursor to be 1 after clicking row 5, got %d", app3.List.Cursor)
	}
}

func TestApp_GrepBackspaceAndTabTraversal(t *testing.T) {
	files := []model.FileInfo{
		{RelPath: "file1.md", Path: "/tmp/file1.md"},
	}
	app := NewApp(files, "/tmp")
	app.Preview.SetContent("Line 1 with keyword\nLine 2 text\nLine 3 another keyword\nLine 4")

	// 1. Start grep search via Ctrl+F
	upd, _ := app.Update(tea.KeyMsg{Type: tea.KeyCtrlF})
	appSearch := upd.(App)
	if !appSearch.Preview.Searching || appSearch.Focus != FocusPreview {
		t.Fatalf("expected Preview.Searching and FocusPreview, got searching=%v, focus=%v", appSearch.Preview.Searching, appSearch.Focus)
	}

	// 2. Type "key"
	for _, ch := range "key" {
		upd, _ = appSearch.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
		appSearch = upd.(App)
	}
	if appSearch.Preview.SearchInput.Value() != "key" {
		t.Fatalf("expected query 'key', got %q", appSearch.Preview.SearchInput.Value())
	}
	if len(appSearch.Preview.MatchLines) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(appSearch.Preview.MatchLines))
	}

	// 3. Press Backspace: must delete character, remain in FocusPreview, remain Searching
	upd, _ = appSearch.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	appSearch = upd.(App)
	if appSearch.Focus != FocusPreview {
		t.Fatalf("expected Focus to stay FocusPreview after Backspace, got %v", appSearch.Focus)
	}
	if !appSearch.Preview.Searching {
		t.Fatal("expected Preview.Searching to remain true after Backspace")
	}
	if appSearch.Preview.SearchInput.Value() != "ke" {
		t.Fatalf("expected query 'ke' after Backspace, got %q", appSearch.Preview.SearchInput.Value())
	}

	// 4. Tab traversal: moves to next match
	initMatch := appSearch.Preview.CurrentMatch
	upd, _ = appSearch.Update(tea.KeyMsg{Type: tea.KeyTab})
	appSearch = upd.(App)
	if appSearch.Preview.CurrentMatch == initMatch && len(appSearch.Preview.MatchLines) > 1 {
		t.Fatalf("expected CurrentMatch to advance after Tab, stayed %d", initMatch)
	}

	// 5. Shift+Tab traversal: moves to previous match
	upd, _ = appSearch.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	appSearch = upd.(App)
	if appSearch.Preview.CurrentMatch != initMatch {
		t.Fatalf("expected CurrentMatch to return to %d after Shift+Tab, got %d", initMatch, appSearch.Preview.CurrentMatch)
	}

	// 6. View line count must equal Height (no layout overflow)
	viewLines := strings.Split(appSearch.Preview.View(), "\n")
	if len(viewLines) != appSearch.Preview.Height {
		t.Fatalf("expected Preview.View() lines to equal Height %d, got %d", appSearch.Preview.Height, len(viewLines))
	}
}

func TestPreviewPane_RenderHighlights(t *testing.T) {
	pp := NewPreviewPane(80, 20)
	raw := "First line with alpha keyword\nSecond line without\nThird line with alpha keyword\nFourth line"
	pp.SetContent(raw)

	// 1. Initially unhighlighted
	if strings.Contains(pp.Viewport.View(), "\x1b[48;5;") {
		t.Fatal("expected no highlight initially")
	}

	// 2. Start search for "alpha"
	pp.StartSearch()
	pp.SearchInput.SetValue("alpha")
	pp.applySearch("alpha")

	if len(pp.MatchLines) != 2 {
		t.Fatalf("expected 2 match lines, got %d", len(pp.MatchLines))
	}

	// Viewport must contain both active highlight (magenta \x1b[48;5;201m) and normal highlight (yellow \x1b[48;5;220m)
	viewContent := pp.Viewport.View()
	if !strings.Contains(viewContent, "\x1b[48;5;201m") {
		t.Fatal("expected active match highlight in magenta")
	}
	if !strings.Contains(viewContent, "\x1b[48;5;220m") {
		t.Fatal("expected secondary match highlight in yellow")
	}

	// 3. Next match switches active line
	pp.NextMatch()
	viewContent2 := pp.Viewport.View()
	if !strings.Contains(viewContent2, "\x1b[48;5;201m") {
		t.Fatal("expected active match highlight after NextMatch")
	}

	// 4. StopSearch restores raw content with zero highlight codes
	pp.StopSearch()
	if strings.Contains(pp.Viewport.View(), "\x1b[48;5;") {
		t.Fatal("expected highlights removed after StopSearch")
	}
}






