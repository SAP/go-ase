// SPDX-FileCopyrightText: 2020 SAP SE
// SPDX-FileCopyrightText: 2021 SAP SE
// SPDX-FileCopyrightText: 2022 SAP SE
// SPDX-FileCopyrightText: 2023 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package ase

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/SAP/go-dblib/tds"
)

func (stmt Stmt) recvDynAck(ctx context.Context) error {
	_, err := stmt.conn.Channel.NextPackageUntil(ctx, true,
		func(pkg tds.Package) (bool, error) {
			ack, ok := pkg.(*tds.DynamicPackage)
			if !ok {
				return false, nil
			}

			if ack.Type&tds.TDS_DYN_ACK != tds.TDS_DYN_ACK {
				return false, fmt.Errorf("DynamicPackage does not have type TDS_DYN_ACK set: %s", ack)
			}

			return true, nil
		},
	)

	if err != nil && !errors.Is(err, io.EOF) {
		return err
	}

	return nil
}
