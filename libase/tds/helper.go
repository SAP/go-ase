// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package tds

import (
	"fmt"
	"io"
)

// writeString writes s padded to padTo and its length to buf.
func writeString(stream io.Writer, s string, padTo int) error {
	if len(s) > padTo {
		return fmt.Errorf("string '%s' is too large, must be at most %d bytes long",
			s, padTo)
	}

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
