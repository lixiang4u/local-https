//go:build linux
// +build linux

package helper

func WindowsAdmin() bool {
	return false
}

func UpdateHosts(line string) error {
	return nil
}
