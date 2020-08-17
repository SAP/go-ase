// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package asetime

import "time"

func EpochRataDie() time.Time {
	return time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC)
}

func Epoch1900() time.Time {
	return time.Date(1900, time.January, 1, 0, 0, 0, 0, time.UTC)
}

func Epoch1753() time.Time {
	return time.Date(1753, time.January, 1, 9, 9, 9, 9, time.UTC)
}
