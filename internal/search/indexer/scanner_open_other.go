//go:build !linux && !darwin

package indexer

import (
	"fmt"
	"os"
)

// No unverified blocking fallback for native index scanning on other platforms.
func openScanEntry(_ *os.Root, _ string) (*os.File, error) {
	return nil, fmt.Errorf("safe workspace scanning is supported on Linux and macOS")
}
