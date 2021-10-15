// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: github.com/Ontology/eventbus/example/testRemoteCrypto/commons/protos.proto

/*
	Package commons is a generated protocol buffer package.

	It is generated from these files:
		github.com/Ontology/eventbus/example/testRemoteCrypto/commons/protos.proto

	It has these top-level messages:
		RunMsg
		SignRequest
		SignResponse
		SetPrivKey
		VerifyRequest
		VerifyResponse
*/
package commons

import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"

import bytes "bytes"

import strings "strings"
import reflect "reflect"

import io "io"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion2 // please upgrade the proto package

type RunMsg struct {
}

func (m *RunMsg) Reset()                    { *m = RunMsg{} }
func (*RunMsg) ProtoMessage()               {}
func (*RunMsg) Descriptor() ([]byte, []int) { return fileDescriptorProtos, []int{0} }

type SignRequest struct {
	Data []byte `protobuf:"bytes,1,opt,name=Data,proto3" json:"Data,omitempty"`
	Seq  string `protobuf:"bytes,2,opt,name=Seq,proto3" json:"Seq,omitempty"`
}

func (m *SignRequest) Reset()                    { *m = SignRequest{} }
func (*SignRequest) ProtoMessage()               {}
func (*SignRequest) Descriptor() ([]byte, []int) { return fileDescriptorProtos, []int{1} }

func (m *SignRequest) GetData() []byte {
	if m != nil {
		return m.Data
	}
	return nil
}

func (m *SignRequest) GetSeq() string {
	if m != nil {
		return m.Seq
	}
	return ""
}

type SignResponse struct {
	Signature []byte `protobuf:"bytes,1,opt,name=Signature,proto3" json:"Signature,omitempty"`
	Seq       string `protobuf:"bytes,2,opt,name=Seq,proto3" json:"Seq,omitempty"`
}

func (m *SignResponse) Reset()                    { *m = SignResponse{} }
func (*SignResponse) ProtoMessage()               {}
func (*SignResponse) Descriptor() ([]byte, []int) { return fileDescriptorProtos, []int{2} }

func (m *SignResponse) GetSignature() []byte {
	if m != nil {
		return m.Signature
	}
	return nil
}

func (m *SignResponse) GetSeq() string {
	if m != nil {
		return m.Seq
	}
	return ""
}

type SetPrivKey struct {
	PrivKey []byte `protobuf:"bytes,1,opt,name=PrivKey,proto3" json:"PrivKey,omitempty"`
}

func (m *SetPrivKey) Reset()                    { *m = SetPrivKey{} }
func (*SetPrivKey) ProtoMessage()               {}
func (*SetPrivKey) Descriptor() ([]byte, []int) { return fileDescriptorProtos, []int{3} }

func (m *SetPrivKey) GetPrivKey() []byte {
	if m != nil {
		return m.PrivKey
	}
	return nil
}

type VerifyRequest struct {
	Signature []byte `protobuf:"bytes,1,opt,name=Signature,proto3" json:"Signature,omitempty"`
	Data      []byte `protobuf:"bytes,2,opt,name=Data,proto3" json:"Data,omitempty"`
	PublicKey []byte `protobuf:"bytes,3,opt,name=PublicKey,proto3" json:"PublicKey,omitempty"`
	Seq       string `protobuf:"bytes,4,opt,name=Seq,proto3" json:"Seq,omitempty"`
}

func (m *VerifyRequest) Reset()                    { *m = VerifyRequest{} }
func (*VerifyRequest) ProtoMessage()               {}
func (*VerifyRequest) Descriptor() ([]byte, []int) { return fileDescriptorProtos, []int{4} }

func (m *VerifyRequest) GetSignature() []byte {
	if m != nil {
		return m.Signature
	}
	return nil
}

func (m *VerifyRequest) GetData() []byte {
	if m != nil {
		return m.Data
	}
	return nil
}

func (m *VerifyRequest) GetPublicKey() []byte {
	if m != nil {
		return m.PublicKey
	}
	return nil
}

func (m *VerifyRequest) GetSeq() string {
	if m != nil {
		return m.Seq
	}
	return ""
}

type VerifyResponse struct {
	Seq      string `protobuf:"bytes,1,opt,name=Seq,proto3" json:"Seq,omitempty"`
	Result   bool   `protobuf:"varint,2,opt,name=Result,proto3" json:"Result,omitempty"`
	ErrorMsg string `protobuf:"bytes,3,opt,name=ErrorMsg,proto3" json:"ErrorMsg,omitempty"`
}

