/*
 * Copyright (C) 2018 The cntmology Authors
 * This file is part of The cntmology library.
 *
 * The cntmology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The cntmology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * alcntm with The cntmology.  If not, see <http://www.gnu.org/licenses/>.
 */

package auth

import (
	"bytes"
	"fmt"
	"time"

	"github.com/cntmio/cntmology/common/serialization"
	"github.com/cntmio/cntmology/errors"
	"github.com/cntmio/cntmology/smartccntmract/event"
	"github.com/cntmio/cntmology/smartccntmract/service/native"
	"github.com/cntmio/cntmology/smartccntmract/service/native/utils"
)

var (
	future = time.Date(2100, 1, 1, 12, 0, 0, 0, time.UTC)
)

func Init() {
	native.Ccntmracts[utils.AuthCcntmractAddress] = RegisterAuthCcntmract
}

/*
 * ccntmract admin management
 */
func initCcntmractAdmin(native *native.NativeService, ccntmractAddr, cntmID []byte) (bool, error) {
	admin, err := getCcntmractAdmin(native, ccntmractAddr)
	if err != nil {
		return false, err
	}
	if admin != nil {
		//admin is already set, just return
		return false, nil
	}
	err = putCcntmractAdmin(native, ccntmractAddr, cntmID)
	if err != nil {
		return false, err
	}
	return true, nil
}

func InitCcntmractAdmin(native *native.NativeService) ([]byte, error) {
	param := new(InitCcntmractAdminParam)
	rd := bytes.NewReader(native.Input)
	if err := param.Deserialize(rd); err != nil {
		return nil, err
	}
	cxt := native.CcntmextRef.CallingCcntmext()
	if cxt == nil {
		return nil, errors.NewErr("no calling ccntmext")
	}
	invokeAddr := cxt.CcntmractAddress

	ret, err := initCcntmractAdmin(native, invokeAddr[:], param.AdminOntID)
	if err != nil {
		return nil, err
	}
	if !ret {
		return utils.BYTE_FALSE, nil
	}
	pushEvent(native, []interface{}{"initCcntmractAdmin", param.AdminOntID})
	return utils.BYTE_TRUE, nil
}

func transfer(native *native.NativeService, ccntmractAddr, newAdminOntID []byte, keyNo uint64) (bool, error) {
	admin, err := getCcntmractAdmin(native, ccntmractAddr)
	if err != nil {
		return false, err
	}
	if admin == nil {
		return false, nil
	}

	ret, err := verifySig(native, admin, keyNo)
	if err != nil {
		return false, err
	}
	if !ret {
		return false, nil
	}

	adminKey, err := concatCcntmractAdminKey(native, ccntmractAddr)
	if err != nil {
		return false, err
	}
	utils.PutBytes(native, adminKey, newAdminOntID)
	return true, nil
}

func Transfer(native *native.NativeService) ([]byte, error) {
	param := new(TransferParam)
	rd := bytes.NewReader(native.Input)
	err := param.Deserialize(rd)
	if err != nil {
		return nil, err
	}
	ret, err := transfer(native, param.CcntmractAddr, param.NewAdminOntID, param.KeyNo)
	if ret {
		event := new(event.NotifyEventInfo)
		event.CcntmractAddress = native.CcntmextRef.CurrentCcntmext().CcntmractAddress
		event.States = []interface{}{"transfer", param.NewAdminOntID}
		native.Notifications = append(native.Notifications, event)
		return utils.BYTE_TRUE, nil
	} else {
		return utils.BYTE_FALSE, nil
	}
}

