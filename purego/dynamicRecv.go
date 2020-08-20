// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package purego

import (
	"context"
	"fmt"

	"github.com/SAP/go-ase/libase/tds"
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

	return err
}

func (stmt Stmt) recvDoneFinal(ctx context.Context) error {
	_, err := stmt.conn.Channel.NextPackageUntil(ctx, true,
		func(pkg tds.Package) (bool, error) {
			done, ok := pkg.(*tds.DonePackage)
			if !ok {
				return false, nil
			}

			if done.Status != tds.TDS_DONE_FINAL {
				return false, fmt.Errorf("DonePackage does not have status TDS_DONE_FINAL set: %s", done)
			}

			return true, nil
		},
	)

	return err
}