func (m *VerifyResponse) Reset()                    { *m = VerifyResponse{} }
func (*VerifyResponse) ProtoMessage()               {}
func (*VerifyResponse) Descriptor() ([]byte, []int) { return fileDescriptorProtos, []int{5} }

func (m *VerifyResponse) GetSeq() string {
	if m != nil {
		return m.Seq
	}
	return ""
}

func (m *VerifyResponse) GetResult() bool {
	if m != nil {
		return m.Result
	}
	return false
}

func (m *VerifyResponse) GetErrorMsg() string {
	if m != nil {
		return m.ErrorMsg
	}
	return ""
}

func init() {
	proto.RegisterType((*RunMsg)(nil), "commons.RunMsg")
	proto.RegisterType((*SignRequest)(nil), "commons.SignRequest")
	proto.RegisterType((*SignResponse)(nil), "commons.SignResponse")
	proto.RegisterType((*SetPrivKey)(nil), "commons.SetPrivKey")
	proto.RegisterType((*VerifyRequest)(nil), "commons.VerifyRequest")
	proto.RegisterType((*VerifyResponse)(nil), "commons.VerifyResponse")
}
func (this *RunMsg) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*RunMsg)
	if !ok {
		that2, ok := that.(RunMsg)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		return this == nil
	} else if this == nil {
		return false
	}
	return true
}
func (this *SignRequest) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*SignRequest)
	if !ok {
		that2, ok := that.(SignRequest)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		return this == nil
	} else if this == nil {
		return false
	}
	if !bytes.Equal(this.Data, that1.Data) {
		return false
	}
	if this.Seq != that1.Seq {
		return false
	}
	return true
}
func (this *SignResponse) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*SignResponse)
	if !ok {
		that2, ok := that.(SignResponse)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		return this == nil
	} else if this == nil {
		return false
	}
	if !bytes.Equal(this.Signature, that1.Signature) {
		return false
	}
	if this.Seq != that1.Seq {
		return false
	}
	return true
}
func (this *SetPrivKey) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*SetPrivKey)
	if !ok {
		that2, ok := that.(SetPrivKey)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		return this == nil
	} else if this == nil {
		return false
	}
	if !bytes.Equal(this.PrivKey, that1.PrivKey) {
		return false
	}
	return true
}
func (this *VerifyRequest) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*VerifyRequest)
	if !ok {
		that2, ok := that.(VerifyRequest)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		return this == nil
	} else if this == nil {
		return false
	}
	if !bytes.Equal(this.Signature, that1.Signature) {
		return false
	}
	if !bytes.Equal(this.Data, that1.Data) {
		return false
	}
	if !bytes.Equal(this.PublicKey, that1.PublicKey) {
		return false
	}
	if this.Seq != that1.Seq {
		return false
	}
	return true
}
func (this *VerifyResponse) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*VerifyResponse)
	if !ok {
		that2, ok := that.(VerifyResponse)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		return this == nil
	} else if this == nil {
		return false
	}
	if this.Seq != that1.Seq {
		return false
	}
	if this.Result != that1.Result {
		return false
	}
	if this.ErrorMsg != that1.ErrorMsg {
		return false
	}
	return true
}
func (this *RunMsg) GoString() string {
	if this == nil {
		return "nil"
	}
	s := make([]string, 0, 4)
	s = append(s, "&commons.RunMsg{")
	s = append(s, "}")
	return strings.Join(s, "")
}
func (this *SignRequest) GoString() string {
	if this == nil {
		return "nil"
	}
	s := make([]string, 0, 6)
	s = append(s, "&commons.SignRequest{")
	s = append(s, "Data: "+fmt.Sprintf("%#v", this.Data)+",\n")
	s = append(s, "Seq: "+fmt.Sprintf("%#v", this.Seq)+",\n")
	s = append(s, "}")
	return strings.Join(s, "")
}
func (this *SignResponse) GoString() string {
	if this == nil {
		return "nil"
	}
	s := make([]string, 0, 6)
	s = append(s, "&commons.SignResponse{")
	s = append(s, "Signature: "+fmt.Sprintf("%#v", this.Signature)+",\n")
	s = append(s, "Seq: "+fmt.Sprintf("%#v", this.Seq)+",\n")
	s = append(s, "}")
	return strings.Join(s, "")
}
func (this *SetPrivKey) GoString() string {
	if this == nil {
		return "nil"
	}
	s := make([]string, 0, 5)
	s = append(s, "&commons.SetPrivKey{")
	s = append(s, "PrivKey: "+fmt.Sprintf("%#v", this.PrivKey)+",\n")
	s = append(s, "}")
	return strings.Join(s, "")
}
func (this *VerifyRequest) GoString() string {
	if this == nil {
		return "nil"
	}
	s := make([]string, 0, 8)
	s = append(s, "&commons.VerifyRequest{")
	s = append(s, "Signature: "+fmt.Sprintf("%#v", this.Signature)+",\n")
	s = append(s, "Data: "+fmt.Sprintf("%#v", this.Data)+",\n")
	s = append(s, "PublicKey: "+fmt.Sprintf("%#v", this.PublicKey)+",\n")
	s = append(s, "Seq: "+fmt.Sprintf("%#v", this.Seq)+",\n")
	s = append(s, "}")
	return strings.Join(s, "")
}
func (this *VerifyResponse) GoString() string {
	if this == nil {
		return "nil"
	}
	s := make([]string, 0, 7)
	s = append(s, "&commons.VerifyResponse{")
	s = append(s, "Seq: "+fmt.Sprintf("%#v", this.Seq)+",\n")
	s = append(s, "Result: "+fmt.Sprintf("%#v", this.Result)+",\n")
	s = append(s, "ErrorMsg: "+fmt.Sprintf("%#v", this.ErrorMsg)+",\n")
	s = append(s, "}")
	return strings.Join(s, "")
}
func valueToGoStringProtos(v interface{}, typ string) string {
	rv := reflect.ValueOf(v)
	if rv.IsNil() {
		return "nil"
	}
	pv := reflect.Indirect(rv).Interface()
	return fmt.Sprintf("func(v %v) *%v { return &v } ( %#v )", typ, typ, pv)
}
func (m *RunMsg) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *RunMsg) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	return i, nil
}