func AssignFuncsToRole(native *native.NativeService) ([]byte, error) {
	//deserialize input param
	param := new(FuncsToRoleParam)
	rd := bytes.NewReader(native.Input)
	if err := param.Deserialize(rd); err != nil {
		return nil, fmt.Errorf("deserialize param failed, caused by %v", err)
	}
	if param.Role == nil {
		invokeEvent(native, "assignFuncsToRole", false)
		return utils.BYTE_FALSE, nil
	}

	//check the caller's permission
	admin, err := getCcntmractAdmin(native, param.CcntmractAddr)
	if err != nil {
		return nil, fmt.Errorf("get ccntmract admin failed, caused by %v", err)
	}
	if admin == nil { //admin has not been set
		invokeEvent(native, "assignFuncsToRole", false)
		return utils.BYTE_FALSE, nil
	}
	if bytes.Compare(admin, param.AdminOntID) != 0 {
		invokeEvent(native, "assignFuncsToRole", false)
		return utils.BYTE_FALSE, nil
	}
	ret, err := verifySig(native, param.AdminOntID, param.KeyNo)
	if err != nil {
		return nil, fmt.Errorf("verify admin's signature failed, caused by %v", err)
	}
	if !ret {
		invokeEvent(native, "assignFuncsToRole", false)
		return utils.BYTE_FALSE, nil
	}

	funcs, err := getRoleFunc(native, param.CcntmractAddr, param.Role)
	if funcs != nil {
		funcNames := append(funcs.funcNames, param.FuncNames...)
		funcs.funcNames = stringSliceUniq(funcNames)
	} else {
		funcs = new(roleFuncs)
		funcs.funcNames = stringSliceUniq(param.FuncNames)
	}
	err = putRoleFunc(native, param.CcntmractAddr, param.Role, funcs)
	if err != nil {
		invokeEvent(native, "assignFuncsToRole", false)
		return utils.BYTE_FALSE, err
	}
	invokeEvent(native, "assignFuncsToRole", true)
	return utils.BYTE_TRUE, nil
}

func assignToRole(native *native.NativeService) ([]byte, error) {
	param := new(OntIDsToRoleParam)
	rd := bytes.NewReader(native.Input)
	if err := param.Deserialize(rd); err != nil {
		return nil, fmt.Errorf("deserialize failed, caused by %v", err)
	}
	if param.Role == nil {
		return nil, errors.NewErr("role is null")
	}

	//check admin's permission
	admin, err := getCcntmractAdmin(native, param.CcntmractAddr)
	if err != nil {
		return nil, fmt.Errorf("get ccntmract admin failed, caused by %v", err)
	}
	if admin == nil {
		invokeEvent(native, "assignOntIDsToRole", false)
		return utils.BYTE_FALSE, nil
	}
	if bytes.Compare(admin, param.AdminOntID) != 0 {
		invokeEvent(native, "assignOntIDsToRole", false)
		return utils.BYTE_FALSE, nil
	}
	valid, err := verifySig(native, param.AdminOntID, param.KeyNo)
	if err != nil {
		return nil, fmt.Errorf("verify admin's signature failed, caused by %v", err)
	}
	if !valid {
		invokeEvent(native, "assignOntIDsToRole", false)
		return utils.BYTE_FALSE, nil
	}

	//init an auth token
	token := new(AuthToken)
	token.expireTime = uint32(future.Unix())
	token.level = 2
	token.role = param.Role

	for _, p := range param.Persons {
		if p == nil {
			ccntminue
		}
		tokens, err := getOntIDToken(native, param.CcntmractAddr, p)
		if err != nil {
			return nil, err
		}
		if tokens == nil {
			tokens = new(roleTokens)
			tokens.tokens = make([]*AuthToken, 1)
			tokens.tokens[0] = token
		} else {
			ret, err := hasRole(native, param.CcntmractAddr, p, param.Role)
			if err != nil {
				return nil, err
			}
			if !ret {
				tokens.tokens = append(tokens.tokens, token)
			} else {
				ccntminue
			}
		}
		err = putOntIDToken(native, param.CcntmractAddr, p, tokens)
		if err != nil {
			return nil, err
		}
	}
	invokeEvent(native, "assignOntIDsToRole", true)
	return utils.BYTE_TRUE, nil
}

func AssignOntIDsToRole(native *native.NativeService) ([]byte, error) {
	ret, err := assignToRole(native)
	return ret, err
}

func getAuthToken(native *native.NativeService, ccntmractAddr, cntmID, role []byte) (*AuthToken, error) {
	tokens, err := getOntIDToken(native, ccntmractAddr, cntmID)
	if err != nil {
		return nil, fmt.Errorf("get token failed, caused by %v", err)
	}
	if tokens != nil {
		for _, token := range tokens.tokens {
			if bytes.Compare(token.role, role) == 0 { //permanent token
				return token, nil
			}
		}
	}
	status, err := getDelegateStatus(native, ccntmractAddr, cntmID)
	if err != nil {
		return nil, fmt.Errorf("get delegate status failed, caused by %v", err)
	}
	if status != nil {
		for _, s := range status.status {
			if bytes.Compare(s.role, role) == 0 && native.Time < s.expireTime { //temporary token
				token := new(AuthToken)
				token.role = s.role
				token.level = s.level
				token.expireTime = s.expireTime
				return token, nil
			}
		}
	}
	return nil, nil
}

