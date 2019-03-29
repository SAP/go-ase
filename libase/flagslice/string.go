package flagslice

import "strings"

// FlagStringSlice implements the flags.Value interface.
// Each occurrence of a flag of this type will append the given
// parameter to the flags' value.
type FlagStringSlice []string

func (fss FlagStringSlice) String() string {
	return strings.Join(fss, " ")
}

func (fss FlagStringSlice) Slice() []string {
	var slice []string
	slice = fss
	return slice
}

func (fss *FlagStringSlice) Set(value string) error {
	*fss = append(*fss, value)
	return nil
}