func (m *SignRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *SignRequest) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.Data) > 0 {
		dAtA[i] = 0xa
		i++
		i = encodeVarintProtos(dAtA, i, uint64(len(m.Data)))
		i += copy(dAtA[i:], m.Data)
	}
	if len(m.Seq) > 0 {
		dAtA[i] = 0x12
		i++
		i = encodeVarintProtos(dAtA, i, uint64(len(m.Seq)))
		i += copy(dAtA[i:], m.Seq)
	}
	return i, nil
}

func (m *SignResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *SignResponse) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.Signature) > 0 {
		dAtA[i] = 0xa
		i++
		i = encodeVarintProtos(dAtA, i, uint64(len(m.Signature)))
		i += copy(dAtA[i:], m.Signature)
	}
	if len(m.Seq) > 0 {
		dAtA[i] = 0x12
		i++
		i = encodeVarintProtos(dAtA, i, uint64(len(m.Seq)))
		i += copy(dAtA[i:], m.Seq)
	}
	return i, nil
}

func (m *SetPrivKey) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *SetPrivKey) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.PrivKey) > 0 {
		dAtA[i] = 0xa
		i++
		i = encodeVarintProtos(dAtA, i, uint64(len(m.PrivKey)))
		i += copy(dAtA[i:], m.PrivKey)
	}
	return i, nil
}

func (m *VerifyRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *VerifyRequest) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.Signature) > 0 {
		dAtA[i] = 0xa
		i++
		i = encodeVarintProtos(dAtA, i, uint64(len(m.Signature)))
		i += copy(dAtA[i:], m.Signature)
	}
	if len(m.Data) > 0 {
		dAtA[i] = 0x12
		i++
		i = encodeVarintProtos(dAtA, i, uint64(len(m.Data)))
		i += copy(dAtA[i:], m.Data)
	}
	if len(m.PublicKey) > 0 {
		dAtA[i] = 0x1a
		i++
		i = encodeVarintProtos(dAtA, i, uint64(len(m.PublicKey)))
		i += copy(dAtA[i:], m.PublicKey)
	}
	if len(m.Seq) > 0 {
		dAtA[i] = 0x22
		i++
		i = encodeVarintProtos(dAtA, i, uint64(len(m.Seq)))
		i += copy(dAtA[i:], m.Seq)
	}
	return i, nil
}

