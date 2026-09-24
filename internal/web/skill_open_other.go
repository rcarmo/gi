//go:build !linux && !darwin

package web

import (
	"fmt"
	"os"
)

func openWebSkill(_ *os.Root, _ string) (*os.File, error) {
	return nil, fmt.Errorf("safe workspace skill reads are supported on Linux and macOS")
}
