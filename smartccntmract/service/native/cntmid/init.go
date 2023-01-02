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
package cntmid

import (
	"github.com/cntmio/cntmology/smartccntmract/service/native"
	"github.com/cntmio/cntmology/smartccntmract/service/native/utils"
)

func Init() {
	native.Ccntmracts[utils.OntIDCcntmractAddress] = RegisterIDCcntmract
}

func RegisterIDCcntmract(srvc *native.NativeService) {
	srvc.Register("regIDWithPublicKey", regIdWithPublicKey)
	srvc.Register("regIDWithCcntmroller", regIdWithCcntmroller)
	srvc.Register("revokeID", revokeID)
	srvc.Register("revokeIDByCcntmroller", revokeIDByCcntmroller)
	srvc.Register("removeCcntmroller", removeCcntmroller)
	srvc.Register("addRecovery", addRecovery)
	srvc.Register("changeRecovery", changeRecovery)
	srvc.Register("addKey", addKey)
	srvc.Register("removeKey", removeKey)
	srvc.Register("addKeyByCcntmroller", addKeyByCcntmroller)
	srvc.Register("removeKeyByCcntmroller", removeKeyByCcntmroller)
	srvc.Register("addKeyByRecovery", addKeyByRecovery)
	srvc.Register("removeKeyByRecovery", removeKeyByRecovery)
	srvc.Register("regIDWithAttributes", regIdWithAttributes)
	srvc.Register("addAttributes", addAttributes)
	srvc.Register("removeAttribute", removeAttribute)
	srvc.Register("addAttributesByCcntmroller", addAttributesByCcntmroller)
	srvc.Register("removeAttributeByCcntmroller", removeAttributeByCcntmroller)
	srvc.Register("verifySignature", verifySignature)
	srvc.Register("verifyCcntmroller", verifyCcntmroller)
	srvc.Register("getPublicKeys", GetPublicKeys)
	srvc.Register("getKeyState", GetKeyState)
	srvc.Register("getAttributes", GetAttributes)
	srvc.Register("getDDO", GetDDO)
	return
}
