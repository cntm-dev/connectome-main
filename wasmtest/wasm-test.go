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

package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"

	"github.com/cntmio/cntmology-crypto/keypair"
	"github.com/cntmio/cntmology/account"
	"github.com/cntmio/cntmology/cmd/utils"
	"github.com/cntmio/cntmology/common"
	"github.com/cntmio/cntmology/common/config"
	"github.com/cntmio/cntmology/common/constants"
	"github.com/cntmio/cntmology/common/log"
	"github.com/cntmio/cntmology/core/genesis"
	"github.com/cntmio/cntmology/core/ledger"
	"github.com/cntmio/cntmology/core/payload"
	"github.com/cntmio/cntmology/core/signature"
	"github.com/cntmio/cntmology/core/types"
	utils2 "github.com/cntmio/cntmology/core/utils"
	"github.com/cntmio/cntmology/events"
	common2 "github.com/cntmio/cntmology/http/base/common"
	"github.com/cntmio/cntmology/smartccntmract/service/wasmvm"
	"github.com/cntmio/cntmology/smartccntmract/states"
	vmtypes "github.com/cntmio/cntmology/vm/neovm/types"
	common3 "github.com/cntmio/cntmology/wasmtest/common"
	"github.com/cntmio/wagon/exec"
	"github.com/cntmio/wagon/wasm"
)

const ccntmractDir2 = "test-ccntmract"
const ccntmractDir = "testwasmdata"
const testcaseMethod = "testcase"

func NewDeployWasmCcntmract(signer *account.Account, code []byte) (*types.Transaction, error) {
	mutable, err := utils.NewDeployCodeTransaction(0, 100000000, code, payload.WASMVM_TYPE, "name", "version",
		"author", "email", "desc")
	if err != nil {
		return nil, err
	}
	err = utils.SignTransaction(signer, mutable)
	if err != nil {
		return nil, err
	}
	tx, err := mutable.IntoImmutable()
	return tx, err
}

func NewDeployNeoCcntmract(signer *account.Account, code []byte) (*types.Transaction, error) {
	mutable, err := utils.NewDeployCodeTransaction(0, 100000000, code, payload.NEOVM_TYPE, "name", "version",
		"author", "email", "desc")
	if err != nil {
		return nil, err
	}
	err = utils.SignTransaction(signer, mutable)
	if err != nil {
		return nil, err
	}
	tx, err := mutable.IntoImmutable()
	return tx, err
}

func NewDeployEvmCcntmract(opts *bind.TransactOpts, code []byte, jsonABI string, params ...interface{}) (*types2.Transaction, error) {
	parsed, err := abi.JSON(strings.NewReader(jsonABI))
	checkErr(err)
	input, err := parsed.Pack("", params...)
	checkErr(err)
	input = append(code, input...)
	deployTx := types2.NewCcntmractCreation(opts.Nonce.Uint64(), opts.Value, opts.GasLimit, opts.GasPrice, input)
	signedTx, err := opts.Signer(opts.From, deployTx)
	checkErr(err)
	return signedTx, err
}

func GenNeoTextCaseTransaction(ccntmract common.Address, database *ledger.Ledger) [][]common3.TestCase {
	params := make([]interface{}, 0)
	method := string("testcase")
	// neovm entry api is def Main(method, args). and testcase method api need no other args, so pass a random args to entry api.
	operation := 1
	params = append(params, method)
	params = append(params, operation)
	tx, err := common2.NewNeovmInvokeTransaction(0, 100000000, ccntmract, params)
	imt, err := tx.IntoImmutable()
	if err != nil {
		panic(err)
	}
	res, err := database.PreExecuteCcntmract(imt)
	if err != nil {
		panic(err)
	}

	ret := res.Result.(string)
	jsonCase, err := common.HexToBytes(ret)

	if err != nil {
		panic(err)
	}
	if len(jsonCase) == 0 {
		panic("failed to get testcase data from ccntmract")
	}
	var testCase [][]common3.TestCase
	err = json.Unmarshal([]byte(jsonCase), &testCase)
	if err != nil {
		panic("failed Unmarshal")
	}
	return testCase
}

