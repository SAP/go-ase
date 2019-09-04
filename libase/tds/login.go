package tds

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/SAP/go-ase/libase/libdsn"
)

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
}

func NewLoginConfig(dsn *libdsn.DsnInfo) (*LoginConfig, error) {
	conf := &LoginConfig{}

	conf.DSN = dsn

	hostname, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve hostname: %v", err)
	}
	conf.Hostname = hostname
	conf.HostProc = strconv.Itoa(os.Getpid())

	conf.CharSet = "utf8"
	conf.Language = "us_english"

	conf.PacketSize = 512

	return conf, nil
}

func (tdsconn *TDSConn) Login(config *LoginConfig) error {
	if config == nil {
		return fmt.Errorf("passed config is nil")
	}

	if tdsconn.conn == nil {
		conn, err := Dial("tcp", fmt.Sprintf("%s:%s", config.DSN.Host, config.DSN.Port))
		if err != nil {
			return err
		}
		tdsconn.conn = conn.conn
	}

	pack, err := config.pack()
	if err != nil {
		return fmt.Errorf("error building login payload: %v", err)
	}

	n, packets, err := tdsconn.sendPackage(pack)

	log.Printf("Sent bytes: %d", n)
	log.Printf("Sent packets: %d", packets)
	if err != nil {
		return fmt.Errorf("failed to sent login payload: %v", err)
	}

	return nil
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

	// No error checking requires since bytes.Buffer.Write* methods
	// always return a nil error.

	// lhostname, lhostlen
	err := writeString(buf, config.Hostname, TDS_MAXNAME)
	if err != nil {
		return nil, fmt.Errorf("error writing hostname: %v", err)
	}

	// lusername, lusernlen
	writeString(buf, config.DSN.Username, TDS_MAXNAME)

	// lpw, lpwnlen
	writeString(buf, config.DSN.Password, TDS_MAXNAME)

	// lhostproc, lhplen
	writeString(buf, config.HostProc, TDS_MAXNAME)

	// TODO create consts for parameters

	// lint2 -> big endian
	buf.WriteByte(0x3)
	// lint4 -> big endian
	buf.WriteByte(0x1)
	// lchar -> ASCII
	buf.WriteByte(0x6)
	// lflt -> big endian
	buf.WriteByte(0x10)
	// ldate -> big endian
	buf.WriteByte(0x9)

	// lusedb
	buf.WriteByte(0x0)
	// ldmpld
	buf.WriteByte(0x0)

	// only relevant for server-server comm
	// linterfacespare
	buf.WriteByte(0x0)
	// ltype
	buf.WriteByte(0x0)

	// deprecated
	// lbufsize
	writeString(buf, "", TDS_NETBUF)

	// lspare
	writeString(buf, "", 3)

	// lappname, lappnlen
	writeString(buf, config.AppName, TDS_MAXNAME)

	// lservname, lservnlen
	writeString(buf, config.ServName, TDS_MAXNAME)

	// only relevant for server-server comm
	// lrempw, lrempwlen
	writeString(buf, "", TDS_RPLEN)

	// ltds
	buf.Write([]byte{0x5, 0x0, 0x0, 0x0})

	// lprogname, lprognlen
	writeString(buf, libraryName, TDS_PROGNLEN)

	// lprogvers
	buf.Write([]byte{versionMajor, versionMinor, versionPatch, 0})

	// lnoshort - do not convert short data types
	buf.WriteByte(0x0)

	// lflt4 big endian
	buf.WriteByte(0x13)
	// ldate4 big endian
	buf.WriteByte(0x17)

	// llanguage, llanglen
	writeString(buf, config.Language, TDS_MAXNAME)

	// lsetlang - notify of language changes
	buf.WriteByte(0x1)

	// loldsecure - deprecated
	buf.Write(make([]byte, TDS_OLDSECURE))
	// lseclogin - deprecated
	buf.WriteByte(0x0)
	// lsecbulk - deprecated
	buf.WriteByte(0x1)

	// lhalogin
	// TODO - values need to be determined by config to allow for
	// failover reconnects in clusters
	buf.WriteByte(0x1)
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
	buf.WriteByte(0x1)

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
