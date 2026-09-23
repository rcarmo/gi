//go:build !linux && !darwin

package tools

import "os"

func openWorkspaceWrite(root *os.Root, name string) (*os.File, error) {
	return root.OpenFile(name, os.O_WRONLY|os.O_CREATE, 0644)
}
