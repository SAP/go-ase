package tds

import "fmt"

// EnvChangeHook defines the signature of functions called by a Conn
// when the server sends a TDS_ENV_CHANGE package.
type EnvChangeHook func(typ EnvChangeType, oldValue, newValue string)

// RegisterEnvChangeHook registers a function to be called when the TDS
// server sends a TDS_ENVCHNAGE token.
//
// The registered functions are called with the EnvChangeType of the
// update, the old value and the new value. The old value may be empty.
//
// The functions should be registered _before_ calling `.Login` as env
// changes happen at the end of the login negotiation.
//
// Note that all registered hooks are called in sequence of being
// registered. Hooks with a longer run time or waiting on locks should
// utilize goroutines or use other means to prevent blocking other
// hooks.
func (tdsChan *Channel) RegisterEnvChangeHooks(fns ...EnvChangeHook) error {
	tdsChan.envChangeHooksLock.Lock()
	defer tdsChan.envChangeHooksLock.Unlock()

	for i, fn := range fns {
		if fn == nil {
			return fmt.Errorf("received nil function as hook at index %d", i)
		}
	}

	tdsChan.envChangeHooks = append(tdsChan.envChangeHooks, fns...)
	return nil
}

func (tdsChan *Channel) callEnvChangeHooks(typ EnvChangeType, oldValue, newValue string) {
	tdsChan.envChangeHooksLock.Lock()
	defer tdsChan.envChangeHooksLock.Unlock()

	for _, fn := range tdsChan.envChangeHooks {
		fn(typ, oldValue, newValue)
	}
}
