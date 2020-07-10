package tds

// EnvChangeHook defines the signature of functions called by a TDSConn
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
func (tdsconn *TDSConn) RegisterEnvChangeHook(fn EnvChangeHook) {
	tdsconn.envChangeHooksLock.Lock()
	defer tdsconn.envChangeHooksLock.Unlock()

	tdsconn.envChangeHooks = append(tdsconn.envChangeHooks, fn)
}

func (tdsconn *TDSConn) callEnvChangeHooks(typ EnvChangeType, oldValue, newValue string) {
	tdsconn.envChangeHooksLock.Lock()
	defer tdsconn.envChangeHooksLock.Unlock()

	for _, fn := range tdsconn.envChangeHooks {
		fn(typ, oldValue, newValue)
	}
}