func hasRole(native *native.NativeService, ccntmractAddr, cntmID, role []byte) (bool, error) {
	token, err := getAuthToken(native, ccntmractAddr, cntmID, role)
	if err != nil {
		return false, err
	}
	if token == nil {
		return false, nil
	}
	return true, nil
}

func getLevel(native *native.NativeService, ccntmractAddr, cntmID, role []byte) (uint8, error) {
	token, err := getAuthToken(native, ccntmractAddr, cntmID, role)
	if err != nil {
		return 0, err
	}
	if token == nil {
		return 0, nil
	}
	return token.level, nil
}

/*
 * if 'from' has the authority and 'to' has not been authorized 'role',
 * then make changes to storage as follows:
 */
func delegate(native *native.NativeService, ccntmractAddr []byte, from []byte, to []byte,
	role []byte, period uint32, level uint8, keyNo uint64) ([]byte, error) {
	var fromHasRole, toHasRole bool
	var fromLevel uint8
	var fromExpireTime uint32
	//check from's permission
	ret, err := verifySig(native, from, keyNo)
	if err != nil {
		return nil, err
	}
	if !ret {
		invokeEvent(native, "delegate", false)
		return utils.BYTE_FALSE, nil
	}
	expireTime := uint32(time.Now().Unix())
	if period+expireTime < period {
		invokeEvent(native, "delegate", false)
		return utils.BYTE_FALSE, nil //invalid param
	}
	expireTime = expireTime + period

	fromToken, err := getAuthToken(native, ccntmractAddr, from, role)
	if err != nil {
		return nil, err
	}
	if fromToken == nil {
		fromHasRole = false
		fromLevel = 0
	} else {
		fromHasRole = true
		fromLevel = fromToken.level
		fromExpireTime = fromToken.expireTime
	}
	toToken, err := getAuthToken(native, ccntmractAddr, to, role)
	if err != nil {
		return nil, err
	}
	if toToken == nil {
		toHasRole = false
	} else {
		toHasRole = true
	}
	if !fromHasRole || toHasRole {
		invokeEvent(native, "delegate", false)
		return utils.BYTE_FALSE, nil
	}

	//check if 'from' has the permission to delegate
	if fromLevel == 2 {
		if level < fromLevel && level > 0 && expireTime < fromExpireTime {
			status, err := getDelegateStatus(native, ccntmractAddr, to)
			if err != nil {
				return nil, err
			}
			if status == nil {
				status = new(Status)
			}
			j := -1
			for i, s := range status.status {
				if bytes.Compare(s.role, role) == 0 {
					j = i
					break
				}
			}
			if j < 0 {
				newStatus := &DelegateStatus{
					root: from,
				}
				newStatus.expireTime = expireTime
				newStatus.role = role
				newStatus.level = uint8(level)
				status.status = append(status.status, newStatus)
			} else {
				status.status[j].level = uint8(level)
				status.status[j].expireTime = expireTime
				status.status[j].root = from
			}
			err = putDelegateStatus(native, ccntmractAddr, to, status)
			if err != nil {
				return nil, err
			}
			invokeEvent(native, "delegate", true)
			return utils.BYTE_TRUE, nil
		}
	}
	//TODO: for fromLevel > 2 case
	invokeEvent(native, "delegate", false)
	return utils.BYTE_FALSE, nil
}

func Delegate(native *native.NativeService) ([]byte, error) {
	param := &DelegateParam{}
	rd := bytes.NewReader(native.Input)
	err := param.Deserialize(rd)
	if err != nil {
		return nil, err
	}
	if param.Period > 1<<32 || param.Level > 1<<8 {
		return nil, fmt.Errorf("period or level is too large")
	}
	return delegate(native, param.CcntmractAddr, param.From, param.To, param.Role,
		uint32(param.Period), uint8(param.Level), param.KeyNo)
}

func withdraw(native *native.NativeService, ccntmractAddr []byte, initiator []byte, delegate []byte,
	role []byte, keyNo uint64) (bool, error) {
	//check from's permission
	ret, err := verifySig(native, initiator, keyNo)
	if err != nil {
		return false, err
	}
	if !ret {
		return false, err
	}
	//code below only works in the case that initiator's level is 2
	//TODO
	initToken, err := getAuthToken(native, ccntmractAddr, initiator, role)
	if err != nil {
		return false, err
	}
	if initToken == nil {
		return false, nil
	}
	status, err := getDelegateStatus(native, ccntmractAddr, delegate)
	if err != nil {
		return false, err
	}
	if status == nil {
		return false, nil
	}
	for i, s := range status.status {
		if bytes.Compare(s.role, role) == 0 &&
			bytes.Compare(s.root, initiator) == 0 {
			newStatus := new(Status)
			newStatus.status = append(status.status[:i], status.status[i+1:]...)
			err = putDelegateStatus(native, ccntmractAddr, delegate, newStatus)
			if err != nil {
				return false, err
			}
			return true, nil
		}
	}
	return false, nil
}

