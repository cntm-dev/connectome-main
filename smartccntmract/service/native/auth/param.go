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
	"fmt"
	"io"
	"math"

	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/common/serialization"
	"github.com/cntmio/cntmology/smartccntmract/service/native/utils"
)

/* **********************************************   */
type InitCcntmractAdminParam struct {
	AdminOntID []byte
}

func (this *InitCcntmractAdminParam) Serialize(w io.Writer) error {
	if err := serialization.WriteVarBytes(w, this.AdminOntID); err != nil {
		return err
	}
	return nil
}

func (this *InitCcntmractAdminParam) Deserialize(rd io.Reader) error {
	var err error
	if this.AdminOntID, err = serialization.ReadVarBytes(rd); err != nil {
		return err
	}
	return nil
}

/* **********************************************   */
type TransferParam struct {
	CcntmractAddr  common.Address
	NewAdminOntID []byte
	KeyNo         uint64
}

func (this *TransferParam) Serialize(w io.Writer) error {
	if err := serializeAddress(w, this.CcntmractAddr); err != nil {
		return err
	}
	if err := serialization.WriteVarBytes(w, this.NewAdminOntID); err != nil {
		return err
	}
	if err := utils.WriteVarUint(w, this.KeyNo); err != nil {
		return nil
	}
	return nil
}

func (this *TransferParam) Deserialize(rd io.Reader) error {
	var err error
	if this.CcntmractAddr, err = utils.ReadAddress(rd); err != nil {
		return err
	}
	if this.NewAdminOntID, err = serialization.ReadVarBytes(rd); err != nil {
		return err
	}
	if this.KeyNo, err = utils.ReadVarUint(rd); err != nil {
		return err
	}
	return nil
}

/* **********************************************   */
type FuncsToRoleParam struct {
	CcntmractAddr common.Address
	AdminOntID   []byte
	Role         []byte
	FuncNames    []string
	KeyNo        uint64
}

func (this *FuncsToRoleParam) Serialize(w io.Writer) error {
	if err := serializeAddress(w, this.CcntmractAddr); err != nil {
		return err
	}
	if err := serialization.WriteVarBytes(w, this.AdminOntID); err != nil {
		return err
	}
	if err := serialization.WriteVarBytes(w, this.Role); err != nil {
		return err
	}
	if err := serialization.WriteVarUint(w, uint64(len(this.FuncNames))); err != nil {
		return err
	}
	for _, fn := range this.FuncNames {
		if err := serialization.WriteString(w, fn); err != nil {
			return err
		}
	}
	if err := utils.WriteVarUint(w, this.KeyNo); err != nil {
		return nil
	}
	return nil
}

func (this *FuncsToRoleParam) Deserialize(rd io.Reader) error {
	var err error
	var fnLen uint64
	var i uint64

	if this.CcntmractAddr, err = utils.ReadAddress(rd); err != nil {
		return err
	}
	if this.AdminOntID, err = serialization.ReadVarBytes(rd); err != nil {
		return err
	}
	if this.Role, err = serialization.ReadVarBytes(rd); err != nil {
		return err
	}
	if fnLen, err = serialization.ReadVarUint(rd, 0); err != nil {
		return err
	}
	this.FuncNames = make([]string, 0)
	for i = 0; i < fnLen; i++ {
		fn, err := serialization.ReadString(rd)
		if err != nil {
			return err
		}
		this.FuncNames = append(this.FuncNames, fn)
	}
	if this.KeyNo, err = utils.ReadVarUint(rd); err != nil {
		return err
	}
	return nil
}

type OntIDsToRoleParam struct {
	CcntmractAddr common.Address
	AdminOntID   []byte
	Role         []byte
	Persons      [][]byte
	KeyNo        uint64
}

func (this *OntIDsToRoleParam) Serialize(w io.Writer) error {
	if err := serializeAddress(w, this.CcntmractAddr); err != nil {
		return err
	}
	if err := serialization.WriteVarBytes(w, this.AdminOntID); err != nil {
		return err
	}
	if err := serialization.WriteVarBytes(w, this.Role); err != nil {
		return err
	}
	if err := serialization.WriteVarUint(w, uint64(len(this.Persons))); err != nil {
		return err
	}
	for _, p := range this.Persons {
		if err := serialization.WriteVarBytes(w, p); err != nil {
			return err
		}
	}
	if err := utils.WriteVarUint(w, this.KeyNo); err != nil {
		return nil
	}
	return nil
}

