package tds

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/binary"
	"encoding/pem"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/SAP/go-ase/libase/libdsn"
)

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

	var withoutEncryption bool
	switch config.Encrypt {
	case TDS_SEC_LOG_ENCRYPT, TDS_SEC_LOG_ENCRYPT2, TDS_SEC_LOG_ENCRYPT3:
		withoutEncryption = false
	default:
		withoutEncryption = true
	}

	// Add servername/password combination to remote servers
	firstRemoteServer := LoginConfigRemoteServer{Name: config.DSN.Host, Password: config.DSN.Password}
	if len(config.RemoteServers) == 0 {
		config.RemoteServers = []LoginConfigRemoteServer{firstRemoteServer}
	} else {
		config.RemoteServers = append([]LoginConfigRemoteServer{firstRemoteServer}, config.RemoteServers...)
	}

	loginMsg := Message{}
	loginMsg.headerType = TDS_BUF_LOGIN

	pack, err := config.pack()
	if err != nil {
		return fmt.Errorf("error building login payload: %v", err)
	}
	loginMsg.AddPackage(pack)

	capPack := NewCapabilityPackage()
	capPack.SetRequestCapability(TDS_DATA_COLUMNSTATUS, true)
	loginMsg.AddPackage(capPack)

	log.Printf("Sending login payload")
	err = tdsconn.Send(loginMsg)
	if err != nil {
		return fmt.Errorf("failed to sent login payload: %v", err)
	}

	log.Printf("Reading response")
	msg, err := tdsconn.Receive()
	if err != nil {
		return fmt.Errorf("error reading login response: %w", err)
	}

	log.Printf("Handling response")
	if IsError(msg.packages[0]) {
		return fmt.Errorf("received error from server: %s", msg.packages[0])
	}

	loginack, ok := msg.packages[0].(*LoginAckPackage)
	if !ok {
		return fmt.Errorf("expected LoginAck as first response, received: %v", msg.packages[0])
	}

	if withoutEncryption {
		// no encryption requested, check loginack for validity and
		// return
		if loginack.Status != TDS_LOG_SUCCEED {
			return fmt.Errorf("login failed: %s", loginack.Status)
		}

		done, ok := msg.packages[1].(*DonePackage)
		if !ok {
			return fmt.Errorf("expected Done as second response, received: %v", msg.packages[1])
		}

		if done.status != TDS_DONE_FINAL {
			return fmt.Errorf("expected DONE(FINAL), received: %s", done)
		}

		return nil
	}

	if loginack.Status != TDS_LOG_NEGOTIATE {
		return fmt.Errorf("expected loginack with negotation, received: %s", loginack)
	}

	params, ok := msg.packages[3].(*ParamsPackage)
	if !ok {
		return fmt.Errorf("expected params package as fourth response, received: %s", msg.packages[3])
	}

	// Handle the different encryption types
	switch config.Encrypt {
	case TDS_SEC_LOG_ENCRYPT3:
		// get cipher suite
		paramCipherSuite, ok := params.Params[0].(*Int4FieldData)
		if !ok {
			return fmt.Errorf("expected cipher suite as first parameter, got: %#v", params.Params[0])
		}
		cipherSuite := binary.BigEndian.Uint32(paramCipherSuite.Data())

		// get public key
		paramPubKey, ok := params.Params[1].(*LongBinaryFieldData)
		if !ok {
			return fmt.Errorf("expected public key as second parameter, got: %#v", params.Params[1])
		}

		block, rest := pem.Decode(paramPubKey.Data())
		if len(rest) > 0 {
			return fmt.Errorf("trailing bytes in public key: %#v", rest)
		}

		publicKey, err := x509.ParsePKCS1PublicKey(block.Bytes)
		if err != nil {
			return fmt.Errorf("error parsing PKCS1 public key: %w", err)
		}

		// get nonce
		paramNonce, ok := params.Params[2].(*LongBinaryFieldData)
		if !ok {
			return fmt.Errorf("expected nonce as third parameter, got: %v", params.Params[2])
		}
		nonce := paramNonce.Data()

		fmt.Printf("cipher suite: %v\n", cipherSuite)
		fmt.Printf("pubkey: %v\n", publicKey)
		fmt.Printf("nonce: %v\n", nonce)

		encryptedPass, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey, append(nonce, []byte(config.DSN.Password)...))
		if err != nil {
			return fmt.Errorf("error encrypting password: %w", err)
		}
		fmt.Printf("encryptedPass: %v\n", encryptedPass)

		// Prepare response
		response := NewMessage()

		// encrypted login password
		response.AddPackage(NewMsgPackage(TDS_MSG_HASARGS, TDS_MSG_SEC_LOGPWD3))

		pwdLongBinaryFmt, err := LookupFieldFmt(TDS_LONGBINARY)
		if err != nil {
			return fmt.Errorf("could not lookup Fieldfmt for TDS_LONGBINARY: %w", err)
		}
		pwdLongBinaryFmt.SetLength(len(encryptedPass))
		response.AddPackage(NewParamFmtPackage(pwdLongBinaryFmt))

		pwdLongBinaryData, err := LookupFieldData(pwdLongBinaryFmt)
		if err != nil {
			return fmt.Errorf("could not lookup FieldData for TDS_LONGBINARY: %w", err)
		}
		pwdLongBinaryData.SetData(encryptedPass)
		response.AddPackage(NewParamsPackage(pwdLongBinaryData))

		// encrypted remote password
		if len(config.RemoteServers) > 0 {
			response.AddPackage(NewMsgPackage(TDS_MSG_HASARGS, TDS_MSG_SEC_REMPWD3))

			paramFmts := make([]FieldFmt, len(config.RemoteServers)*2)
			params := make([]FieldData, len(config.RemoteServers)*2)
			for i, remoteServer := range config.RemoteServers {
				varCharFmt, err := LookupFieldFmt(TDS_VARCHAR)
				if err != nil {
					return err
				}
				varCharFmt.SetLength(len(remoteServer.Name))
				paramFmts[i] = varCharFmt

				varCharData, err := LookupFieldData(varCharFmt)
				if err != nil {
					return err
				}
				varCharData.SetData([]byte(remoteServer.Name))
				params[i] = varCharData

				encryptedServerPass, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey,
					append(nonce, []byte(remoteServer.Password)...))
				if err != nil {
					return err
				}

				longBinaryFmt, err := LookupFieldFmt(TDS_LONGBINARY)
				if err != nil {
					return err
				}
				longBinaryFmt.SetLength(len(encryptedServerPass))
				paramFmts[i+1] = longBinaryFmt

				longBinaryData, err := LookupFieldData(longBinaryFmt)
				if err != nil {
					return err
				}
				longBinaryData.SetData(encryptedServerPass)
				params[i+1] = longBinaryData
			}
			response.AddPackage(NewParamFmtPackage(paramFmts...))
			response.AddPackage(NewParamsPackage(params...))
		}

		log.Printf("sending response with encrypted passwords")
		err = tdsconn.Send(*response)
		if err != nil {
			return fmt.Errorf("error sending response with encrypted passwords: %w", err)
		}
	default:
		return fmt.Errorf("unhandled: %v", config.Encrypt)
	}

	return nil
}

