// Code generated by "stringer -type=LanguageStatus"; DO NOT EDIT.

package tds

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[TDS_LANGUAGE_NOARGS-0]
	_ = x[TDS_LANGUAGE_HASARGS-1]
	_ = x[TDS_LANG_BATCH_PARAMS-4]
}

const (
	_LanguageStatus_name_0 = "TDS_LANGUAGE_NOARGSTDS_LANGUAGE_HASARGS"
	_LanguageStatus_name_1 = "TDS_LANG_BATCH_PARAMS"
)

var (
	_LanguageStatus_index_0 = [...]uint8{0, 19, 39}
)

func (i LanguageStatus) String() string {
	switch {
	case 0 <= i && i <= 1:
		return _LanguageStatus_name_0[_LanguageStatus_index_0[i]:_LanguageStatus_index_0[i+1]]
	case i == 4:
		return _LanguageStatus_name_1
	default:
		return "LanguageStatus(" + strconv.FormatInt(int64(i), 10) + ")"
	}
}
