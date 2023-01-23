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
	"fmt"

	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/errors"
	"github.com/cntmio/cntmology/smartccntmract/service/native/utils"
)

type SetKeyAccessParam struct {
	OntId     []byte
	SetIndex  uint32
	Access    string
	SignIndex uint32
}

func (this *SetKeyAccessParam) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeVarBytes(sink, this.OntId)
	utils.EncodeVarUint(sink, uint64(this.SetIndex))
	utils.EncodeString(sink, this.Access)
	utils.EncodeVarUint(sink, uint64(this.SignIndex))
}

func (this *SetKeyAccessParam) Deserialization(source *common.ZeroCopySource) error {
	cntmId, err := utils.DecodeVarBytes(source)
	if err != nil {
		return fmt.Errorf("serialization.ReadString, deserialize cntmId error: %v", err)
	}
	setIndex, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("serialization.ReadString, deserialize setIndex error: %v", err)
	}
	access, err := utils.DecodeString(source)
	if err != nil {
		return fmt.Errorf("serialization.ReadString, deserialize access error: %v", err)
	}
	signIndex, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("serialization.ReadString, deserialize signIndex error: %v", err)
	}
	this.OntId = cntmId
	this.SetIndex = uint32(setIndex)
	this.Access = access
	this.SignIndex = uint32(signIndex)
	return nil
}

type ProofParam struct {
	ProofType      string
	Created        string
	Creator        string
	SignatureValue string
}

func (this *ProofParam) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeString(sink, this.ProofType)
	utils.EncodeString(sink, this.Created)
	utils.EncodeString(sink, this.Creator)
	utils.EncodeString(sink, this.SignatureValue)
}

func (this *ProofParam) Deserialization(source *common.ZeroCopySource) error {
	ProofType, err := utils.DecodeString(source)
	if err != nil {
		return fmt.Errorf("serialization.ReadString, deserialize ProofType error: %v", err)
	}
	Created, err := utils.DecodeString(source)
	if err != nil {
		return fmt.Errorf("serialization.ReadString, deserialize Created error: %v", err)
	}
	Creator, err := utils.DecodeString(source)
	if err != nil {
		return fmt.Errorf("serialization.ReadString, deserialize Creator error: %v", err)
	}
	SignatureValue, err := utils.DecodeString(source)
	if err != nil {
		return fmt.Errorf("serialization.ReadString, deserialize SignatureValue error: %v", err)
	}
	this.ProofType = ProofType
	this.Created = Created
	this.Creator = Creator
	this.SignatureValue = SignatureValue
	return nil
}

type ServiceParam struct {
	OntId          []byte
	ServiceId      []byte
	Type           []byte
	ServiceEndpint []byte
	Index          uint32
}

func (this *ServiceParam) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeVarBytes(sink, this.OntId)
	utils.EncodeVarBytes(sink, this.ServiceId)
	utils.EncodeVarBytes(sink, this.Type)
	utils.EncodeVarBytes(sink, this.ServiceEndpint)
	utils.EncodeVarUint(sink, uint64(this.Index))
}

func (this *ServiceParam) Deserialization(source *common.ZeroCopySource) error {
	OntId, err := utils.DecodeVarBytes(source)
	if err != nil {
		return fmt.Errorf("serialization.DecodeVarBytes, deserialize ProofType error: %v", err)
	}
	ServiceId, err := utils.DecodeVarBytes(source)
	if err != nil {
		return fmt.Errorf("serialization.DecodeVarBytes, deserialize Created error: %v", err)
	}
	Type, err := utils.DecodeVarBytes(source)
	if err != nil {
		return fmt.Errorf("serialization.DecodeVarBytes, deserialize Type error: %v", err)
	}
	ServiceEndpint, err := utils.DecodeVarBytes(source)
	if err != nil {
		return fmt.Errorf("serialization.DecodeVarBytes, deserialize ServiceEndpint error: %v", err)
	}
	Index, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("serialization.DecodeVarUint, deserialize Index error: %v", err)
	}
	this.OntId = OntId
	this.ServiceId = ServiceId
	this.Type = Type
	this.ServiceEndpint = ServiceEndpint
	this.Index = uint32(Index)
	return nil
}

type ServiceRemoveParam struct {
	OntId     []byte
	ServiceId []byte
	Index     uint32
}

func (this *ServiceRemoveParam) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeVarBytes(sink, this.OntId)
	utils.EncodeVarBytes(sink, this.ServiceId)
	utils.EncodeVarUint(sink, uint64(this.Index))
}

func (this *ServiceRemoveParam) Deserialization(source *common.ZeroCopySource) error {
	OntId, err := utils.DecodeVarBytes(source)
	if err != nil {
		return fmt.Errorf("serialization.DecodeVarBytes, deserialize ProofType error: %v", err)
	}
	ServiceId, err := utils.DecodeVarBytes(source)
	if err != nil {
		return fmt.Errorf("serialization.DecodeVarBytes, deserialize Created error: %v", err)
	}
	Index, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("serialization.DecodeVarUint, deserialize SignatureValue error: %v", err)
	}
	this.OntId = OntId
	this.ServiceId = ServiceId
	this.Index = uint32(Index)
	return nil
}

