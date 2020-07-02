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
	DSN      *libdsn.DsnInfo
	Hostname string

	// TODO name
	HostProc string
	AppName  string
	ServName string

	Language string
	CharSet  string

	PacketSize uint16

	RemoteServers []LoginConfigRemoteServer

	// Encrypt allows any TDSMsgId but only negotation-relevant security
	// bits such as TDS_MSG_SEC_ENCRYPT will be recognized.
	Encrypt TDSMsgId
}

func NewLoginConfig(dsn *libdsn.DsnInfo) (*LoginConfig, error) {
	conf := &LoginConfig{}

	conf.DSN = dsn

	hostname, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve hostname: %w", err)
	}
	conf.Hostname = hostname
	conf.HostProc = strconv.Itoa(os.Getpid())

	conf.AppName = fmt.Sprintf("%s:%s", conf.DSN.Host, conf.DSN.Port)

	conf.CharSet = "utf8"
	conf.Language = "us_english"

	conf.PacketSize = 512
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

	// No error checking required since bytes.Buffer.Write* methods
	// always return a nil error.

	// lhostname, lhostlen
	err := writeString(buf, config.Hostname, TDS_MAXNAME)
	if err != nil {
		return nil, fmt.Errorf("error writing hostname: %w", err)
	}

	// lusername, lusernlen
	writeString(buf, config.DSN.Username, TDS_MAXNAME)

	// lpw, lpwnlen
	switch config.Encrypt {
	case TDS_MSG_SEC_ENCRYPT, TDS_MSG_SEC_ENCRYPT2, TDS_MSG_SEC_ENCRYPT3, TDS_MSG_SEC_ENCRYPT4:
		writeString(buf, "", TDS_MAXNAME)
	default:
		writeString(buf, config.DSN.Password, TDS_MAXNAME)
	}

	// lhostproc, lhplen
	writeString(buf, config.HostProc, TDS_MAXNAME)

	// lint2
	writeBasedOnEndian(buf, 3, 2)
	// lint4
	writeBasedOnEndian(buf, 1, 0)
	// lchar -> ASCII
	buf.WriteByte(6)
	// lflt
	writeBasedOnEndian(buf, 10, 4)
	// ldate
	writeBasedOnEndian(buf, 9, 8)

	// lusedb
	buf.WriteByte(1)
	// ldmpld
	buf.WriteByte(1)

	// only relevant for server-server comm
	// linterfacespare
	buf.WriteByte(0)
	// ltype
	buf.WriteByte(0)

	// deprecated
	// lbufsize
	buf.Write(make([]byte, TDS_NETBUF))

	// lspare
	buf.Write(make([]byte, 3))

	// lappname, lappnlen
	writeString(buf, config.AppName, TDS_MAXNAME)

	// lservname, lservnlen
	writeString(buf, config.ServName, TDS_MAXNAME)

	// TODO only relevant for server-server comm, replace?
	// lrempw, lrempwlen
	switch config.Encrypt {
	case TDS_MSG_SEC_ENCRYPT, TDS_MSG_SEC_ENCRYPT2, TDS_MSG_SEC_ENCRYPT3, TDS_MSG_SEC_ENCRYPT4:
		writeString(buf, "", TDS_RPLEN)
	default:
		writeString(buf, "", TDS_RPLEN)
	}

	// ltds
	buf.Write([]byte{0x5, 0x0, 0x0, 0x0})

	// lprogname, lprognlen
	writeString(buf, libraryName, TDS_PROGNLEN)

	// lprogvers
	// buf.Write([]byte{versionMajor, versionMinor, versionPatch, 0})
	buf.Write([]byte{0, 1, 0, 0})
	// TODO write correct version
	// ver, err := NewTDSVersionString(libraryVersion)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to create TDSVersion from libraryVersion: %w", err)
	// }
	// buf.Write(ver.Bytes())

	// lnoshort - do not convert short data types
	buf.WriteByte(0)

	// lflt4
	writeBasedOnEndian(buf, 13, 12)
	// ldate4
	writeBasedOnEndian(buf, 17, 16)

	// llanguage, llanglen
	writeString(buf, config.Language, TDS_MAXNAME)

	// lsetlang - notify of language changes
	buf.WriteByte(1)

	// loldsecure - deprecated
	buf.Write(make([]byte, TDS_OLDSECURE))

	// lseclogin
	switch config.Encrypt {
	case TDS_MSG_SEC_ENCRYPT:
		buf.WriteByte(0x01)
	case TDS_MSG_SEC_ENCRYPT2:
		buf.WriteByte(0x1 | 0x20)
	case TDS_MSG_SEC_ENCRYPT3, TDS_MSG_SEC_ENCRYPT4:
		// TODO is this also correct for encrypt4?
		buf.WriteByte(0x1 | 0x20 | 0x80)
	default:
		buf.WriteByte(0x0)
	}

	// lsecbulk - deprecated
	buf.WriteByte(1)

	// lhalogin
	// TODO - values need to be determined by config to allow for
	// failover reconnects in clusters
	buf.WriteByte(1)
	// lhasessionid
	// TODO session id for HA failover, find out if this needs to be
	// user set or retrieved from the server
	buf.Write(make([]byte, TDS_HA))

	// lsecspare - unused
	// TODO TDS_SECURE unknown
	buf.Write(make([]byte, TDS_SECURE))

	// lcharset, lcharsetlen
	writeString(buf, config.CharSet, TDS_MAXNAME)

	// lsetcharset - notify of charset changes
	buf.WriteByte(1)

	// lpacketsize - 256 to 65535 bytes
	// TODO Choose default packet size
	if config.PacketSize < 256 {
		return nil, fmt.Errorf("packet size too low, must be at least 256 bytes")
	}
	writeString(buf, strconv.Itoa(int(config.PacketSize)), TDS_PKTLEN)

	// ldummy - apparently unused
	buf.Write(make([]byte, TDS_DUMMY))

	return &TokenlessPackage{Data: buf}, nil
}
