package test

import (
	"testing"
	"bytes"
	"fmt"
	"os"
	"github.com/Ontology/core/types"
	"github.com/Ontology/common"
	. "github.com/Ontology/http/base/common"
	"github.com/Ontology/crypto"
	"math/big"
	"github.com/Ontology/smartccntmract/service/native/states"
	"github.com/Ontology/common/serialization"
	"time"
	"github.com/Ontology/vm/neovm"
	"github.com/Ontology/core/utils"
	"github.com/Ontology/account"
	vmtypes "github.com/Ontology/vm/types"
	ctypes "github.com/Ontology/core/types"
	. "github.com/Ontology/common"
)

func TestTxDeserialize(t *testing.T) {
	bys, _ := common.HexToBytes("")
	var txn types.Transaction
	if err := txn.Deserialize(bytes.NewReader(bys)); err != nil {
		fmt.Print("Deserialize Err:", err)
		os.Exit(0)
	}
	fmt.Printf("TxType:%x\n", txn.TxType)
	os.Exit(0)
}
func TestAddress(t *testing.T) {
	crypto.SetAlg("")
	pubkey, _ := common.HexToBytes("0399b851bc2cd05506d6821d4bc5a92139b00ac4bc7399cd9ca0aac86a468d1c05")
	pk, err := crypto.DecodePoint(pubkey)
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
	ui60 := types.AddressFromPubKey(pk)
	addr := common.ToHexString(ui60.ToArray())
	fmt.Println(addr)
	fmt.Println(ui60.ToBase58())
}
func TestTransfer(t *testing.T) {
	ccntmract := "ff00000000000000000000000000000000000001"

	ct, _ := common.HexToBytes(ccntmract)
	ctu, _ := common.Uint160ParseFromBytes(ct)
	from := "0121dca8ffcba308e697ee9e734ce686f4181658"

	f, _ := common.HexToBytes(from)
	fu, _ := common.Uint160ParseFromBytes(f)
	to := "01c6d97beeb85c7fef8cea8edd564f52c0236fb1"

	tt, _ := common.HexToBytes(to)
	tu, _ := common.Uint160ParseFromBytes(tt)

	var sts []*states.State
	sts = append(sts, &states.State{
		From:  fu,
		To:    tu,
		Value: big.NewInt(100),
	})
	transfers := new(states.Transfers)
	fmt.Println("ctu:", ctu)
	transfers.Params = append(transfers.Params, &states.TokenTransfer{
		Ccntmract: ctu,
		States:   sts,
	})

	bf := new(bytes.Buffer)
	if err := serialization.WriteVarBytes(bf, []byte("Token.Common.Transfer")); err != nil {
		fmt.Println("Serialize transfer falg error.")
		os.Exit(1)
	}
	if err := transfers.Serialize(bf); err != nil {
		fmt.Println("Serialize transfers struct error.")
		os.Exit(1)
	}

	fmt.Println(common.ToHexString(bf.Bytes()))

}

func TestInvokefunction(t *testing.T) {
	var funcName string
	builder := neovm.NewParamsBuilder(new(bytes.Buffer))
	err := BuildSmartCcntmractParamInter(builder, []interface{}{funcName, "",""})
	if err != nil {
	}
	codeParams := builder.ToArray();
	tx := utils.NewInvokeTransaction(vmtypes.VmCode{
		VmType: vmtypes.NativeVM,
		Code: codeParams,
	})
	tx.Nonce = uint32(time.Now().Unix())

	acct := account.Open(account.WalletFileName, []byte("passwordtest"))
	acc, err := acct.GetDefaultAccount(); if err != nil {
		fmt.Println("GetDefaultAccount error:", err)
		os.Exit(1)
	}
	hash := tx.Hash()
	sign, _ := crypto.Sign(acc.PrivateKey, hash[:])
	tx.Sigs = append(tx.Sigs, &ctypes.Sig{
		PubKeys: []*crypto.PubKey{acc.PublicKey},
		M:       1,
		SigData: [][]byte{sign},
	})

	txbf := new(bytes.Buffer)
	if err := tx.Serialize(txbf); err != nil {
		fmt.Println("Serialize transaction error.")
		os.Exit(1)
	}
	common.ToHexString(txbf.Bytes())
}
func BuildSmartCcntmractParamInter(builder *neovm.ParamsBuilder, smartCcntmractParams []interface{}) error {
	//虚拟机参数入栈时会反序
	for i := len(smartCcntmractParams) - 1; i >= 0; i-- {
		switch v := smartCcntmractParams[i].(type) {
		case bool:
			builder.EmitPushBool(v)
		case int:
			builder.EmitPushInteger(big.NewInt(int64(v)))
		case uint:
			builder.EmitPushInteger(big.NewInt(int64(v)))
		case int32:
			builder.EmitPushInteger(big.NewInt(int64(v)))
		case uint32:
			builder.EmitPushInteger(big.NewInt(int64(v)))
		case int64:
			builder.EmitPushInteger(big.NewInt(int64(v)))
		case Fixed64:
			builder.EmitPushInteger(big.NewInt(int64(v.GetData())))
		case uint64:
			val := big.NewInt(0)
			builder.EmitPushInteger(val.SetUint64(uint64(v)))
		case string:
			builder.EmitPushByteArray([]byte(v))
		case *big.Int:
			builder.EmitPushInteger(v)
		case []byte:
			builder.EmitPushByteArray(v)
		case []interface{}:
			err := BuildSmartCcntmractParamInter(builder, v)
			if err != nil {
				return err
			}
			builder.EmitPushInteger(big.NewInt(int64(len(v))))
			builder.Emit(neovm.PACK)
		default:
			return fmt.Errorf("unsupported param:%s", v)
		}
	}
	return nil
}