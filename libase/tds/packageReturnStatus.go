// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package tds

import "fmt"

var _ Package = (*ReturnStatusPackage)(nil)

type ReturnStatusPackage struct {
	returnValue int32
}

func (pkg *ReturnStatusPackage) ReadFrom(ch BytesChannel) error {
	var err error

	pkg.returnValue, err = ch.Int32()
	if err != nil {
		return ErrNotEnoughBytes
	}

	return nil
}

func (pkg ReturnStatusPackage) WriteTo(ch BytesChannel) error {
	return ch.WriteInt32(pkg.returnValue)
}

func (pkg ReturnStatusPackage) ReturnValue() int {
	return int(pkg.returnValue)
}

func (pkg ReturnStatusPackage) String() string {
	return fmt.Sprintf("%T(%d)", pkg, pkg.returnValue)
}