type Service struct {
	ServiceId      []byte
	Type           []byte
	ServiceEndpint []byte
}

func (this *Service) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeVarBytes(sink, this.ServiceId)
	utils.EncodeVarBytes(sink, this.Type)
	utils.EncodeVarBytes(sink, this.ServiceEndpint)
}

func (this *Service) Deserialization(source *common.ZeroCopySource) error {
	ServiceId, err := utils.DecodeVarBytes(source)
	if err != nil {
		return fmt.Errorf("serialization.DecodeVarBytes, deserialize Created error: %v", err)
	}
	Type, err := utils.DecodeVarBytes(source)
	if err != nil {
		return fmt.Errorf("serialization.DecodeVarBytes, deserialize Type error: %v", err)
	}
	ServiceEndpint, err := utils.DecodeVarBytes(source)
	if err != nil {
		return fmt.Errorf("serialization.DecodeVarBytes, deserialize ServiceEndpint error: %v", err)
	}
	this.ServiceId = ServiceId
	this.Type = Type
	this.ServiceEndpint = ServiceEndpint
	return nil
}

type Services []Service

func (services *Services) Serialization(sink *common.ZeroCopySink) {
	serviceNum := len(*services)
	utils.EncodeVarUint(sink, uint64(serviceNum))
	for _, service := range *services {
		utils.EncodeVarBytes(sink, service.ServiceId)
		utils.EncodeVarBytes(sink, service.Type)
		utils.EncodeVarBytes(sink, service.ServiceEndpint)
	}
}

func (services *Services) Deserialization(source *common.ZeroCopySource) error {
	serviceNum, err := utils.DecodeVarUint(source)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "deserialize services length error!")
	}
	for i := 0; uint64(i) < serviceNum; i++ {
		service := Service{}
		var irregular, eof bool
		service.ServiceId, _, irregular, eof = source.NextVarBytes()
		if irregular || eof {
			return errors.NewDetailErr(err, errors.ErrNoCode, fmt.Sprintf("deserialize service id %v error!", service.ServiceId))
		}
		service.Type, _, irregular, eof = source.NextVarBytes()
		if irregular || eof {
			return errors.NewDetailErr(err, errors.ErrNoCode, fmt.Sprintf("deserialize service type %v error!", service.Type))
		}
		service.ServiceEndpint, _, irregular, eof = source.NextVarBytes()
		if irregular || eof {
			return errors.NewDetailErr(err, errors.ErrNoCode, fmt.Sprintf("deserialize service endpint%v error!", service.ServiceEndpint))
		}
		*services = append(*services, service)
	}
	return nil
}

type Ccntmext struct {
	OntId    []byte
	Ccntmexts [][]byte
	Index    uint32
}

func (this *Ccntmext) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeVarBytes(sink, this.OntId)
	ccntmextNum := len(this.Ccntmexts)
	utils.EncodeVarUint(sink, uint64(ccntmextNum))
	for _, c := range this.Ccntmexts {
		utils.EncodeVarBytes(sink, c)
	}
	utils.EncodeVarUint(sink, uint64(this.Index))
}

func (this *Ccntmext) Deserialization(source *common.ZeroCopySource) error {
	OntId, err := utils.DecodeVarBytes(source)
	if err != nil {
		return fmt.Errorf("serialization.DecodeVarBytes, deserialize Created error: %v", err)
	}
	cNum, err := utils.DecodeVarUint(source)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "deserialize ccntmext length error!")
	}
	var ccntmexts [][]byte
	for i := 0; uint64(i) < cNum; i++ {
		var irregular, eof bool
		item, _, irregular, eof := source.NextVarBytes()
		if irregular || eof {
			return errors.NewDetailErr(err, errors.ErrNoCode, fmt.Sprintf("deserialize ccntmext %v error!", item))
		}
		ccntmexts = append(ccntmexts, item)
	}
	index, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("serialization.DecodeVarUint, deserialize index error: %v", err)
	}
	this.OntId = OntId
	this.Ccntmexts = ccntmexts
	this.Index = uint32(index)
	return nil
}

type Ccntmexts [][]byte

func (ccntmexts *Ccntmexts) Serialization(sink *common.ZeroCopySink) {
	ccntmextNum := len(*ccntmexts)
	utils.EncodeVarUint(sink, uint64(ccntmextNum))
	for _, c := range *ccntmexts {
		utils.EncodeVarBytes(sink, c)
	}
}

func (ccntmexts *Ccntmexts) Deserialization(source *common.ZeroCopySource) error {
	cNum, err := utils.DecodeVarUint(source)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "deserialize ccntmext length error!")
	}
	for i := 0; uint64(i) < cNum; i++ {
		var irregular, eof bool
		item, _, irregular, eof := source.NextVarBytes()
		if irregular || eof {
			return errors.NewDetailErr(err, errors.ErrNoCode, fmt.Sprintf("deserialize ccntmext %v error!", item))
		}
		*ccntmexts = append(*ccntmexts, item)
	}
	return nil
}

