package asset

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	. "DNA/cli/common"
	"DNA/account"
	. "DNA/common"
	. "DNA/core/asset"
	"DNA/core/ccntmract"
	"DNA/core/signature"
	"DNA/core/transaction"
	"DNA/net/httpjsonrpc"

	"github.com/urfave/cli"
)

const (
	RANDBYTELEN    = 4
	REFERTXHASHLEN = 64
)

func newCcntmractCcntmextWithoutProgramHashes(data signature.SignableData) *ccntmract.CcntmractCcntmext {
	return &ccntmract.CcntmractCcntmext{
		Data:       data,
		Codes:      make([][]byte, 1),
		Parameters: make([][][]byte, 1),
	}
}

func openWallet(name string, passwd []byte) account.Client {
	if name == account.WalletFileName {
		fmt.Println("Using default wallet: ", account.WalletFileName)
	}
	wallet := account.Open(name, passwd)
	if wallet == nil {
		fmt.Println("Failed to open wallet: ", name)
		os.Exit(1)
	}
	return wallet
}

func getUintHash(programHashStr, assetHashStr string) (Uint160, Uint256, error) {
	programHashHex, err := hex.DecodeString(programHashStr)
	if err != nil {
		fmt.Println("Decoding program hash string failed")
		return Uint160{}, Uint256{}, err
	}
	var programHash Uint160
	if err := programHash.Deserialize(bytes.NewReader(programHashHex)); err != nil {
		fmt.Println("Deserialization program hash failed")
		return Uint160{}, Uint256{}, err
	}
	assetHashHex, err := hex.DecodeString(assetHashStr)
	if err != nil {
		fmt.Println("Decoding asset hash string failed")
		return Uint160{}, Uint256{}, err
	}
	var assetHash Uint256
	if err := assetHash.Deserialize(bytes.NewReader(assetHashHex)); err != nil {
		fmt.Println("Deserialization asset hash failed")
		return Uint160{}, Uint256{}, err
	}
	return programHash, assetHash, nil
}

func signTransaction(signer *account.Account, tx *transaction.Transaction) error {
	signature, err := signature.SignBySigner(tx, signer)
	if err != nil {
		fmt.Println("SignBySigner failed")
		return err
	}
	transactionCcntmract, err := ccntmract.CreateSignatureCcntmract(signer.PubKey())
	if err != nil {
		fmt.Println("CreateSignatureCcntmract failed")
		return err
	}
	transactionCcntmractCcntmext := newCcntmractCcntmextWithoutProgramHashes(tx)
	if err := transactionCcntmractCcntmext.AddCcntmract(transactionCcntmract, signer.PubKey(), signature); err != nil {
		fmt.Println("AddCcntmract failed")
		return err
	}
	tx.SetPrograms(transactionCcntmractCcntmext.GetPrograms())
	return nil
}

func makeRegTransaction(admin, issuer *account.Account, name string, value Fixed64) (string, error) {
	asset := &Asset{name, byte(0x00), AssetType(Share), UTXO}
	transactionCcntmract, err := ccntmract.CreateSignatureCcntmract(admin.PubKey())
	if err != nil {
		fmt.Println("CreateSignatureCcntmract failed")
		return "", err
	}
	tx, _ := transaction.NewRegisterAssetTransaction(asset, value, issuer.PubKey(), transactionCcntmract.ProgramHash)
	tx.Nonce = uint64(rand.Int63())
	if err := signTransaction(issuer, tx); err != nil {
		fmt.Println("sign regist transaction failed")
		return "", err
	}
	var buffer bytes.Buffer
	if err := tx.Serialize(&buffer); err != nil {
		fmt.Println("serialize registtransaction failed")
		return "", err
	}
	return hex.EncodeToString(buffer.Bytes()), nil
}

func makeIssueTransaction(issuer *account.Account, programHashStr, assetHashStr string, value Fixed64) (string, error) {
	programHash, assetHash, err := getUintHash(programHashStr, assetHashStr)
	if err != nil {
		return "", err
	}
	issueTxOutput := &transaction.TxOutput{
		AssetID:     assetHash,
		Value:       value,
		ProgramHash: programHash,
	}
	outputs := []*transaction.TxOutput{issueTxOutput}
	tx, _ := transaction.NewIssueAssetTransaction(outputs)
	tx.Nonce = uint64(rand.Int63())
	if err := signTransaction(issuer, tx); err != nil {
		fmt.Println("sign issue transaction failed")
		return "", err
	}
	var buffer bytes.Buffer
	if err := tx.Serialize(&buffer); err != nil {
		fmt.Println("serialization of issue transaction failed")
		return "", err
	}
	return hex.EncodeToString(buffer.Bytes()), nil
}

