//go:build !linux

package Process

func bindToParent(parent int) bool {
	return false
}