type AddNewAuthKeyParam struct {
	OntId        []byte
	NewPublicKey *NewPublicKey
	SignIndex    uint32
}

func (this *AddNewAuthKeyParam) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeVarBytes(sink, this.OntId)
	this.NewPublicKey.Serialization(sink)
	utils.EncodeVarUint(sink, uint64(this.SignIndex))
}

func (this *AddNewAuthKeyParam) Deserialization(source *common.ZeroCopySource) error {
	cntmId, err := utils.DecodeVarBytes(source)
	if err != nil {
		return fmt.Errorf("utils.DecodeVarBytes, deserialize cntmId error: %v", err)
	}
	newPublicKey := new(NewPublicKey)
	err = newPublicKey.Deserialization(source)
	if err != nil {
		return fmt.Errorf("newPublicKey.Deserialization, deserialize newPublicKey error: %v", err)
	}
	signIndex, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("utils.DecodeVarUint, deserialize signIndex error: %v", err)
	}
	this.OntId = cntmId
	this.NewPublicKey = newPublicKey
	this.SignIndex = uint32(signIndex)
	return nil
}

type SetAuthKeyParam struct {
	OntId     []byte
	Index     uint32
	SignIndex uint32
}

func (this *SetAuthKeyParam) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeVarBytes(sink, this.OntId)
	utils.EncodeVarUint(sink, uint64(this.Index))
	utils.EncodeVarUint(sink, uint64(this.SignIndex))
}

func (this *SetAuthKeyParam) Deserialization(source *common.ZeroCopySource) error {
	cntmId, err := utils.DecodeVarBytes(source)
	if err != nil {
		return fmt.Errorf("utils.DecodeVarBytes, deserialize cntmId error: %v", err)
	}
	index, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("utils.DecodeVarUint, deserialize index error: %v", err)
	}
	signIndex, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("utils.DecodeVarUint, deserialize signIndex error: %v", err)
	}
	this.OntId = cntmId
	this.Index = uint32(index)
	this.SignIndex = uint32(signIndex)
	return nil
}

type RemoveAuthKeyParam struct {
	OntId     []byte
	Index     uint32
	SignIndex uint32
}

func (this *RemoveAuthKeyParam) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeVarBytes(sink, this.OntId)
	utils.EncodeVarUint(sink, uint64(this.Index))
	utils.EncodeVarUint(sink, uint64(this.SignIndex))
}

func (this *RemoveAuthKeyParam) Deserialization(source *common.ZeroCopySource) error {
	cntmId, err := utils.DecodeVarBytes(source)
	if err != nil {
		return fmt.Errorf("utils.DecodeVarBytes, deserialize cntmId error: %v", err)
	}
	index, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("utils.DecodeVarUint, deserialize index error: %v", err)
	}
	signIndex, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("utils.DecodeVarUint, deserialize signIndex error: %v", err)
	}
	this.OntId = cntmId
	this.Index = uint32(index)
	this.SignIndex = uint32(signIndex)
	return nil
}

type NewPublicKey struct {
	key        []byte
	ccntmroller []byte
}

func (this *NewPublicKey) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeVarBytes(sink, this.key)
	utils.EncodeVarBytes(sink, this.ccntmroller)
}

func (this *NewPublicKey) Deserialization(source *common.ZeroCopySource) error {
	key, err := utils.DecodeVarBytes(source)
	if err != nil {
		return err
	}
	ccntmroller, err := utils.DecodeVarBytes(source)
	if err != nil {
		return err
	}

	this.key = key
	this.ccntmroller = ccntmroller
	return nil
}

type SearchServiceParam struct {
	OntId     []byte `json:"id"`
	ServiceId []byte `json:"serviceId"`
}

func (this *SearchServiceParam) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeVarBytes(sink, this.OntId)
	utils.EncodeVarBytes(sink, this.ServiceId)
}

func (this *SearchServiceParam) Deserialization(source *common.ZeroCopySource) error {
	OntId, err := utils.DecodeVarBytes(source)
	if err != nil {
		return err
	}
	ServiceId, err := utils.DecodeVarBytes(source)
	if err != nil {
		return err
	}

	this.OntId = OntId
	this.ServiceId = ServiceId
	return nil
}

type Document struct {
	Ccntmexts       []string         `json:"@ccntmext"`
	Id             string           `json:"id"`
	PublicKey      []*publicKeyJson `json:"publicKey"`
	Authentication []interface{}    `json:"authentication"`
	Ccntmroller     interface{}      `json:"ccntmroller"`
	Recovery       *GroupJson       `json:"recovery"`
	Service        []*serviceJson   `json:"service"`
	Attribute      []*attributeJson `json:"attribute"`
	Created        uint32           `json:"created"`
	Updated        uint32           `json:"updated"`
	Proof          string           `json:"proof"`
}
