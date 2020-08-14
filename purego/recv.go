package purego

import (
	"context"
	"fmt"

	"github.com/SAP/go-ase/libase/tds"
)

func (conn Conn) nextPackage(ctx context.Context) (tds.Package, error) {
	pkg, err := conn.Channel.NextPackage(ctx, true)
	if err != nil {
		return nil, fmt.Errorf("error receiving response: %w", err)
	}

	if eed, ok := pkg.(*tds.EEDPackage); ok {
		// Received an EEDPackage, execution failed

		// Consume any remaining packages in queue until DonePackage
		conn.Channel.NextPackageUntil(ctx, true, func(pkg tds.Package) bool {
			done, ok := pkg.(*tds.DonePackage)
			if !ok {
				return false
			}

			// If this is a Done package check that there aren't
			// more packages to catch before the transmission is
			// done.
			if done.Status&tds.TDS_DONE_MORE == tds.TDS_DONE_MORE {
				return false
			}
			return true
		})

		return nil, fmt.Errorf("server reported error: %s", eed.Msg)
	}

	return pkg, nil
}
