// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package purego

import (
	"context"
	"fmt"

	"github.com/SAP/go-ase/libase/tds"
)

// nextPackage handles the fact that a TDS server may return an EED at
// any time.
func (stmt Stmt) nextPackage(ctx context.Context) (tds.Package, error) {
	pkg, err := stmt.conn.Channel.NextPackage(ctx, true)
	if err != nil {
		return nil, fmt.Errorf("error receiving response: %w", err)
	}

	if eed, ok := pkg.(*tds.EEDPackage); ok {
		// Received an EEDPackage, dynamic statement allocation failed

		// Consume any remaining packages in queue until DonePackage
		stmt.conn.Channel.NextPackageUntil(ctx, true, func(pkg tds.Package) bool {
			_, ok := pkg.(*tds.DonePackage)
			return ok
		})

		return nil, fmt.Errorf("server reported error during dynamic statement creation: %s", eed)
	}

	return pkg, nil
}

func (stmt Stmt) recvDynAck(ctx context.Context) error {
	pkg, err := stmt.nextPackage(ctx)
	if err != nil {
		return err
	}

	ack, ok := pkg.(*tds.DynamicPackage)
	if !ok {
		return fmt.Errorf("expected a DynamicPackage, got: %v", pkg)
	}

	if ack.Type&tds.TDS_DYN_ACK != tds.TDS_DYN_ACK {
		return fmt.Errorf("DynamicPackage does not have type TDS_DYN_ACK set: %s", ack)
	}

	return nil
}

func (stmt Stmt) recvDoneFinal(ctx context.Context) error {
	pkg, err := stmt.nextPackage(ctx)
	if err != nil {
		return err
	}

	done, ok := pkg.(*tds.DonePackage)
	if !ok {
		return fmt.Errorf("expected a DonePackage, got: %v", pkg)
	}

	if done.Status&tds.TDS_DONE_FINAL != tds.TDS_DONE_FINAL {
		return fmt.Errorf("DonePackage does not have status TDS_DONE_FINAL set: %s", done)
	}

	return nil
}
