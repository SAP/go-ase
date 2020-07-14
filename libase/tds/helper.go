package tds

import (
	"io"
)

// writeString writes s padded to padTo and its length to buf.
func writeString(stream io.Writer, s string, padTo int) error {
	_, err := stream.Write([]byte(s))
	if err != nil {
		return err
	}

	_, err = stream.Write(make([]byte, padTo-len(s)))
	if err != nil {
		return err
	}

	_, err = stream.Write([]byte{byte(len(s))})
	return err
}

func deBitmask(bitmask int, maxValue int) []int {
	curVal := 1
	ret := []int{}

	for curVal <= maxValue {
		if bitmask&curVal == curVal {
			ret = append(ret, curVal)
		}
		curVal = curVal << 1
	}

	return ret
}
