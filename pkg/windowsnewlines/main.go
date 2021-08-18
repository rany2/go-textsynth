package windowsnewlines

import (
	"bytes"

	"github.com/rany2/go-textsynth/pkg/normalizenewlines"
)

// Run normalizes \r and \n into \r\n
func Run(d []byte) []byte {
	// Convert all possible line endings to Unix
	d = normalizenewlines.Run(d)
	// Convert Unix to Windows line endings
	d = bytes.Replace(d, []byte{10}, []byte{13, 10}, -1)
	return d
}
