package flagslice

import "strings"

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
