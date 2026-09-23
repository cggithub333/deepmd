package tui

import (
	"encoding/base64"
	"fmt"
	"os"

	"github.com/atotto/clipboard"
)

// CopyToClipboard copies text to the system clipboard using atotto/clipboard (wl-copy, xclip)
// and emits OSC 52 ANSI escape sequences as a universal fallback for terminal emulators.
func CopyToClipboard(text string) error {
	_ = clipboard.WriteAll(text)

	// Emit OSC 52 terminal clipboard escape sequence
	b64 := base64.StdEncoding.EncodeToString([]byte(text))
	if os.Getenv("TMUX") != "" {
		fmt.Printf("\x1bPtmux;\x1b\x1b]52;c;%s\x07\x1b\\", b64)
	} else {
		fmt.Printf("\x1b]52;c;%s\x07", b64)
	}
	return nil
}
