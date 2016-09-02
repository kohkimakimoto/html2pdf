// +build !windows

package html2pdf

import (
	"syscall"
)

func Umask(newmask int) int {
	return syscall.Umask(newmask)
}
