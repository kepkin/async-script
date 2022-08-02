package async_script

import (
	"bytes"
)

func Replace(what, to string) Op {
	return Transformer(ScanLinesWithoutDrop, replaceTransform([]byte(what), []byte(to)))
}

func replaceTransform(what []byte, to []byte) func([]byte) []byte {
	return func(in []byte) []byte {
		return bytes.Replace(in, what, to, -1)
	}
}

func ScanLinesWithoutDrop(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		// We have a full newline-terminated line.
		return i + 1, data[0 : i+1], nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), data, nil
	}
	// Request more data.
	return 0, nil, nil
}