//go:generate stringer -type=LoginSec
type LoginSec uint8

const (
	TDS_SEC_LOG_ENCRYPT    LoginSec = 0x01
	TDS_SEC_LOG_CHALLENGE           = 0x02
	TDS_SEC_LOG_LABELS              = 0x04
	TDS_SEC_LOG_APPDEFINED          = 0x08
	TDS_SEC_LOG_SECSESS             = 0x10
	TDS_SEC_LOG_ENCRYPT2            = 0x20
	TDS_SEC_LOG_ENCRYPT3            = 0x80
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

	Encrypt LoginSec
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

	conf.AppName = fmt.Sprintf("%s:%s", conf.DSN.Host, conf.DSN.Port)

	conf.CharSet = "utf8"
	conf.Language = "us_english"

	conf.PacketSize = 512

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
	switch config.Encrypt {
	case TDS_SEC_LOG_ENCRYPT, TDS_SEC_LOG_ENCRYPT2, TDS_SEC_LOG_ENCRYPT3:
		writeString(buf, "", TDS_MAXNAME)
	default:
		writeString(buf, config.DSN.Password, TDS_MAXNAME)
	}

	// lhostproc, lhplen
	writeString(buf, config.HostProc, TDS_MAXNAME)

	// TODO create consts for parameters

	// lint2 -> big endian
	buf.WriteByte(0x2)
	// lint4 -> big endian
	buf.WriteByte(0x0)
	// lchar -> ASCII
	buf.WriteByte(0x6)
	// lflt -> big endian
	buf.WriteByte(0x4)
	// ldate -> big endian
	buf.WriteByte(0x8)

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
	case TDS_SEC_LOG_ENCRYPT, TDS_SEC_LOG_ENCRYPT2, TDS_SEC_LOG_ENCRYPT3:
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
	// ver, err := NewTDSVersionString(libraryVersion)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to create TDSVersion from libraryVersion: %w", err)
	// }
	// buf.Write(ver.Bytes())

	// lnoshort - do not convert short data types
	buf.WriteByte(0x0)

	// lflt4 big endian
	buf.WriteByte(12)
	// ldate4 big endian
	buf.WriteByte(16)

	// llanguage, llanglen
	writeString(buf, config.Language, TDS_MAXNAME)

	// lsetlang - notify of language changes
	buf.WriteByte(0x1)

	// loldsecure - deprecated
	buf.Write(make([]byte, TDS_OLDSECURE))

	// lseclogin
	switch config.Encrypt {
	case TDS_SEC_LOG_ENCRYPT:
		buf.WriteByte(byte(TDS_SEC_LOG_ENCRYPT))
	case TDS_SEC_LOG_ENCRYPT2:
		buf.WriteByte(byte(TDS_SEC_LOG_ENCRYPT2))
	case TDS_SEC_LOG_ENCRYPT3:
		buf.WriteByte(byte(TDS_SEC_LOG_ENCRYPT3))
	default:
		buf.WriteByte(0x0)
	}

	// // lseclogin
	// switch config.Encrypt {
	// case TDS_SEC_LOG_ENCRYPT:
	// 	buf.WriteByte(byte(TDS_SEC_LOG_ENCRYPT))
	// case TDS_SEC_LOG_ENCRYPT2:
	// 	buf.WriteByte(byte(TDS_SEC_LOG_ENCRYPT | TDS_SEC_LOG_ENCRYPT2))
	// case TDS_SEC_LOG_ENCRYPT3:
	// 	buf.WriteByte(byte(TDS_SEC_LOG_ENCRYPT | TDS_SEC_LOG_ENCRYPT2 | TDS_SEC_LOG_ENCRYPT3))
	// default:
	// 	buf.WriteByte(0x0)
	// }

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
