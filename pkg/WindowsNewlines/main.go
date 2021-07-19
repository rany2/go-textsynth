package WindowsNewlines

import (
	"bytes"
	. "github.com/rany2/go-textsynth/pkg/NormalizeNewlines"
)

// WindowsNewlines normalizes \r and \n into \r\n
func WindowsNewlines(d []byte) []byte {
	// Convert all possible line endings to Unix
	d = NormalizeNewlines(d)
	// Convert Unix to Windows line endings
	d = bytes.Replace(d, []byte{10}, []byte{13, 10}, -1)
	return d
}
