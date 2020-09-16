// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package tds

import "fmt"

// EEDHook defines the signature of functions called by a Conn when the
// server sends an EEDPackage.
type EEDHook func(eedPackage EEDPackage)

// RegisterEEDHooks registers a function to be called when the TDS
// server sends an EEDPackage.
//
// The registered functions are called with the full EEDPackage.
//
// Note that all registered hooks are called in sequence of being
// registered. Hooks with a longer run time or waiting on locks should
// utilize goroutines or use other means to prevent blocking other
// hooks.
func (tdsChan *Channel) RegisterEEDHooks(fns ...EEDHook) error {
	tdsChan.envChangeHooksLock.Lock()
	defer tdsChan.envChangeHooksLock.Unlock()

	for i, fn := range fns {
		if fn == nil {
			return fmt.Errorf("tds: received nil function as hook at index %d", i)
		}
	}

	tdsChan.eedHooks = append(tdsChan.eedHooks, fns...)
	return nil
}

func (tdsChan *Channel) callEEDHooks(eed EEDPackage) {
	tdsChan.envChangeHooksLock.Lock()
	defer tdsChan.envChangeHooksLock.Unlock()

	for _, fn := range tdsChan.eedHooks {
		fn(eed)
	}
}