func (m *VerifyResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *VerifyResponse) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.Seq) > 0 {
		dAtA[i] = 0xa
		i++
		i = encodeVarintProtos(dAtA, i, uint64(len(m.Seq)))
		i += copy(dAtA[i:], m.Seq)
	}
	if m.Result {
		dAtA[i] = 0x10
		i++
		if m.Result {
			dAtA[i] = 1
		} else {
			dAtA[i] = 0
		}
		i++
	}
	if len(m.ErrorMsg) > 0 {
		dAtA[i] = 0x1a
		i++
		i = encodeVarintProtos(dAtA, i, uint64(len(m.ErrorMsg)))
		i += copy(dAtA[i:], m.ErrorMsg)
	}
	return i, nil
}

func encodeVarintProtos(dAtA []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return offset + 1
}
func (m *RunMsg) Size() (n int) {
	var l int
	_ = l
	return n
}

func (m *SignRequest) Size() (n int) {
	var l int
	_ = l
	l = len(m.Data)
	if l > 0 {
		n += 1 + l + sovProtos(uint64(l))
	}
	l = len(m.Seq)
	if l > 0 {
		n += 1 + l + sovProtos(uint64(l))
	}
	return n
}

func (m *SignResponse) Size() (n int) {
	var l int
	_ = l
	l = len(m.Signature)
	if l > 0 {
		n += 1 + l + sovProtos(uint64(l))
	}
	l = len(m.Seq)
	if l > 0 {
		n += 1 + l + sovProtos(uint64(l))
	}
	return n
}

func (m *SetPrivKey) Size() (n int) {
	var l int
	_ = l
	l = len(m.PrivKey)
	if l > 0 {
		n += 1 + l + sovProtos(uint64(l))
	}
	return n
}

func (m *VerifyRequest) Size() (n int) {
	var l int
	_ = l
	l = len(m.Signature)
	if l > 0 {
		n += 1 + l + sovProtos(uint64(l))
	}
	l = len(m.Data)
	if l > 0 {
		n += 1 + l + sovProtos(uint64(l))
	}
	l = len(m.PublicKey)
	if l > 0 {
		n += 1 + l + sovProtos(uint64(l))
	}
	l = len(m.Seq)
	if l > 0 {
		n += 1 + l + sovProtos(uint64(l))
	}
	return n
}

func (m *VerifyResponse) Size() (n int) {
	var l int
	_ = l
	l = len(m.Seq)
	if l > 0 {
		n += 1 + l + sovProtos(uint64(l))
	}
	if m.Result {
		n += 2
	}
	l = len(m.ErrorMsg)
	if l > 0 {
		n += 1 + l + sovProtos(uint64(l))
	}
	return n
}

