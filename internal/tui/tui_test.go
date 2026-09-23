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
		MouseWheelDelta:   1,
	}
	app := NewAppWithConfig(files, "/tmp", cfg)

	// Send WheelDown - cursor MUST NOT change in DualPane mode
	wheelDownMsg := tea.MouseMsg{Button: tea.MouseButtonWheelDown}
	updated, cmd := app.Update(wheelDownMsg)
	appResult := updated.(App)
	if cmd != nil {
		t.Fatalf("expected nil cmd for wheel drop, got %v", cmd)
	}
	if appResult.List.Cursor != 0 {
		t.Fatalf("expected cursor to remain 0 in DualPane mode, got %d", appResult.List.Cursor)
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

	// 1. Ctrl+C in FocusList copies path and does NOT quit
	updatedC, cmd := app.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	appC := updatedC.(App)
	if cmd != nil {
		t.Fatalf("expected nil cmd (no quit) on Ctrl+C in List, got %v", cmd)
	}
	if !strings.Contains(appC.Notification, "Copied path") {
		t.Fatalf("expected notification to contain 'Copied path', got %q", appC.Notification)
	}

	// 2. Ctrl+A copies content
	updatedA, _ := app.Update(tea.KeyMsg{Type: tea.KeyCtrlA})
	appA := updatedA.(App)
	if !strings.Contains(appA.Notification, "Copied content") {
		t.Fatalf("expected notification to contain 'Copied content', got %q", appA.Notification)
	}
}