func makeTransferTransaction(signer *account.Account, programHashStr, assetHashStr string, value Fixed64) (string, error) {
	programHash, assetHash, err := getUintHash(programHashStr, assetHashStr)
	if err != nil {
		return "", err
	}
	myProgramHashStr := ToHexString(signer.ProgramHash.ToArray())

	resp, err := httpjsonrpc.Call(Address(), "getunspendoutput", 0, []interface{}{myProgramHashStr, assetHashStr})
	if err != nil {
		fmt.Println("HTTP JSON call failed")
		return "", err
	}
	r := make(map[string]interface{})
	err = json.Unmarshal(resp, &r)
	if err != nil {
		fmt.Println("Unmarshal JSON failed")
		return "", err
	}

	inputs := []*transaction.UTXOTxInput{}
	outputs := []*transaction.TxOutput{}
	transferTxOutput := &transaction.TxOutput{
		AssetID:     assetHash,
		Value:       value,
		ProgramHash: programHash,
	}
	outputs = append(outputs, transferTxOutput)

	unspend := r["result"].(map[string]interface{})
	expected := transferTxOutput.Value
	for k, v := range unspend {
		h := k[0:REFERTXHASHLEN]
		i := k[REFERTXHASHLEN+1:]
		b, _ := hex.DecodeString(h)
		var referHash Uint256
		referHash.Deserialize(bytes.NewReader(b))
		referIndex, _ := strconv.Atoi(i)

		out := v.(map[string]interface{})
		value := Fixed64(out["Value"].(float64))
		if value == expected {
			transferUTXOInput := &transaction.UTXOTxInput{
				ReferTxID:          referHash,
				ReferTxOutputIndex: uint16(referIndex),
			}
			expected = 0
			inputs = append(inputs, transferUTXOInput)
			break
		} else if value > expected {
			transferUTXOInput := &transaction.UTXOTxInput{
				ReferTxID:          referHash,
				ReferTxOutputIndex: uint16(referIndex),
			}
			inputs = append(inputs, transferUTXOInput)
			getChangeOutput := &transaction.TxOutput{
				AssetID:     assetHash,
				Value:       value - expected,
				ProgramHash: signer.ProgramHash,
			}
			expected = 0
			outputs = append(outputs, getChangeOutput)
			break
		} else if value < expected {
			transferUTXOInput := &transaction.UTXOTxInput{
				ReferTxID:          referHash,
				ReferTxOutputIndex: uint16(referIndex),
			}
			expected -= value
			inputs = append(inputs, transferUTXOInput)
			if expected == 0 {
				break
			}
		}
	}
	if expected != 0 {
		return "", errors.New("transfer failed, ammount is not enough")
	}
	tx, _ := transaction.NewTransferAssetTransaction(inputs, outputs)
	tx.Nonce = uint64(rand.Int63())
	if err := signTransaction(signer, tx); err != nil {
		fmt.Println("sign transfer transaction failed")
		return "", err
	}
	var buffer bytes.Buffer
	if err := tx.Serialize(&buffer); err != nil {
		fmt.Println("serialization of transfer transaction failed")
		return "", err
	}
	return hex.EncodeToString(buffer.Bytes()), nil
}

func assetAction(c *cli.Ccntmext) error {
	if c.NumFlags() == 0 {
		cli.ShowSubcommandHelp(c)
		return nil
	}
	reg := c.Bool("reg")
	issue := c.Bool("issue")
	transfer := c.Bool("transfer")
	if !reg && !issue && !transfer {
		cli.ShowSubcommandHelp(c)
		return nil
	}

	wallet := openWallet(c.String("wallet"), []byte(c.String("password")))
	admin, _ := wallet.GetDefaultAccount()
	value := c.Int64("value")
	if value == 0 {
		fmt.Println("invalid value [--value]")
		return nil
	}

	var txHex string
	var err error
	if reg {
		name := c.String("name")
		if name == "" {
			rbuf := make([]byte, RANDBYTELEN)
			rand.Read(rbuf)
			name = "DNA-" + ToHexString(rbuf)
		}
		issuer := admin
		txHex, err = makeRegTransaction(admin, issuer, name, Fixed64(value))
	} else {
		asset := c.String("asset")
		to := c.String("to")
		if asset == "" || to == "" {
			fmt.Println("missing flag [--asset] or [--to]")
			return nil
		}
		if issue {
			txHex, err = makeIssueTransaction(admin, to, asset, Fixed64(value))
		} else if transfer {
			txHex, err = makeTransferTransaction(admin, to, asset, Fixed64(value))
		}
		if err != nil {
			fmt.Println(err)
			return nil
		}
	}
	resp, err := httpjsonrpc.Call(Address(), "sendrawtransaction", 0, []interface{}{txHex})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	FormatOutput(resp)

	return nil
}

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:        "asset",
		Usage:       "asset registration, issuance and transfer",
		Description: "With nodectl asset, you could ccntmrol assert through transaction.",
		ArgsUsage:   "[args]",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "reg, r",
				Usage: "regist a new kind of asset",
			},
			cli.BoolFlag{
				Name:  "issue, i",
				Usage: "issue asset that has been registered",
			},
			cli.BoolFlag{
				Name:  "transfer, t",
				Usage: "transfer asset",
			},
			cli.StringFlag{
				Name:  "wallet, w",
				Usage: "wallet name",
				Value: account.WalletFileName,
			},
			cli.StringFlag{
				Name:  "password, p",
				Usage: "wallet password",
				Value: account.DefaultPin,
			},
			cli.StringFlag{
				Name:  "asset, a",
				Usage: "uniq id for asset",
			},
			cli.StringFlag{
				Name:  "name",
				Usage: "asset name",
			},
			cli.StringFlag{
				Name:  "to",
				Usage: "asset to whom",
			},
			cli.Int64Flag{
				Name:  "value, v",
				Usage: "asset ammount",
			},
		},
		Action: assetAction,
		OnUsageError: func(c *cli.Ccntmext, err error, isSubcommand bool) error {
			PrintError(c, err, "asset")
			return cli.NewExitError("", 1)
		},
	}
}