func ExactTestCase(code []byte) [][]common3.TestCase {
	m, err := wasm.ReadModule(bytes.NewReader(code), func(name string) (*wasm.Module, error) {
		switch name {
		case "env":
			return wasmvm.NewHostModule(), nil
		}
		return nil, fmt.Errorf("module %q unknown", name)
	})
	checkErr(err)

	compiled, err := exec.CompileModule(m)
	checkErr(err)

	vm, err := exec.NewVMWithCompiled(compiled, 10*1024*1024)
	checkErr(err)

	param := common.NewZeroCopySink(nil)
	param.WriteString(testcaseMethod)
	host := &wasmvm.Runtime{Input: param.Bytes()}
	vm.HostData = host
	vm.RecoverPanic = true
	envGasLimit := uint64(100000000000000)
	envExecStep := uint64(100000000000000)
	vm.ExecMetrics = &exec.Gas{GasLimit: &envGasLimit, GasPrice: 0, GasFactor: 5, ExecStep: &envExecStep}
	vm.CallStackDepth = 1024

	entry := compiled.RawModule.Export.Entries["invoke"]
	index := int64(entry.Index)
	_, err = vm.ExecCode(index)
	checkErr(err)

	var testCase [][]common3.TestCase
	source := common.NewZeroCopySource(host.Output)
	jsonCase, _, _, _ := source.NextString()

	if len(jsonCase) == 0 {
		panic("failed to get testcase data from ccntmract")
	}

	err = json.Unmarshal([]byte(jsonCase), &testCase)
	checkErr(err)

	return testCase
}

func LoadCcntmracts(dir string) (map[string][]byte, error) {
	ccntmracts := make(map[string][]byte)
	fnames, err := filepath.Glob(filepath.Join(dir, "*"))
	if err != nil {
		return nil, err
	}
	for _, name := range fnames {
		if !(strings.HasSuffix(name, ".wasm") || strings.HasSuffix(name, ".avm")) {
			ccntminue
		}
		raw, err := ioutil.ReadFile(name)
		if err != nil {
			return nil, err
		}
		ccntmracts[path.Base(name)] = raw
	}

	return ccntmracts, nil
}

func loadCcntmract(filePath string) []byte {
	if common.FileExisted(filePath) {
		raw, err := ioutil.ReadFile(filePath)
		checkErr(err)
		code, err := hex.DecodeString(strings.TrimSpace(string(raw)))
		if err != nil {
			return raw
		} else {
			return code
		}
	} else {
		panic("no existed file:" + filePath)
	}
}

