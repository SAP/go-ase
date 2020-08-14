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
	pkg, err := stmt.conn.nextPackage(ctx)
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
	pkg, err := stmt.conn.nextPackage(ctx)
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