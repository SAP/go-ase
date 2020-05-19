package tds

type DataFieldStatus uint8

const (
	noStatus               DataFieldStatus = 0
	TDS_PARAM_RETURN                       = 1
	TDS_PARAM_COLUMNSTATUS                 = 8
	TDS_PARAM_NULLALLOWED                  = 20
)
