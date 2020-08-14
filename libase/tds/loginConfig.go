package tds

import (
	"bytes"
	"fmt"
	"os"
	"strconv"

	"github.com/SAP/go-ase/libase/libdsn"
)

type LoginConfigRemoteServer struct {
	Name, Password string
}

type LoginConfig struct {
	DSN      *libdsn.Info
	Hostname string

	// TODO name
	HostProc string
	AppName  string
	ServName string

	Language string
	CharSet  string

	RemoteServers []LoginConfigRemoteServer

	// Encrypt allows any TDSMsgId but only negotation-relevant security
	// bits such as TDS_MSG_SEC_ENCRYPT will be recognized.
	Encrypt TDSMsgId
}

func NewLoginConfig(dsn *libdsn.Info) (*LoginConfig, error) {
	conf := &LoginConfig{}

	conf.DSN = dsn

	hostname, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve hostname: %w", err)
	}
	conf.Hostname = hostname
	conf.HostProc = strconv.Itoa(os.Getpid())

	conf.ServName = conf.DSN.Host
	// Should be overwritten by clients
	conf.AppName = "github.com/SAP/go-ase/libase/tds"

	conf.CharSet = "utf8"
	conf.Language = "us_english"

	conf.Encrypt = TDS_MSG_SEC_ENCRYPT4

	return conf, nil
}

// TODO lower-casing
const (
	TDS_MAXNAME   = 30
	TDS_NETBUF    = 4
	TDS_RPLEN     = 255
	TDS_VERSIZE   = 4
	TDS_PROGNLEN  = 10
	TDS_OLDSECURE = 2
	TDS_HA        = 6
	TDS_SECURE    = 2
	TDS_PKTLEN    = 6
	TDS_DUMMY     = 4
)

