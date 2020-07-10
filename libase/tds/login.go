package tds

import (
	"fmt"
	"log"
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
	case TDS_MSG_SEC_ENCRYPT, TDS_MSG_SEC_ENCRYPT2, TDS_MSG_SEC_ENCRYPT3:
		return fmt.Errorf("encryption methods below TDS_MSG_SEC_ENCRYPT3 are not supported by go-ase")
	case TDS_MSG_SEC_ENCRYPT4:
		withoutEncryption = false
	default:
		withoutEncryption = true
	}

	// Add servername/password combination to remote servers
	// The first 'remote' server is the current server with an empty
	// server name.
	firstRemoteServer := LoginConfigRemoteServer{Name: "", Password: config.DSN.Password}
	if len(config.RemoteServers) == 0 {
		config.RemoteServers = []LoginConfigRemoteServer{firstRemoteServer}
	} else {
		config.RemoteServers = append([]LoginConfigRemoteServer{firstRemoteServer}, config.RemoteServers...)
	}

	loginMsg := NewMessage()
	loginMsg.headerType = TDS_BUF_LOGIN

	pack, err := config.pack()
	if err != nil {
		return fmt.Errorf("error building login payload: %w", err)
	}
	loginMsg.AddPackage(pack)

	loginMsg.AddPackage(tdsconn.caps)

	log.Printf("Sending login payload")
	err = tdsconn.Send(*loginMsg)
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

	if negotiationMsg.MsgId != TDS_MSG_SEC_ENCRYPT4 {
		return fmt.Errorf("expected TDS_MSG_SEC_ENCRYPT4, received: %s", negotiationMsg.MsgId)
	}

	params, ok := msg.packages[3].(*ParamsPackage)
	if !ok {
		return fmt.Errorf("expected params package as fourth response, received: %s", msg.packages[3])
	}

	// get asymmetric encryption type
	paramAsymmetricType, ok := params.DataFields[0].(*Int4FieldData)
	if !ok {
		return fmt.Errorf("expected cipher suite as first parameter, got: %#v", params.DataFields[0])
	}
	asymmetricType := uint16(endian.Uint32(paramAsymmetricType.Data()))

	if asymmetricType != 0x0001 {
		return fmt.Errorf("unhandled asymmetric encryption: %b", asymmetricType)
	}

	// get public key
	paramPubKey, ok := params.DataFields[1].(*LongBinaryFieldData)
	if !ok {
		return fmt.Errorf("expected public key as second parameter, got: %#v", params.DataFields[1])
	}

	// get nonce
	paramNonce, ok := params.DataFields[2].(*LongBinaryFieldData)
	if !ok {
		return fmt.Errorf("expected nonce as third parameter, got: %v", params.DataFields[2])
	}

	// encrypt password
	encryptedPass, err := rsaEncrypt(paramPubKey.Data(), paramNonce.Data(), []byte(config.DSN.Password))
	if err != nil {
		return fmt.Errorf("error encrypting password: %w", err)
	}

	// Prepare response
	response := NewMessage()

	response.AddPackage(NewMsgPackage(TDS_MSG_HASARGS, TDS_MSG_SEC_LOGPWD3))

	passFmt, passData, err := LookupFieldFmtData(TDS_LONGBINARY)
	if err != nil {
		return fmt.Errorf("failed to look up fields for TDS_LONGBINARY: %w", err)
	}
	// TDS does not support TDS_WIDETABLES in login negotiation
	response.AddPackage(NewParamFmtPackage(false, passFmt))
	passData.SetData(encryptedPass)
	response.AddPackage(NewParamsPackage(passData))

	if len(config.RemoteServers) > 0 {
		// encrypted remote password
		response.AddPackage(NewMsgPackage(TDS_MSG_HASARGS, TDS_MSG_SEC_REMPWD3))

		paramFmts := make([]FieldFmt, len(config.RemoteServers)*2)
		params := make([]FieldData, len(config.RemoteServers)*2)
		for i := 0; i < len(paramFmts); i += 2 {
			remoteServer := config.RemoteServers[i/2]

			remnameFmt, remnameData, err := LookupFieldFmtData(TDS_VARCHAR)
			if err != nil {
				return fmt.Errorf("failed to look up fields for TDS_VARCHAR: %w", err)
			}

			paramFmts[i] = remnameFmt
			remnameData.SetData([]byte(remoteServer.Name))
			params[i] = remnameData

			encryptedServerPass, err := rsaEncrypt(paramPubKey.Data(), paramNonce.Data(),
				[]byte(remoteServer.Password))
			if err != nil {
				return fmt.Errorf("error encryption remote server password: %w", err)
			}

			passFmt, passData, err := LookupFieldFmtData(TDS_LONGBINARY)
			if err != nil {
				return fmt.Errorf("failed to look up fields for TDS_LONGBINARY")
			}

			paramFmts[i+1] = passFmt
			passData.SetData(encryptedServerPass)
			params[i+1] = passData
		}
		response.AddPackage(NewParamFmtPackage(false, paramFmts...))
		response.AddPackage(NewParamsPackage(params...))
	}

	// TODO verify response

	symmetricKey, err := generateSymmetricKey(aes_256_cbc)
	if err != nil {
		return fmt.Errorf("error generating session key: %w", err)
	}

	encryptedSymKey, err := rsaEncrypt(paramPubKey.Data(), paramNonce.Data(),
		symmetricKey)
	if err != nil {
		return fmt.Errorf("error encrypting session key: %w", err)
	}

	msg = &Message{}

	msg.AddPackage(NewMsgPackage(TDS_MSG_HASARGS, TDS_MSG_SEC_SYMKEY))

	symkeyFmt, symkeyData, err := LookupFieldFmtData(TDS_LONGBINARY)
	if err != nil {
		return fmt.Errorf("failed to look up fields for TDS_LONGBINARY: %w", err)
	}
	symkeyData.SetData(encryptedSymKey)

	msg.AddPackage(NewParamFmtPackage(false, symkeyFmt))
	msg.AddPackage(NewParamsPackage(symkeyData))

	log.Printf("sending response with encrypted passwords")
	err = tdsconn.Send(*msg)
	if err != nil {
		return fmt.Errorf("error sending message with symmetric key: %w", err)
	}

	// TODO handle response

	// TODO generate cipher

	return nil
}
