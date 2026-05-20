package vfs

import (
	"fmt"
	"path"
	"strings"
)

func ParsePath(raw string) (namespace, pathPart string, err error) {
	trimmed := strings.TrimSpace(raw)
	if !strings.HasPrefix(trimmed, "vfs://") {
		return "", "", fmt.Errorf("invalid vfs path: must start with vfs://")
	}
	trimmed = strings.TrimPrefix(trimmed, "vfs://")
	trimmed = strings.TrimLeft(trimmed, "/")
	if trimmed == "" {
		return "", "", fmt.Errorf("invalid vfs path: missing namespace")
	}
	parts := strings.SplitN(trimmed, "/", 2)
	namespace = strings.TrimSpace(parts[0])
	if namespace == "" {
		return "", "", fmt.Errorf("invalid vfs path: missing namespace")
	}
	if len(parts) == 1 {
		return namespace, "", nil
	}
	normalizedPath, err := NormalizePath(parts[1])
	if err != nil {
		return "", "", err
	}
	return namespace, normalizedPath, nil
}

func NormalizePath(rawPath string) (string, error) {
	clean := path.Clean(strings.TrimLeft(strings.TrimSpace(rawPath), "/"))
	if clean == "." {
		return "", nil
	}
	if strings.HasPrefix(clean, "../") || clean == ".." {
		return "", fmt.Errorf("invalid vfs path: traversal outside namespace")
	}
	return clean, nil
}
