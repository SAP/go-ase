package tds

type ErrorPackage struct {
	Length       uint16
	ErrorNumber  int32
	State        uint8
	Class        uint8
	MsgLength    uint16
	ErrorMsg     string
	ServerLength uint8
	ServerName   string
	ProcLength   uint8
	ProcName     string
	LineNr       uint16

	ch   *channel
	err  error
	done bool
}

func (pkg *ErrorPackage) ReadFrom(ch *channel) {
	pkg.ch = ch

	defer func() {
		pkg.done = true
		pkg.ch = nil
	}()

	pkg.Length, pkg.err = ch.Uint16()
	if pkg.err != nil {
		return
	}

	pkg.ErrorNumber, pkg.err = ch.Int32()
	if pkg.err != nil {
		return
	}

	pkg.MsgLength, pkg.err = ch.Uint16()
	if pkg.err != nil {
		return
	}

	pkg.ErrorMsg, pkg.err = ch.String(int(pkg.MsgLength))
	if pkg.err != nil {
		return
	}

	pkg.ServerLength, pkg.err = ch.Uint8()
	if pkg.err != nil {
		return
	}

	pkg.ServerName, pkg.err = ch.String(int(pkg.ServerLength))
	if pkg.err != nil {
		return
	}

	pkg.ProcLength, pkg.err = ch.Uint8()
	if pkg.err != nil {
		return
	}

	pkg.ProcName, pkg.err = ch.String(int(pkg.ProcLength))
	if pkg.err != nil {
		return
	}

	pkg.LineNr, pkg.err = ch.Uint16()
	if pkg.err != nil {
		return
	}
}

func (pkg ErrorPackage) Error() error {
	if pkg.err == ErrChannelExhausted {
		return nil
	}

	return pkg.err
}

func (pkg ErrorPackage) Finished() bool {
	return pkg.done
}

func (pkg ErrorPackage) String() string {
	return pkg.ErrorMsg
}

func (pkg ErrorPackage) Packets() chan Packet {
	return nil
}
