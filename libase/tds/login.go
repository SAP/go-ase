package tds

import (
	"fmt"
)

func (tdsChan *TDSChannel) Login(config *LoginConfig) error {
	if config == nil {
		return fmt.Errorf("passed config is nil")
	}

	var withoutEncryption bool
	switch config.Encrypt {
	case TDS_MSG_SEC_ENCRYPT, TDS_MSG_SEC_ENCRYPT2, TDS_MSG_SEC_ENCRYPT3:
		return fmt.Errorf("encryption methods below TDS_MSG_SEC_ENCRYPT4 are not supported by go-ase")
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

	tdsChan.currentHeaderType = TDS_BUF_LOGIN

	pack, err := config.pack()
	if err != nil {
		return fmt.Errorf("error building login payload: %w", err)
	}

	err = tdsChan.AddPackage(pack)
	if err != nil {
		return fmt.Errorf("error adding login payload package: %w", err)
	}

	err = tdsChan.AddPackage(tdsChan.tdsConn.caps)
	if err != nil {
		return fmt.Errorf("error adding login capabilities package: %w", err)
	}

	pkg, err := tdsChan.NextPackage(true)
	if err != nil {
		return err
	}

	loginack, ok := pkg.(*LoginAckPackage)
	if !ok {
		return fmt.Errorf("expected LoginAck as first response, received: %v", pkg)
	}

	if withoutEncryption {
		// no encryption requested, check loginack for validity and
		// return
		if loginack.Status != TDS_LOG_SUCCEED {
			return fmt.Errorf("login failed: %s", loginack.Status)
		}

		pkg, err = tdsChan.NextPackage(true)
		if err != nil {
			return err
		}

		done, ok := pkg.(*DonePackage)
		if !ok {
			return fmt.Errorf("expected Done as second response, received: %v", pkg)
		}

		if done.status != TDS_DONE_FINAL {
			return fmt.Errorf("expected DONE(FINAL), received: %s", done)
		}

		return nil
	}

	if loginack.Status != TDS_LOG_NEGOTIATE {
		return fmt.Errorf("expected loginack with negotation, received: %s", loginack)
	}

	pkg, err = tdsChan.NextPackage(true)
	if err != nil {
		return err
	}

	negotiationMsg, ok := pkg.(*MsgPackage)
	if !ok {
		return fmt.Errorf("expected msg package as second response, received: %s", pkg)
	}

	if negotiationMsg.MsgId != TDS_MSG_SEC_ENCRYPT4 {
		return fmt.Errorf("expected TDS_MSG_SEC_ENCRYPT4, received: %s", negotiationMsg.MsgId)
	}

	pkg, err = tdsChan.NextPackage(true)
	if err != nil {
		return err
	}

	params, ok := pkg.(*ParamsPackage)
	if !ok {
		return fmt.Errorf("expected params package as fourth response, received: %s", pkg)
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
	tdsChan.AddPackage(NewMsgPackage(TDS_MSG_HASARGS, TDS_MSG_SEC_LOGPWD3))
	if err != nil {
		return fmt.Errorf("error adding message package for password transmission: %w", err)
	}

	passFmt, passData, err := LookupFieldFmtData(TDS_LONGBINARY)
	if err != nil {
		return fmt.Errorf("failed to look up fields for TDS_LONGBINARY: %w", err)
	}

	// TDS does not support TDS_WIDETABLES in login negotiation
	tdsChan.AddPackage(NewParamFmtPackage(false, passFmt))
	if err != nil {
		return fmt.Errorf("error adding ParamFmt password package: %w", err)
	}

	passData.SetData(encryptedPass)
	tdsChan.AddPackage(NewParamsPackage(passData))
	if err != nil {
		return fmt.Errorf("error adding Params password package: %w", err)
	}

	if len(config.RemoteServers) > 0 {
		// encrypted remote password
		tdsChan.AddPackage(NewMsgPackage(TDS_MSG_HASARGS, TDS_MSG_SEC_REMPWD3))
		if err != nil {
			return fmt.Errorf("error adding message package for remote servers: %w", err)
		}

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
		tdsChan.AddPackage(NewParamFmtPackage(false, paramFmts...))
		if err != nil {
			return fmt.Errorf("error adding package ParamFmt for remote servers: %w", err)
		}

		tdsChan.AddPackage(NewParamsPackage(params...))
		if err != nil {
			return fmt.Errorf("error adding package Params for remote servers: %w", err)
		}
	}

	symmetricKey, err := generateSymmetricKey(tdsChan.tdsConn.odce)
	if err != nil {
		return fmt.Errorf("error generating session key: %w", err)
	}

	encryptedSymKey, err := rsaEncrypt(paramPubKey.Data(), paramNonce.Data(),
		symmetricKey)
	if err != nil {
		return fmt.Errorf("error encrypting session key: %w", err)
	}

	tdsChan.AddPackage(NewMsgPackage(TDS_MSG_HASARGS, TDS_MSG_SEC_SYMKEY))

	symkeyFmt, symkeyData, err := LookupFieldFmtData(TDS_LONGBINARY)
	if err != nil {
		return fmt.Errorf("failed to look up fields for TDS_LONGBINARY: %w", err)
	}
	symkeyData.SetData(encryptedSymKey)

	tdsChan.AddPackage(NewParamFmtPackage(false, symkeyFmt))
	tdsChan.AddPackage(NewParamsPackage(symkeyData))

	// Filter received packages to single out the login-relevant ones.
	relevantPackages := []Package{}
	for pkg, err := tdsChan.NextPackage(false); err != nil; pkg, err = tdsChan.NextPackage(false) {
		switch pkg.(type) {
		case *EnvChangePackage, *EEDPackage:
			// these packages are handled in .Receive, ignore
			continue
		default:
			relevantPackages = append(relevantPackages, pkg)
		}
	}

	loginAck, ok := relevantPackages[0].(*LoginAckPackage)
	if !ok {
		return fmt.Errorf("expected login ack package, received %T instead: %v",
			relevantPackages[0], relevantPackages[0])
	}

	if loginAck.Status != TDS_LOG_SUCCEED {
		return fmt.Errorf("expected login ack with status TDS_LOG_SUCCEED, received %s",
			loginAck.Status)
	}

	_, ok = relevantPackages[1].(*CapabilityPackage)
	if !ok {
		return fmt.Errorf("expected capability package, received %T instead: %v",
			relevantPackages[1], relevantPackages[1])
	}

	// TODO handle caps response

	done, ok := relevantPackages[2].(*DonePackage)
	if !ok {
		return fmt.Errorf("expected done package, received %T instead: %v",
			relevantPackages[2], relevantPackages[2])
	}

	if done.status != TDS_DONE_FINAL {
		return fmt.Errorf("expected done package with status TDS_DONE_FINAL, received %s",
			done.status)
	}

	if done.tranState != TDS_TRAN_COMPLETED {
		return fmt.Errorf("expected done package with transtate TDS_TRAN_COMPLETED, received %s",
			done.tranState)
	}

	return nil
}
