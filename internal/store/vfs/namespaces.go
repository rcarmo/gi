package vfs

import "strings"

func IsVirtualNamespace(namespace string) bool {
	switch strings.TrimSpace(namespace) {
	case NamespaceChat, NamespaceReference:
		return true
	default:
		return false
	}
}

func IsReadOnlyNamespace(namespace string) bool {
	return IsVirtualNamespace(strings.TrimSpace(namespace))
}
