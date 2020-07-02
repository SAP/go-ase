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
		return fmt.Errorf("connection has not been dialed")
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
		return fmt.Errorf("error building login payload: %w", err)
	}
	loginMsg.AddPackage(pack)

	loginMsg.AddPackage(tdsconn.caps)

	log.Printf("Sending login payload")
	err = tdsconn.Send(loginMsg)
	if err != nil {
		return fmt.Errorf("failed to sent login payload: %w", err)
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

	negotiationMsg, ok := msg.packages[1].(*MsgPackage)
	if !ok {
		return fmt.Errorf("expected msg package as second response, received: %s", msg.packages[1])
	}

	if negotiationMsg.MsgId != TDS_MSG_SEC_ENCRYPT3 {
		return fmt.Errorf("expected TDS_MSG_SEC_ENCRYPT3, received: %s", negotiationMsg.MsgId)
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
		cipherSuite := binary.BigEndian.Uint16(paramCipherSuite.Data()[2:])

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
		response.AddPackage(NewParamFmtPackage(
			tdsconn.caps.HasRequestCapability(TDS_WIDETABLES),
			pwdLongBinaryFmt))

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
			response.AddPackage(NewParamFmtPackage(
				tdsconn.caps.HasRequestCapability(TDS_WIDETABLES),
				paramFmts...))
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