func (config *LoginConfig) pack() (Package, error) {
	buf := &bytes.Buffer{}

	// lhostname, lhostlen
	err := writeString(buf, config.Hostname, TDS_MAXNAME)
	if err != nil {
		return nil, fmt.Errorf("error writing hostname: %w", err)
	}

	// lusername, lusernlen
	err = writeString(buf, config.DSN.Username, TDS_MAXNAME)
	if err != nil {
		return nil, fmt.Errorf("error writing username: %w", err)
	}

	// lpw, lpwnlen
	switch config.Encrypt {
	case TDS_MSG_SEC_ENCRYPT, TDS_MSG_SEC_ENCRYPT2, TDS_MSG_SEC_ENCRYPT3, TDS_MSG_SEC_ENCRYPT4:
		err = writeString(buf, "", TDS_MAXNAME)
	default:
		err = writeString(buf, config.DSN.Password, TDS_MAXNAME)
	}
	if err != nil {
		return nil, fmt.Errorf("error writing password: %w", err)
	}

	// lhostproc, lhplen
	err = writeString(buf, config.HostProc, TDS_MAXNAME)
	if err != nil {
		return nil, fmt.Errorf("error writing hostproc: %w", err)
	}

	// lint2
	_, err = writeBasedOnEndian(buf, 3, 2)
	if err != nil {
		return nil, fmt.Errorf("error writing int2: %w", err)
	}

	// lint4
	_, err = writeBasedOnEndian(buf, 1, 0)
	if err != nil {
		return nil, fmt.Errorf("error writing int4: %w", err)
	}

	// lchar -> ASCII
	err = buf.WriteByte(6)
	if err != nil {
		return nil, fmt.Errorf("error writing char: %w", err)
	}

	// lflt
	_, err = writeBasedOnEndian(buf, 10, 4)
	if err != nil {
		return nil, fmt.Errorf("error writing flt: %w", err)
	}

	// ldate
	_, err = writeBasedOnEndian(buf, 9, 8)
	if err != nil {
		return nil, fmt.Errorf("error writing date: %w", err)
	}

	// lusedb
	err = buf.WriteByte(1)
	if err != nil {
		return nil, fmt.Errorf("error writing usedb: %w", err)
	}

	// ldmpld
	err = buf.WriteByte(1)
	if err != nil {
		return nil, fmt.Errorf("error writing dmpld: %w", err)
	}

	// only relevant for server-server comm
	// linterfacespare
	err = buf.WriteByte(0)
	if err != nil {
		return nil, fmt.Errorf("error writing interfacespare: %w", err)
	}

	// ltype
	err = buf.WriteByte(0)
	if err != nil {
		return nil, fmt.Errorf("error writing type: %w", err)
	}

	// deprecated
	// lbufsize
	_, err = buf.Write(make([]byte, TDS_NETBUF))
	if err != nil {
		return nil, fmt.Errorf("error writing bufsize: %w", err)
	}

	// lspare
	_, err = buf.Write(make([]byte, 3))
	if err != nil {
		return nil, fmt.Errorf("error writing spare: %w", err)
	}

	// lappname, lappnlen
	err = writeString(buf, config.AppName, TDS_MAXNAME)
	if err != nil {
		return nil, fmt.Errorf("error writing appname: %w", err)
	}

	// lservname, lservnlen
	err = writeString(buf, config.ServName, TDS_MAXNAME)
	if err != nil {
		return nil, fmt.Errorf("error writing servname: %w", err)
	}

	// TODO only relevant for server-server comm, replace?
	// lrempw, lrempwlen
	switch config.Encrypt {
	case TDS_MSG_SEC_ENCRYPT, TDS_MSG_SEC_ENCRYPT2, TDS_MSG_SEC_ENCRYPT3, TDS_MSG_SEC_ENCRYPT4:
		err = writeString(buf, "", TDS_RPLEN)
	default:
		err = writeString(buf, "", TDS_RPLEN)
	}
	if err != nil {
		return nil, fmt.Errorf("error writing rempw: %w", err)
	}

	// ltds
	_, err = buf.Write([]byte{0x5, 0x0, 0x0, 0x0})
	if err != nil {
		return nil, fmt.Errorf("error writing tds version: %w", err)
	}

	// lprogname, lprognlen
	err = writeString(buf, libraryName, TDS_PROGNLEN)
	if err != nil {
		return nil, fmt.Errorf("error writing progname: %w", err)
	}

	// lprogvers
	_, err = buf.Write(libraryVersion.Bytes())
	if err != nil {
		return nil, fmt.Errorf("error writing progversion: %w", err)
	}

	// lnoshort - do not convert short data types
	err = buf.WriteByte(0)
	if err != nil {
		return nil, fmt.Errorf("error writing noshort: %w", err)
	}

	// lflt4
	_, err = writeBasedOnEndian(buf, 13, 12)
	if err != nil {
		return nil, fmt.Errorf("error writing flt4: %w", err)
	}

	// ldate4
	_, err = writeBasedOnEndian(buf, 17, 16)
	if err != nil {
		return nil, fmt.Errorf("error writing date4: %w", err)
	}

	// llanguage, llanglen
	err = writeString(buf, config.Language, TDS_MAXNAME)
	if err != nil {
		return nil, fmt.Errorf("error writing language: %w", err)
	}

	// lsetlang - notify of language changes
	err = buf.WriteByte(1)
	if err != nil {
		return nil, fmt.Errorf("error writing setlang: %w", err)
	}

	// loldsecure - deprecated
	_, err = buf.Write(make([]byte, TDS_OLDSECURE))
	if err != nil {
		return nil, fmt.Errorf("error writing oldsecure: %w", err)
	}

	// lseclogin
	switch config.Encrypt {
	case TDS_MSG_SEC_ENCRYPT:
		err = buf.WriteByte(0x01)
	case TDS_MSG_SEC_ENCRYPT2:
		err = buf.WriteByte(0x1 | 0x20)
	case TDS_MSG_SEC_ENCRYPT3, TDS_MSG_SEC_ENCRYPT4:
		err = buf.WriteByte(0x1 | 0x20 | 0x80)
	default:
		err = buf.WriteByte(0x0)
	}
	if err != nil {
		return nil, fmt.Errorf("error writing seclogin: %w", err)
	}

	// lsecbulk - deprecated
	err = buf.WriteByte(1)
	if err != nil {
		return nil, fmt.Errorf("error writing secbulk: %w", err)
	}

	// lhalogin
	// TODO - values need to be determined by config to allow for
	// failover reconnects in clusters
	err = buf.WriteByte(1)
	if err != nil {
		return nil, fmt.Errorf("error writing ailover: %w", err)
	}

	// lhasessionid
	// TODO session id for HA failover, find out if this needs to be
	// user set or retrieved from the server
	_, err = buf.Write(make([]byte, TDS_HA))
	if err != nil {
		return nil, fmt.Errorf("error writing hasessionid: %w", err)
	}

	// lsecspare - unused
	// TODO TDS_SECURE unknown
	_, err = buf.Write(make([]byte, TDS_SECURE))
	if err != nil {
		return nil, fmt.Errorf("error writing secspare: %w", err)
	}

	// lcharset, lcharsetlen
	err = writeString(buf, config.CharSet, TDS_MAXNAME)
	if err != nil {
		return nil, fmt.Errorf("error writing charset: %w", err)
	}

	// lsetcharset - notify of charset changes
	err = buf.WriteByte(1)
	if err != nil {
		return nil, fmt.Errorf("error writing setcharset: %w", err)
	}

	// lpacketsize - 256 to 65535 bytes
	// Write default packet size - will be renegotiated anyhow.
	err = writeString(buf, "512", TDS_PKTLEN)
	if err != nil {
		return nil, fmt.Errorf("error writing packetsize: %w", err)
	}

	// ldummy - apparently unused
	_, err = buf.Write(make([]byte, TDS_DUMMY))
	if err != nil {
		return nil, fmt.Errorf("error writing dummy: %w", err)
	}

	return &TokenlessPackage{Data: buf}, nil
}
