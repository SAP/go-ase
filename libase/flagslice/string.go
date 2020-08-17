// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package flagslice

import "strings"

// FlagStringSlice implements the flags.Value interface.
// Each occurrence of a flag of this type will append the given
// parameter to the flags' value.
type FlagStringSlice []string

// String implements the Stringer interface.
func (fss FlagStringSlice) String() string {
	return strings.Join(fss, " ")
}

// Slice returns the FlagStringSlice as a string slice.
func (fss FlagStringSlice) Slice() []string {
	return ([]string)(fss)
}

// Set appends the given value to the FlagStringSlice.
func (fss *FlagStringSlice) Set(value string) error {
	*fss = append(*fss, value)
	return nil
}
