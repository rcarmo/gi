package vfs

import "strings"

func IsVirtualNamespace(namespace string) bool {
	switch strings.TrimSpace(namespace) {
	case "chat":
		return true
	default:
		return false
	}
}

func IsReadOnlyNamespace(namespace string) bool {
	ns := strings.TrimSpace(namespace)
	if ns == "reference" {
		return true
	}
	return IsVirtualNamespace(ns)
}
