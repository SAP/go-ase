package tds

func IsError(pkg Package) bool {
	switch pkg.(type) {
	case *EEDPackage, *ErrorPackage:
		return true
	}

	return false
}
