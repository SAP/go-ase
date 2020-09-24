// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package tds

import (
	"errors"
	"fmt"
)

type EEDError struct {
	EEDPackages  []*EEDPackage
	WrappedError error
}

func (err *EEDError) Add(eed *EEDPackage) {
	err.EEDPackages = append(err.EEDPackages, eed)
}

func (err EEDError) Is(other error) bool {
	if err.WrappedError == nil {
		return false
	}
	return errors.Is(err.WrappedError, other)
}

func (err EEDError) Error() string {
	s := fmt.Sprintf("%s: received EED messages: ", err.WrappedError)

	for _, eed := range err.EEDPackages {
		s += fmt.Sprintf("%d: %s; ", eed.MsgNumber, eed.Msg)
	}

	return s
}