func sovProtos(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozProtos(x uint64) (n int) {
	return sovProtos(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (this *RunMsg) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&RunMsg{`,
		`}`,
	}, "")
	return s
}
func (this *SignRequest) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&SignRequest{`,
		`Data:` + fmt.Sprintf("%v", this.Data) + `,`,
		`Seq:` + fmt.Sprintf("%v", this.Seq) + `,`,
		`}`,
	}, "")
	return s
}
func (this *SignResponse) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&SignResponse{`,
		`Signature:` + fmt.Sprintf("%v", this.Signature) + `,`,
		`Seq:` + fmt.Sprintf("%v", this.Seq) + `,`,
		`}`,
	}, "")
	return s
}
func (this *SetPrivKey) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&SetPrivKey{`,
		`PrivKey:` + fmt.Sprintf("%v", this.PrivKey) + `,`,
		`}`,
	}, "")
	return s
}
func (this *VerifyRequest) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&VerifyRequest{`,
		`Signature:` + fmt.Sprintf("%v", this.Signature) + `,`,
		`Data:` + fmt.Sprintf("%v", this.Data) + `,`,
		`PublicKey:` + fmt.Sprintf("%v", this.PublicKey) + `,`,
		`Seq:` + fmt.Sprintf("%v", this.Seq) + `,`,
		`}`,
	}, "")
	return s
}
func (this *VerifyResponse) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&VerifyResponse{`,
		`Seq:` + fmt.Sprintf("%v", this.Seq) + `,`,
		`Result:` + fmt.Sprintf("%v", this.Result) + `,`,
		`ErrorMsg:` + fmt.Sprintf("%v", this.ErrorMsg) + `,`,
		`}`,
	}, "")
	return s
}
func valueToStringProtos(v interface{}) string {
	rv := reflect.ValueOf(v)
	if rv.IsNil() {
		return "nil"
	}
	pv := reflect.Indirect(rv).Interface()
	return fmt.Sprintf("*%v", pv)
}
func (m *RunMsg) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowProtos
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: RunMsg: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: RunMsg: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		default:
			iNdEx = preIndex
			skippy, err := skipProtos(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthProtos
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *SignRequest) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowProtos
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: SignRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: SignRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrcntm wireType = %d for field Data", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowProtos
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthProtos
			}
			postIndex := iNdEx + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Data = append(m.Data[:0], dAtA[iNdEx:postIndex]...)
			if m.Data == nil {
				m.Data = []byte{}
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrcntm wireType = %d for field Seq", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowProtos
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthProtos
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Seq = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipProtos(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthProtos
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *SignResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowProtos
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: SignResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: SignResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrcntm wireType = %d for field Signature", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowProtos
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthProtos
			}
			postIndex := iNdEx + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Signature = append(m.Signature[:0], dAtA[iNdEx:postIndex]...)
			if m.Signature == nil {
				m.Signature = []byte{}
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrcntm wireType = %d for field Seq", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowProtos
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthProtos
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Seq = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipProtos(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthProtos
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *SetPrivKey) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowProtos
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: SetPrivKey: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: SetPrivKey: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrcntm wireType = %d for field PrivKey", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowProtos
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthProtos
			}
			postIndex := iNdEx + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.PrivKey = append(m.PrivKey[:0], dAtA[iNdEx:postIndex]...)
			if m.PrivKey == nil {
				m.PrivKey = []byte{}
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipProtos(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthProtos
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *VerifyRequest) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowProtos
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: VerifyRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: VerifyRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrcntm wireType = %d for field Signature", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowProtos
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthProtos
			}
			postIndex := iNdEx + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Signature = append(m.Signature[:0], dAtA[iNdEx:postIndex]...)
			if m.Signature == nil {
				m.Signature = []byte{}
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrcntm wireType = %d for field Data", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowProtos
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthProtos
			}
			postIndex := iNdEx + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Data = append(m.Data[:0], dAtA[iNdEx:postIndex]...)
			if m.Data == nil {
				m.Data = []byte{}
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrcntm wireType = %d for field PublicKey", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowProtos
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthProtos
			}
			postIndex := iNdEx + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.PublicKey = append(m.PublicKey[:0], dAtA[iNdEx:postIndex]...)
			if m.PublicKey == nil {
				m.PublicKey = []byte{}
			}
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrcntm wireType = %d for field Seq", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowProtos
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthProtos
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Seq = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipProtos(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthProtos
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *VerifyResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowProtos
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: VerifyResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: VerifyResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrcntm wireType = %d for field Seq", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowProtos
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthProtos
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Seq = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrcntm wireType = %d for field Result", wireType)
			}
			var v int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowProtos
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				v |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			m.Result = bool(v != 0)
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrcntm wireType = %d for field ErrorMsg", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowProtos
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthProtos
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ErrorMsg = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipProtos(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthProtos
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipProtos(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowProtos
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowProtos
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
			return iNdEx, nil
		case 1:
			iNdEx += 8
			return iNdEx, nil
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowProtos
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			iNdEx += length
			if length < 0 {
				return 0, ErrInvalidLengthProtos
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowProtos
					}
					if iNdEx >= l {
						return 0, io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					innerWire |= (uint64(b) & 0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				innerWireType := int(innerWire & 0x7)
				if innerWireType == 4 {
					break
				}
				next, err := skipProtos(dAtA[start:])
				if err != nil {
					return 0, err
				}
				iNdEx = start + next
			}
			return iNdEx, nil
		case 4:
			return iNdEx, nil
		case 5:
			iNdEx += 4
			return iNdEx, nil
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
	}
	panic("unreachable")
}

var (
	ErrInvalidLengthProtos = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowProtos   = fmt.Errorf("proto: integer overflow")
)

func init() {
	proto.RegisterFile("github.com/Ontology/eventbus/example/testRemoteCrypto/commons/protos.proto", fileDescriptorProtos)
}

var fileDescriptorProtos = []byte{
	// 341 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x7c, 0x91, 0xbf, 0x6e, 0xf2, 0x30,
	0x14, 0xc5, 0x63, 0x40, 0x40, 0xee, 0xc7, 0x57, 0x55, 0x1e, 0x2a, 0x54, 0x21, 0x0b, 0x65, 0xa8,
	0x18, 0x2a, 0x32, 0xb0, 0x77, 0xe8, 0x9f, 0xa5, 0x15, 0x2a, 0x32, 0x12, 0x7b, 0x82, 0x6e, 0xd3,
	0x48, 0x49, 0x1c, 0x6c, 0x07, 0x35, 0x5b, 0x1f, 0xa1, 0x8f, 0xd1, 0x47, 0xe9, 0xc8, 0xd8, 0xb1,
	0xa4, 0x4b, 0x47, 0x1e, 0xa1, 0x22, 0x38, 0xd0, 0xa1, 0xea, 0xe4, 0x7b, 0x8e, 0xee, 0xf1, 0xf9,
	0x59, 0x86, 0xdb, 0x20, 0xd4, 0x8f, 0x99, 0x3f, 0x9c, 0x8b, 0xd8, 0xbd, 0x4f, 0xb4, 0x88, 0x44,
	0x90, 0xbb, 0xb8, 0xc4, 0x44, 0xfb, 0x99, 0x72, 0xf1, 0xc9, 0x8b, 0xd3, 0x08, 0x5d, 0x8d, 0x4a,
	0x73, 0x8c, 0x85, 0xc6, 0x2b, 0x99, 0xa7, 0x5a, 0xb8, 0x73, 0x11, 0xc7, 0x22, 0x51, 0x6e, 0x2a,
	0x85, 0x16, 0x6a, 0x58, 0x1e, 0xb4, 0x65, 0x5c, 0xa7, 0x0d, 0x4d, 0x9e, 0x25, 0x63, 0x15, 0x38,
	0x23, 0xf8, 0x37, 0x0d, 0x83, 0x84, 0xe3, 0x22, 0x43, 0xa5, 0x29, 0x85, 0xc6, 0xb5, 0xa7, 0xbd,
	0x2e, 0xe9, 0x93, 0x41, 0x87, 0x97, 0x33, 0x3d, 0x86, 0xfa, 0x14, 0x17, 0xdd, 0x5a, 0x9f, 0x0c,
	0x6c, 0xbe, 0x1d, 0x9d, 0x0b, 0xe8, 0xec, 0x42, 0x2a, 0x15, 0x89, 0x42, 0xda, 0x03, 0x7b, 0xab,
	0x3d, 0x9d, 0x49, 0x34, 0xd1, 0x83, 0xf1, 0x4b, 0xfe, 0x0c, 0x60, 0x8a, 0x7a, 0x22, 0xc3, 0xe5,
	0x1d, 0xe6, 0xb4, 0x0b, 0x2d, 0x33, 0x9a, 0x6c, 0x25, 0x9d, 0x05, 0xfc, 0x9f, 0xa1, 0x0c, 0x1f,
	0xf2, 0x0a, 0xef, 0xef, 0xa2, 0x0a, 0xbe, 0xf6, 0x03, 0xbe, 0x07, 0xf6, 0x24, 0xf3, 0xa3, 0x70,
	0xbe, 0xbd, 0xbe, 0xbe, 0x4b, 0xec, 0x8d, 0x0a, 0xad, 0x71, 0x40, 0x9b, 0xc1, 0x51, 0x55, 0x69,
	0x1e, 0x67, 0x76, 0xc8, 0x7e, 0x87, 0x9e, 0x40, 0x93, 0xa3, 0xca, 0x22, 0x5d, 0x36, 0xb5, 0xb9,
	0x51, 0xf4, 0x14, 0xda, 0x37, 0x52, 0x0a, 0x39, 0x56, 0x41, 0x59, 0x65, 0xf3, 0xbd, 0xbe, 0x3c,
	0x5f, 0xad, 0x99, 0xf5, 0xbe, 0x66, 0xd6, 0x66, 0xcd, 0xc8, 0x73, 0xc1, 0xc8, 0x6b, 0xc1, 0xc8,
	0x5b, 0xc1, 0xc8, 0xaa, 0x60, 0xe4, 0xa3, 0x60, 0xe4, 0xab, 0x60, 0xd6, 0xa6, 0x60, 0xe4, 0xe5,
	0x93, 0x59, 0x7e, 0xb3, 0xfc, 0xaf, 0xd1, 0x77, 0x00, 0x00, 0x00, 0xff, 0xff, 0x77, 0x8c, 0xec,
	0x18, 0xfd, 0x01, 0x00, 0x00,
}