func Withdraw(native *native.NativeService) ([]byte, error) {
	param := &WithdrawParam{}
	rd := bytes.NewReader(native.Input)
	param.Deserialize(rd)
	ret, err := withdraw(native, param.CcntmractAddr, param.Initiator, param.Delegate, param.Role, param.KeyNo)
	if err == nil {
		invokeEvent(native, "withdraw", ret)
		if ret {
			return utils.BYTE_TRUE, nil
		} else {
			return utils.BYTE_FALSE, nil
		}
	}
	return utils.BYTE_FALSE, err
}

/*
 *  VerifyToken(ccntmractAddr []byte, caller []byte, fn []byte) (bool, error)
 *  @caller the cntm ID of the caller
 *  @fn the name of the func to call
 *  @tokenSig the signature on the message
 */
func verifyToken(native *native.NativeService, ccntmractAddr []byte, caller []byte, fn []byte, keyNo uint64) (bool, error) {
	//check caller's identity
	ret, err := verifySig(native, caller, keyNo)
	if err != nil {
		return false, err
	}
	if !ret {
		return false, nil
	}
	tokens, err := getOntIDToken(native, ccntmractAddr, caller)
	if err != nil {
		return false, err
	}
	if tokens != nil {
		for _, token := range tokens.tokens {
			funcs, err := getRoleFunc(native, ccntmractAddr, token.role)
			if err != nil {
				return false, nil
			}
			for _, f := range funcs.funcNames {
				if bytes.Compare(fn, []byte(f)) == 0 {
					return true, nil
				}
			}
		}
	}

	status, err := getDelegateStatus(native, ccntmractAddr, caller)
	if err != nil {
		return false, nil
	}
	if status != nil {
		for _, s := range status.status {
			funcs, err := getRoleFunc(native, ccntmractAddr, s.role)
			if err != nil {
				return false, nil
			}
			for _, f := range funcs.funcNames {
				if bytes.Compare(fn, []byte(f)) == 0 {
					return true, nil
				}
			}
		}
	}
	return false, nil
}

func VerifyToken(native *native.NativeService) ([]byte, error) {
	param := &VerifyTokenParam{}
	rd := bytes.NewReader(native.Input)
	err := param.Deserialize(rd)
	if err != nil {
		return nil, err
	}
	ret, err := verifyToken(native, param.CcntmractAddr, param.Caller, param.Fn, param.KeyNo)
	if err != nil {
		return nil, err
	}
	if !ret {
		invokeEvent(native, "verifyToken", false)
		return utils.BYTE_FALSE, nil
	}
	invokeEvent(native, "verifyToken", true)
	return utils.BYTE_TRUE, nil
}

func verifySig(native *native.NativeService, cntmID []byte, keyNo uint64) (bool, error) {
	bf := new(bytes.Buffer)
	if err := serialization.WriteVarBytes(bf, cntmID); err != nil {
		return false, err
	}
	if err := serialization.WriteVarUint(bf, keyNo); err != nil {
		return false, err
	}
	args := bf.Bytes()
	ret, err := native.CcntmextRef.AppCall(utils.OntIDCcntmractAddress, "verifySignature", []byte{}, args)
	if err != nil {
		return false, err
	}
	valid, ok := ret.([]byte)
	if !ok {
		return false, errors.NewErr("verifySignature return non-bool value")
	}
	if bytes.Compare(valid, utils.BYTE_TRUE) == 0 {
		return true, nil
	} else {
		return false, nil
	}
}

func RegisterAuthCcntmract(native *native.NativeService) {
	native.Register("initCcntmractAdmin", InitCcntmractAdmin)
	native.Register("assignFuncsToRole", AssignFuncsToRole)
	native.Register("delegate", Delegate)
	native.Register("withdraw", Withdraw)
	native.Register("assignOntIDsToRole", AssignOntIDsToRole)
	native.Register("verifyToken", VerifyToken)
	native.Register("transfer", Transfer)
}
