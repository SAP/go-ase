package capability

var (
	ExampleVersionCapOtherthing   = NewCapability("otherthing", "1.5.0")
	ExampleVersionBug50           = NewCapability("bug reported in ticket #50", "0.1.0", "0.2.0", "0.9.0", "1.1.0")
	ExampleVersionCapSpeedyAction = NewCapability("cap1", "1.0.0")

	ExampleVersionTarget = Target{nil, []*Capability{
		ExampleVersionCapOtherthing,
		ExampleVersionBug50,
		ExampleVersionCapSpeedyAction,
	}}
)

// ExampleSatisfyInterface shows how a struct could be modelled to satisfy the
// Version interface.
//
// The example struct models a struct storing information to connect to
// and communicate with a remote server.
type ExampleSatisfyInterface struct {
	Dsn string
	// Other members

	// ServerVersion stores the version of the server after connecting
	// to it.
	ServerVersion string
	// caps stores the available capabilities of the server
	caps map[*Capability]bool
}

// Connect connects to the server, stores the server version in the
// struct and calls the targets SetVersion
func (v *ExampleSatisfyInterface) Connect() {
	v.ServerVersion = "1.0.5"

	// Set capabilities based on server version
	ExampleVersionTarget.SetVersion(v)
}

func (v ExampleSatisfyInterface) VersionString() string {
	return v.ServerVersion
}

func (v ExampleSatisfyInterface) SetCapability(cap *Capability, set bool) {
	v.caps[cap] = set
}

// Do emulates any action on the remote server.
func (v ExampleSatisfyInterface) Do(s string) string {
	return s
}

func (v ExampleSatisfyInterface) Has(cap *Capability) bool {
	can, ok := v.caps[cap]
	if !ok {
		return false
	}
	return can
}

func ExampleVersion() {
	// Create connection to server and connect
	connection := &ExampleSatisfyInterface{}
	connection.Connect()

	// Perform actions against server
	connection.Do("something")

	// The action 'otherthing' has only been implemented from version
	// 1.5.0 but should be used by the application if the server
	// supports it.
	// By guarding the call with a call to .Has the capability can be
	// supported without restricting users to specific versions of the
	// application.
	if connection.Has(ExampleVersionCapOtherthing) {
		connection.Do("otherthing")
	}

	// Ticket #50 reported a serious bug that would cause the remote
	// server to shutdown on specific actions.
	// The bug appeared in version 0.1.0, a fix was implemented in
	// 0.2.0. The bug reappeared in 0.9.0 and has been fixed again in
	// 1.1.0.
	// Assuming a workaround has been found it can be applied outside of
	// these version ranges:
	if connection.Has(ExampleVersionBug50) {
		// Workaround for bug
	}
	connection.Do("action that would break the server without the bugfix or workaround")

	// The application should execute an action on the server that takes
	// a long time to finish. In version 1.0.0 a new action performing
	// the same task in less time was added. The capabilities can be
	// used to support both pre-1.0.0 and post-1.0.0 servers:
	if connection.Has(ExampleVersionCapSpeedyAction) {
		connection.Do("speedy action")
	} else {
		connection.Do("slow action")
	}
}

type ExampleEmbed struct {
	DefaultVersion
	Dsn string
}

type ExampleMember struct {
	Version DefaultVersion
	Dsn     string
}
