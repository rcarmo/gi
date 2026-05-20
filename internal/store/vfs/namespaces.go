package vfs

import "strings"

func IsVirtualNamespace(namespace string) bool {
	switch strings.TrimSpace(namespace) {
	case NamespaceChat:
		return true
	default:
		return false
	}
}

func IsReadOnlyNamespace(namespace string) bool {
	ns := strings.TrimSpace(namespace)
	if ns == NamespaceReference {
		return true
	}
	return IsVirtualNamespace(ns)
}
