// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package tds

import (
	"context"
	"errors"
	"fmt"

	"github.com/SAP/go-ase/libase/types"
)

func (tdsChan *Channel) Login(ctx context.Context, config *LoginConfig) error {
	if config == nil {
		return errors.New("passed config is nil")
	}

	tdsChan.CurrentHeaderType = TDS_BUF_LOGIN

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

	pack, err := config.pack()
	if err != nil {
		return fmt.Errorf("error building login payload: %w", err)
	}

	err = tdsChan.QueuePackage(ctx, pack)
	if err != nil {
		return fmt.Errorf("error adding login payload package: %w", err)
	}

	err = tdsChan.QueuePackage(ctx, tdsChan.tdsConn.Caps)
	if err != nil {
		return fmt.Errorf("error adding login capabilities package: %w", err)
	}

	err = tdsChan.SendRemainingPackets(ctx)
	if err != nil {
		return fmt.Errorf("error sending packets: %w", err)
	}

	pkg, err := tdsChan.NextPackage(ctx, true)
	if err != nil {
		return fmt.Errorf("error reading LoginAck package: %w", err)
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

		pkg, err = tdsChan.NextPackage(ctx, true)
		if err != nil {
			return fmt.Errorf("error reading Done package: %w", err)
		}

		done, ok := pkg.(*DonePackage)
		if !ok {
			return fmt.Errorf("expected Done as second response, received: %v", pkg)
		}

		if done.Status&TDS_DONE_FINAL != TDS_DONE_FINAL {
			return fmt.Errorf("expected DONE(FINAL), received: %s", done)
		}

		return nil
	}

	if loginack.Status != TDS_LOG_NEGOTIATE {
		return fmt.Errorf("expected loginack with negotation, received: %s", loginack)
	}

	pkg, err = tdsChan.NextPackage(ctx, true)
	if err != nil {
		return fmt.Errorf("error reading Msg package: %w", err)
	}

	negotiationMsg, ok := pkg.(*MsgPackage)
	if !ok {
		return fmt.Errorf("expected msg package as second response, received: %s", pkg)
	}

	if negotiationMsg.MsgId != TDS_MSG_SEC_ENCRYPT4 {
		return fmt.Errorf("expected TDS_MSG_SEC_ENCRYPT4, received: %s", negotiationMsg.MsgId)
	}

	pkg, err = tdsChan.NextPackage(ctx, true)
	if err != nil {
		return fmt.Errorf("error reading ParamFmt package: %w", err)
	}

	paramFmt, ok := pkg.(*ParamFmtPackage)
	if !ok {
		return fmt.Errorf("expected paramfmt package as third response, recevied: %v", pkg)
	}

	if len(paramFmt.Fmts) != 3 {
		return fmt.Errorf("invalid paramfmt package, expected 3 fields, got %d: %v",
			len(paramFmt.Fmts), paramFmt)
	}

	pkg, err = tdsChan.NextPackage(ctx, true)
	if err != nil {
		return fmt.Errorf("error reading Params package: %w", err)
	}

	params, ok := pkg.(*ParamsPackage)
	if !ok {
		return fmt.Errorf("expected params package as fourth response, received: %s", pkg)
	}

	if len(params.DataFields) != 3 {
		return fmt.Errorf("invalid params package, expected 3 fields, got %d: %v",
			len(params.DataFields), params)
	}

	pkg, err = tdsChan.NextPackage(ctx, true)
	if err != nil {
		return fmt.Errorf("error reading Done package: %w", err)
	}

	_, ok = pkg.(*DonePackage)
	if !ok {
		return fmt.Errorf("expected done package as fifth response, received: %v", pkg)
	}

	// get asymmetric encryption type
	paramAsymmetricType, ok := params.DataFields[0].(*Int4FieldData)
	if !ok {
		return fmt.Errorf("expected cipher suite as first parameter, got: %#v", params.DataFields[0])
	}

	asymmetricType, ok := paramAsymmetricType.Value().(int32)
	if !ok {
		return fmt.Errorf("param field for asymmetric type contains value of type %T instead of int32",
			paramAsymmetricType.Value())
	}

	if asymmetricType != 1 {
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
	paramPubKeyData, ok := paramPubKey.Value().([]byte)
	if !ok {
		return fmt.Errorf("param field for public key contains value of type %T instead of []byte",
			paramPubKey.Value())
	}

	paramNonceData, ok := paramNonce.Value().([]byte)
	if !ok {
		return fmt.Errorf("param field for nonce contains value of type %T instead of []byte",
			paramNonce.Value())
	}

	encryptedPass, err := rsaEncrypt(paramPubKeyData, paramNonceData, []byte(config.DSN.Password))
	if err != nil {
		return fmt.Errorf("error encrypting password: %w", err)
	}

	// Prepare response
	err = tdsChan.QueuePackage(ctx, NewMsgPackage(TDS_MSG_HASARGS, TDS_MSG_SEC_LOGPWD3))
	if err != nil {
		return fmt.Errorf("error queueing message package for password transmission: %w", err)
	}

	passFmt, passData, err := LookupFieldFmtData(types.LONGBINARY)
	if err != nil {
		return fmt.Errorf("failed to look up fields for LONGBINARY: %w", err)
	}

	// TDS does not support TDS_WIDETABLES in login negotiation
	err = tdsChan.QueuePackage(ctx, NewParamFmtPackage(false, passFmt))
	if err != nil {
		return fmt.Errorf("error queueing ParamFmt password package: %w", err)
	}

	passData.SetValue(encryptedPass)
	err = tdsChan.QueuePackage(ctx, NewParamsPackage(passData))
	if err != nil {
		return fmt.Errorf("error queueing Params password package: %w", err)
	}

	if len(config.RemoteServers) > 0 {
		// encrypted remote password
		err = tdsChan.QueuePackage(ctx, NewMsgPackage(TDS_MSG_HASARGS, TDS_MSG_SEC_REMPWD3))
		if err != nil {
			return fmt.Errorf("error queueing message package for remote servers: %w", err)
		}

		paramFmts := make([]FieldFmt, len(config.RemoteServers)*2)
		params := make([]FieldData, len(config.RemoteServers)*2)
		for i := 0; i < len(paramFmts); i += 2 {
			remoteServer := config.RemoteServers[i/2]

			remnameFmt, remnameData, err := LookupFieldFmtData(types.VARCHAR)
			if err != nil {
				return fmt.Errorf("failed to look up fields for VARCHAR: %w", err)
			}

			paramFmts[i] = remnameFmt
			remnameData.SetValue([]byte(remoteServer.Name))
			params[i] = remnameData

			encryptedServerPass, err := rsaEncrypt(paramPubKeyData, paramNonceData,
				[]byte(remoteServer.Password))
			if err != nil {
				return fmt.Errorf("error encryption remote server password: %w", err)
			}

			passFmt, passData, err := LookupFieldFmtData(types.LONGBINARY)
			if err != nil {
				return fmt.Errorf("failed to look up fields for LONGBINARY")
			}

			paramFmts[i+1] = passFmt
			passData.SetValue(encryptedServerPass)
			params[i+1] = passData
		}

		err = tdsChan.QueuePackage(ctx, NewParamFmtPackage(false, paramFmts...))
		if err != nil {
			return fmt.Errorf("error queueing package ParamFmt for remote servers: %w", err)
		}

		err = tdsChan.QueuePackage(ctx, NewParamsPackage(params...))
		if err != nil {
			return fmt.Errorf("error queueing package Params for remote servers: %w", err)
		}
	}

	symmetricKey, err := generateSymmetricKey(tdsChan.tdsConn.odce)
	if err != nil {
		return fmt.Errorf("error generating session key: %w", err)
	}

	encryptedSymKey, err := rsaEncrypt(paramPubKeyData, paramNonceData, symmetricKey)
	if err != nil {
		return fmt.Errorf("error encrypting session key: %w", err)
	}

	err = tdsChan.QueuePackage(ctx, NewMsgPackage(TDS_MSG_HASARGS, TDS_MSG_SEC_SYMKEY))
	if err != nil {
		return fmt.Errorf("error queueing package Msg for symmetric key: %w", err)
	}

	symkeyFmt, symkeyData, err := LookupFieldFmtData(types.LONGBINARY)
	if err != nil {
		return fmt.Errorf("failed to look up fields for LONGBINARY: %w", err)
	}
	symkeyData.SetValue(encryptedSymKey)

	err = tdsChan.QueuePackage(ctx, NewParamFmtPackage(false, symkeyFmt))
	if err != nil {
		return fmt.Errorf("error queueing package ParamFmt for symmetric key: %w", err)
	}

	err = tdsChan.QueuePackage(ctx, NewParamsPackage(symkeyData))
	if err != nil {
		return fmt.Errorf("error queueing package Params for symmetric key: %w", err)
	}

	err = tdsChan.SendRemainingPackets(ctx)
	if err != nil {
		return fmt.Errorf("error sending login payload: %w", err)
	}

	pkg, err = tdsChan.NextPackageUntil(ctx, true,
		func(pkg Package) (bool, error) {
			loginAck, ok := pkg.(*LoginAckPackage)
			if !ok {
				return false, nil
			}

			if loginAck.Status != TDS_LOG_SUCCEED {
				return false, fmt.Errorf("expected login ack with status TDS_LOG_SUCCEED, received %s",
					loginAck.Status)
			}

			return true, nil
		},
	)
	if err != nil {
		return fmt.Errorf("error reading LoginAck package: %w", err)
	}

	pkg, err = tdsChan.NextPackage(ctx, true)
	if err != nil {
		return fmt.Errorf("error reading Capability package: %w", err)
	}

	capsResponse, ok := pkg.(*CapabilityPackage)
	if !ok {
		return fmt.Errorf("expected capability package, received %T instead: %v", pkg, pkg)
	}

	for capType, capTypeCaps := range capsResponse.Capabilities {
		// Skip over capability types that aren't requested
		if len(capTypeCaps.capabilities) == 1 {
			continue
		}

		// Check if all caps have been zeroed - this means the server
		// didn't understand the capability requests at all
		allZeroed := true
		for _, bit := range capTypeCaps.capabilities {
			if bit {
				allZeroed = false
			}
		}

		if allZeroed {
			return fmt.Errorf("server did not understand capability requests for %s, aborting", capType)
		}
	}

	// Override requested capabilities with server response
	tdsChan.tdsConn.Caps = capsResponse

	pkg, err = tdsChan.NextPackage(ctx, true)
	if err != nil {
		return fmt.Errorf("error reading Done package: %w", err)
	}

	done, ok := pkg.(*DonePackage)
	if !ok {
		return fmt.Errorf("expected done package, received %T instead: %v", pkg, pkg)
	}

	if done.Status&TDS_DONE_FINAL != TDS_DONE_FINAL {
		return fmt.Errorf("expected done package with status TDS_DONE_FINAL, received %s",
			done.Status)
	}

	if done.TranState&TDS_TRAN_COMPLETED != TDS_TRAN_COMPLETED {
		return fmt.Errorf("expected done package with transtate TDS_TRAN_COMPLETED, received %s",
			done.TranState)
	}

	tdsChan.Reset()

	return nil
}