func init() {
	log.InitLog(log.InfoLog, log.PATH, log.Stdout)
	runtime.GOMAXPROCS(4)
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func execTxCheckRes(tx *types.Transaction, testCase common3.TestCase, database *ledger.Ledger, addr common.Address, acct *account.Account) {
	res, err := database.PreExecuteCcntmract(tx)
	log.Infof("testcase consume gas: %d", res.Gas)
	checkErr(err)

	height := database.GetCurrentBlockHeight()
	header, err := database.GetHeaderByHeight(height)
	checkErr(err)
	blockTime := header.Timestamp + 1

	execEnv := ExecEnv{Time: blockTime, Height: height + 1, Tx: tx, BlockHash: header.Hash(), Ccntmract: addr}
	checkExecResult(testCase, res, execEnv)

	block, _ := makeBlock(acct, []*types.Transaction{tx})
	err = database.AddBlock(block, nil, common.UINT256_EMPTY)
	checkErr(err)
	log.Infof("execTxCheckRes success: %s", testCase.Method)
}

func main() {
	datadir := "testdata"
	err := os.RemoveAll(datadir)
	defer func() {
		_ = os.RemoveAll(datadir)
		_ = os.RemoveAll(log.PATH)
	}()
	checkErr(err)
	log.Trace("Node version: ", config.Version)

	acct := account.NewAccount("")
	buf := keypair.SerializePublicKey(acct.PublicKey)
	config.DefConfig.Genesis.ConsensusType = "solo"
	config.DefConfig.Genesis.SOLO.GenBlockTime = 3
	config.DefConfig.Genesis.SOLO.Bookkeepers = []string{hex.EncodeToString(buf)}
	config.DefConfig.P2PNode.NetworkId = 3

	bookkeepers := []keypair.PublicKey{acct.PublicKey}
	//Init event hub
	events.Init()

	log.Info("1. Loading the Ledger")
	genblock, err := genesis.BuildGenesisBlock(bookkeepers, config.DefConfig.Genesis)
	checkErr(err)
	database, err := ledger.InitLedger(datadir, 1000000, bookkeepers, genblock)
	checkErr(err)
	ledger.DefLedger = database

	log.Info("loading wasm ccntmract")
	ccntmract, err := LoadCcntmracts(ccntmractDir)
	checkErr(err)

	log.Infof("deploying %d wasm ccntmracts", len(ccntmract))
	txes := make([]*types.Transaction, 0, len(ccntmract))
	for file, ccntm := range ccntmract {
		var tx *types.Transaction
		var err error
		if strings.HasSuffix(file, ".wasm") {
			tx, err = NewDeployWasmCcntmract(acct, ccntm)
		} else if strings.HasSuffix(file, ".avm") {
			tx, err = NewDeployNeoCcntmract(acct, ccntm)
		}

		checkErr(err)

		res, err := database.PreExecuteCcntmract(tx)
		log.Infof("deploy %s consume gas: %d", file, res.Gas)
		checkErr(err)
		txes = append(txes, tx)
	}

	block, _ := makeBlock(acct, txes)
	err = database.AddBlock(block, common.UINT256_EMPTY)
	checkErr(err)

	addrMap := make(map[string]common.Address)
	for file, code := range ccntmract {
		addrMap[path.Base(file)] = common.AddressFromVmCode(code)
	}

	testCcntmext := common3.TestCcntmext{
		Admin:   acct.Address,
		AddrMap: addrMap,
	}

	for file, ccntm := range ccntmract {
		log.Infof("exacting testcase from %s", file)
		addr := common.AddressFromVmCode(ccntm)
		if strings.HasSuffix(file, ".avm") {
			testCases := GenNeoTextCaseTransaction(addr, database)
			for _, testCase := range testCases[0] { // only handle group 0 currently
				val, _ := json.Marshal(testCase)
				log.Info("executing testcase: ", string(val))
				tx, err := common3.GenNeoVMTransaction(testCase, addr, &testCcntmext)
				checkErr(err)
				execTxCheckRes(tx, testCase, database, addr, acct)
			}
		} else if strings.HasSuffix(file, ".wasm") {
			testCases := ExactTestCase(ccntm)
			for _, testCase := range testCases[0] { // only handle group 0 currently
				val, _ := json.Marshal(testCase)
				log.Info("executing testcase: ", string(val))
				tx, err := common3.GenWasmTransaction(testCase, addr, &testCcntmext)
				checkErr(err)

				execTxCheckRes(tx, testCase, database, addr, acct)
			}
		}
	}

	log.Info("ccntmract test succeed")
}

type ExecEnv struct {
	Ccntmract  common.Address
	Time      uint32
	Height    uint32
	Tx        *types.Transaction
	BlockHash common.Uint256
}

func checkExecResult(testCase common3.TestCase, result *states.PreExecResult, execEnv ExecEnv) {
	assertEq(result.State, byte(1))
	if execEnv.Tx.IsEipTx() {
		res := parseEthResult(testCase.Method, result.Result, testCase.JsonAbi)
		compareEthResult(res, testCase.Expect)
		return
	}
	ret := result.Result.(string)
	switch testCase.Method {
	case "timestamp":
		sink := common.NewZeroCopySink(nil)
		sink.WriteUint64(uint64(execEnv.Time))
		assertEq(ret, hex.EncodeToString(sink.Bytes()))
	case "block_height":
		sink := common.NewZeroCopySink(nil)
		sink.WriteUint32(uint32(execEnv.Height))
		assertEq(ret, hex.EncodeToString(sink.Bytes()))
	case "self_address", "entry_address":
		assertEq(ret, hex.EncodeToString(execEnv.Ccntmract[:]))
	case "caller_address":
		assertEq(ret, hex.EncodeToString(common.ADDRESS_EMPTY[:]))
	case "current_txhash":
		hash := execEnv.Tx.Hash()
		assertEq(ret, hex.EncodeToString(hash[:]))
	case "current_blockhash":
		assertEq(ret, hex.EncodeToString(execEnv.BlockHash[:]))
	//case "sha256":
	//	let data :&[u8]= source.read().unwrap();
	//	sink.write(runtime::sha256(&data))
	//}
	default:
		if len(testCase.Expect) != 0 {
			expect, err := utils.ParseParams(testCase.Expect)
			checkErr(err)
			if execEnv.Tx.TxType == types.InvokeNeo {
				val := buildNeoVmValueFromExpect(expect)
				cv, err := val.ConvertNeoVmValueHexString()
				checkErr(err)
				assertEq(cv, result.Result)
			} else if execEnv.Tx.TxType == types.InvokeWasm {
				exp, err := utils2.BuildWasmCcntmractParam(expect)
				checkErr(err)
				assertEq(ret, hex.EncodeToString(exp))
			} else {
				panic("error tx type")
			}
		}
		if len(testCase.Notify) != 0 {
			js, _ := json.Marshal(result.Notify)
			assertEq(true, strings.Ccntmains(string(js), testCase.Notify))
		}
	}
}

func compareEthResult(result interface{}, expect string) {
	res := strings.Split(expect, ":")
	switch res[0] {
	case "int":
		data := result.(*big.Int)
		exp, err := strconv.ParseUint(res[1], 10, 64)
		checkErr(err)
		if data.Uint64() != exp {
			panic(data)
		}
	case "bool":
		data := result.(bool)
		if res[1] == "true" {
			if !data {
				panic(data)
			}
		} else {
			if data {
				panic(data)
			}
		}
	}
}

func buildNeoVmValueFromExpect(expectlist []interface{}) *vmtypes.VmValue {
	if len(expectlist) > 1 {
		panic("only support return one value")
	}
	expect := expectlist[0]

	switch expect.(type) {
	case string:
		val, err := vmtypes.VmValueFromBytes([]byte(expect.(string)))
		if err != nil {
			panic(err)
		}
		return &val
	case []byte:
		val, err := vmtypes.VmValueFromBytes(expect.([]byte))
		if err != nil {
			panic(err)
		}
		return &val
	case int64:
		val := vmtypes.VmValueFromInt64(expect.(int64))
		return &val
	case bool:
		val := vmtypes.VmValueFromBool(expect.(bool))
		return &val
	case common.Address:
		addr := expect.(common.Address)
		val, err := vmtypes.VmValueFromBytes(addr[:])
		if err != nil {
			panic(err)
		}
		return &val
	default:
		fmt.Printf("unspport param type %s", reflect.TypeOf(expect))
		panic("unspport param type")
	}
}

func GenAccounts(num int) []*account.Account {
	var accounts []*account.Account
	for i := 0; i < num; i++ {
		acc := account.NewAccount("")
		accounts = append(accounts, acc)
	}
	return accounts
}

func makeBlock(acc *account.Account, txs []*types.Transaction) (*types.Block, error) {
	nextBookkeeper, err := types.AddressFromBookkeepers([]keypair.PublicKey{acc.PublicKey})
	if err != nil {
		return nil, fmt.Errorf("GetBookkeeperAddress error:%s", err)
	}
	prevHash := ledger.DefLedger.GetCurrentBlockHash()
	height := ledger.DefLedger.GetCurrentBlockHeight()

	nonce := uint64(height)
	var txHash []common.Uint256
	for _, t := range txs {
		txHash = append(txHash, t.Hash())
	}

	txRoot := common.ComputeMerkleRoot(txHash)

	blockRoot := ledger.DefLedger.GetBlockRootWithNewTxRoots(height+1, []common.Uint256{txRoot})
	header := &types.Header{
		Version:          0,
		PrevBlockHash:    prevHash,
		TransactionsRoot: txRoot,
		BlockRoot:        blockRoot,
		Timestamp:        constants.GENESIS_BLOCK_TIMESTAMP + height + 1,
		Height:           height + 1,
		ConsensusData:    nonce,
		NextBookkeeper:   nextBookkeeper,
	}
	block := &types.Block{
		Header:       header,
		Transactions: txs,
	}

	blockHash := block.Hash()

	sig, err := signature.Sign(acc, blockHash[:])
	if err != nil {
		return nil, fmt.Errorf("signature, Sign error:%s", err)
	}

	block.Header.Bookkeepers = []keypair.PublicKey{acc.PublicKey}
	block.Header.SigData = [][]byte{sig}
	return block, nil
}

func assertEq(a interface{}, b interface{}) {
	if reflect.DeepEqual(a, b) == false {
		panic(fmt.Sprintf("not equal: a= %v, b=%v", a, b))
	}
}

func JsonString(v interface{}) string {
	buf, err := json.MarshalIndent(v, "", "  ")
	checkErr(err)

	return string(buf)
}