func (this *OntIDsToRoleParam) Deserialize(rd io.Reader) error {
	var err error
	var pLen uint64
	if this.CcntmractAddr, err = utils.ReadAddress(rd); err != nil {
		return err
	}
	if this.AdminOntID, err = serialization.ReadVarBytes(rd); err != nil {
		return err
	}
	if this.Role, err = serialization.ReadVarBytes(rd); err != nil {
		return err
	}
	if pLen, err = serialization.ReadVarUint(rd, 0); err != nil {
		return err
	}
	this.Persons = make([][]byte, 0)
	for i := uint64(0); i < pLen; i++ {
		p, err := serialization.ReadVarBytes(rd)
		if err != nil {
			return err
		}
		this.Persons = append(this.Persons, p)
	}
	if this.KeyNo, err = utils.ReadVarUint(rd); err != nil {
		return err
	}
	return nil
}

type DelegateParam struct {
	CcntmractAddr common.Address
	From         []byte
	To           []byte
	Role         []byte
	Period       uint64
	Level        uint64
	KeyNo        uint64
}

func (this *DelegateParam) Serialize(w io.Writer) error {
	if err := serializeAddress(w, this.CcntmractAddr); err != nil {
		return err
	}
	if err := serialization.WriteVarBytes(w, this.From); err != nil {
		return err
	}
	if err := serialization.WriteVarBytes(w, this.To); err != nil {
		return err
	}
	if err := serialization.WriteVarBytes(w, this.Role); err != nil {
		return err
	}
	if err := utils.WriteVarUint(w, this.Period); err != nil {
		return err
	}
	if err := utils.WriteVarUint(w, uint64(this.Level)); err != nil {
		return err
	}
	if err := utils.WriteVarUint(w, this.KeyNo); err != nil {
		return err
	}
	return nil
}

func (this *DelegateParam) Deserialize(rd io.Reader) error {
	var err error
	var level uint64
	if this.CcntmractAddr, err = utils.ReadAddress(rd); err != nil {
		return err
	}
	if this.From, err = serialization.ReadVarBytes(rd); err != nil {
		return err
	}
	if this.To, err = serialization.ReadVarBytes(rd); err != nil {
		return err
	}
	if this.Role, err = serialization.ReadVarBytes(rd); err != nil {
		return err
	}
	if this.Period, err = utils.ReadVarUint(rd); err != nil {
		return err
	}
	if level, err = utils.ReadVarUint(rd); err != nil {
		return err
	}
	if this.KeyNo, err = utils.ReadVarUint(rd); err != nil {
		return err
	}
	if level > math.MaxInt8 || this.Period > math.MaxUint32 {
		return fmt.Errorf("period or level too large: (%d, %d)", this.Period, level)
	}
	this.Level = level
	return nil
}

type WithdrawParam struct {
	CcntmractAddr common.Address
	Initiator    []byte
	Delegate     []byte
	Role         []byte
	KeyNo        uint64
}

func (this *WithdrawParam) Serialize(w io.Writer) error {
	if err := serializeAddress(w, this.CcntmractAddr); err != nil {
		return err
	}
	if err := serialization.WriteVarBytes(w, this.Initiator); err != nil {
		return err
	}
	if err := serialization.WriteVarBytes(w, this.Delegate); err != nil {
		return err
	}
	if err := serialization.WriteVarBytes(w, this.Role); err != nil {
		return err
	}
	if err := utils.WriteVarUint(w, this.KeyNo); err != nil {
		return err
	}
	return nil
}
func (this *WithdrawParam) Deserialize(rd io.Reader) error {
	var err error
	if this.CcntmractAddr, err = utils.ReadAddress(rd); err != nil {
		return err
	}
	if this.Initiator, err = serialization.ReadVarBytes(rd); err != nil {
		return err
	}
	if this.Delegate, err = serialization.ReadVarBytes(rd); err != nil {
		return err
	}
	if this.Role, err = serialization.ReadVarBytes(rd); err != nil {
		return err
	}
	if this.KeyNo, err = utils.ReadVarUint(rd); err != nil {
		return err
	}
	return nil
}

type VerifyTokenParam struct {
	CcntmractAddr common.Address
	Caller       []byte
	Fn           string
	KeyNo        uint64
}

func (this *VerifyTokenParam) Serialize(w io.Writer) error {
	if err := serializeAddress(w, this.CcntmractAddr); err != nil {
		return err
	}
	if err := serialization.WriteVarBytes(w, this.Caller); err != nil {
		return err
	}
	if err := serialization.WriteString(w, this.Fn); err != nil {
		return err
	}
	if err := utils.WriteVarUint(w, this.KeyNo); err != nil {
		return err
	}
	return nil
}

func (this *VerifyTokenParam) Deserialize(rd io.Reader) error {
	var err error
	if this.CcntmractAddr, err = utils.ReadAddress(rd); err != nil {
		return err
	}
	if this.Caller, err = serialization.ReadVarBytes(rd); err != nil {
		return err //deserialize caller error
	}
	if this.Fn, err = serialization.ReadString(rd); err != nil {
		return err
	}
	if this.KeyNo, err = utils.ReadVarUint(rd); err != nil {
		return err
	}
	return nil
}