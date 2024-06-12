package main

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"log"
	"math/big"
	"math/rand"
	"os"
	"strings"
	"time"
)

type Account struct {
	Address    string `json:"address"`
	PrivateKey string `json:"private_key"`
}

type Address struct {
	Address  string `json:"address"`
	Mnemonic string `json:"mnemonic"`
}

const (
	infuraURL      = "https://sepolia.infura.io/v3/9cac1960c91b4ca59a1e42882135c9c4"
	erc20Address   = "0xf29AbB93A56b7Fbe707203ea8DDcD008d07456fc"
	approveAddress = "0xf7Cb22834A5Fcb77cdbFb30a74FDb9FCc47dE155" // gra
)

var erc20ABI = `[{"constant":false,"inputs":[{"name":"_spender","type":"address"},{"name":"_value","type":"uint256"}],"name":"approve","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"}]`
var gravityABISend1 = `[{"inputs":[{"internalType":"address","name":"_tokenContract","type":"address"},{"internalType":"string","name":"_destination","type":"string"},{"internalType":"uint256","name":"_amount","type":"uint256"}],"name":"sendToCosmos","outputs":[],"stateMutability":"nonpayable","type":"function"}]`

var gravityABISend = `[{"inputs":[{"internalType":"address","name":"_tokenContract","type":"address"},{"internalType":"string","name":"_destination","type":"string"},{"internalType":"uint256","name":"_amount","type":"uint256"}],"name":"sendToCosmos","outputs":[],"stateMutability":"nonpayable","type":"function"}]`

// CreateEVMAddress create address
func CreateEVMAddress() {
	var accounts []Account
	var addresss []string

	for i := 0; i < 30; i++ {
		privateKey, err := crypto.GenerateKey()
		if err != nil {
			log.Fatalf("Failed to generate private key: %v", err)
		}

		privateKeyBytes := crypto.FromECDSA(privateKey)
		privateKeyHex := hex.EncodeToString(privateKeyBytes)

		publicKey := privateKey.Public()
		publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
		if !ok {
			log.Fatalf("Failed to cast public key to ECDSA")
		}

		address := crypto.PubkeyToAddress(*publicKeyECDSA).Hex()

		account := Account{
			Address:    address,
			PrivateKey: privateKeyHex,
		}
		addresss = append(addresss, address)
		accounts = append(accounts, account)
	}
	addressFile, err := os.Create("address.json")
	if err != nil {
		log.Fatalf("Failed to create input file: %v", err)
	}
	defer addressFile.Close()

	addressEncoder := json.NewEncoder(addressFile)
	addressEncoder.SetIndent("", "  ")
	if err := addressEncoder.Encode(addresss); err != nil {
		log.Fatalf("Failed to encode accounts to input.json: %v", err)
	}
	// 写入 input.json 文件
	inputFile, err := os.Create("input.json")
	if err != nil {
		log.Fatalf("Failed to create input file: %v", err)
	}
	defer inputFile.Close()

	inputEncoder := json.NewEncoder(inputFile)
	inputEncoder.SetIndent("", "  ")
	if err := inputEncoder.Encode(accounts[:15]); err != nil {
		log.Fatalf("Failed to encode accounts to input.json: %v", err)
	}

	// 写入 output.json 文件
	outputFile, err := os.Create("output.json")
	if err != nil {
		log.Fatalf("Failed to create output file: %v", err)
	}
	defer outputFile.Close()

	outputEncoder := json.NewEncoder(outputFile)
	outputEncoder.SetIndent("", "  ")
	if err := outputEncoder.Encode(accounts[15:]); err != nil {
		log.Fatalf("Failed to encode accounts to output.json: %v", err)
	}

	fmt.Println("100 EVM addresses and their private keys have been written to input.json and output.json")
}

// createCosmosAddress
func createCosmosAddress() {
	var accounts []Account

	for i := 0; i < 30; i++ {
		privKey := secp256k1.GenPrivKey()
		privKeyBytes := privKey.Bytes()
		privKeyHex := fmt.Sprintf("%X", privKeyBytes)

		pubKey := privKey.PubKey()
		address := sdk.AccAddress(pubKey.Address()).String()

		account := Account{
			Address:    address,
			PrivateKey: privKeyHex,
		}

		accounts = append(accounts, account)
	}

	// 写入 cosinput.json 文件
	inputFile, err := os.Create("cosinput.json")
	if err != nil {
		log.Fatalf("Failed to create input file: %v", err)
	}
	defer inputFile.Close()

	inputEncoder := json.NewEncoder(inputFile)
	inputEncoder.SetIndent("", "  ")
	if err := inputEncoder.Encode(accounts[:15]); err != nil {
		log.Fatalf("Failed to encode accounts to cosinput.json: %v", err)
	}

	// 写入 cosoutput.json 文件
	outputFile, err := os.Create("cosoutput.json")
	if err != nil {
		log.Fatalf("Failed to create output file: %v", err)
	}
	defer outputFile.Close()

	outputEncoder := json.NewEncoder(outputFile)
	outputEncoder.SetIndent("", "  ")
	if err := outputEncoder.Encode(accounts[15:]); err != nil {
		log.Fatalf("Failed to encode accounts to cosoutput.json: %v", err)
	}

	fmt.Println("30 Cosmos addresses and their private keys have been written to cosinput.json and cosoutput.json")
}

func main() {
	// 读取 cosinput.json 文件
	inputAccounts, err := readAccounts("input.json")
	if err != nil {
		log.Fatalf("Failed to read cosinput.json: %v", err)
	}

	// 读取 cosoutput.json 文件
	outputAccounts, err := readAccounts("output.json")
	if err != nil {
		log.Fatalf("Failed to read cosoutput.json: %v", err)
	}

	// 读取 me.json 文件
	me, err := readaddress("me.json")
	if err != nil {
		log.Fatalf("Failed to read cosoutput.json: %v", err)
	}
	fmt.Println(len(me))
	// 合并两个账户列表
	allAccounts := append(inputAccounts, outputAccounts...)
	client, err := ethclient.Dial(infuraURL)
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}

	parsedABI, err := abi.JSON(strings.NewReader(gravityABISend))
	if err != nil {
		log.Fatalf("Failed to parse gravityABISend ABI: %v", err)
	}

	erc20Address := common.HexToAddress(erc20Address)
	//spenderAddress := common.HexToAddress(approveAddress)
	gravityAddress := common.HexToAddress(approveAddress)
	amount := big.NewInt(2000000000000000000) // 1 token, assuming 18 decimal places

	for _, account := range allAccounts {
		// 使用当前时间的Unix时间戳作为种子值
		rand.Seed(time.Now().UnixNano())
		randomNumber := rand.Intn(10)
		privKey, err := crypto.HexToECDSA(account.PrivateKey)
		if err != nil {
			log.Fatalf("Failed to convert private key: %v", err)
		}

		auth, err := bind.NewKeyedTransactorWithChainID(privKey, big.NewInt(11155111)) // Mainnet chain ID is 1
		if err != nil {
			log.Fatalf("Failed to create transactor: %v", err)
		}

		nonce, err := client.PendingNonceAt(context.Background(), auth.From)
		if err != nil {
			log.Fatalf("Failed to get account nonce: %v", err)
		}

		gasPrice, err := client.SuggestGasPrice(context.Background())
		if err != nil {
			log.Fatalf("Failed to get gas price: %v", err)
		}

		auth.Nonce = big.NewInt(int64(nonce))
		auth.Value = big.NewInt(0)     // in wei
		auth.GasLimit = uint64(300000) // in units
		auth.GasPrice = gasPrice

		//func (_Gravity *GravityTransactor) SendToCosmos(opts *bind.TransactOpts, _tokenContract common.Address, _destination string, _amount *big.Int) (*types.Transaction, error) {
		input, err := parsedABI.Pack("sendToCosmos", erc20Address, me[randomNumber].Address, amount)
		if err != nil {
			log.Fatalf("Failed to pack input data: %v", err)
		}

		tx := types.NewTransaction(nonce, gravityAddress, big.NewInt(0), auth.GasLimit, auth.GasPrice, input)

		signedTx, err := auth.Signer(auth.From, tx)
		if err != nil {
			log.Fatalf("Failed to sign transaction: %v", err)
		}

		err = client.SendTransaction(context.Background(), signedTx)
		if err != nil {
			log.Fatalf("Failed to send transaction: %v", err)
		}

		fmt.Printf("Transaction sent: %s\n", signedTx.Hash().Hex())
	}
}

func Approve(client ethclient.Client, allAccounts []Account) (err error) {

	parsedABI, err := abi.JSON(strings.NewReader(erc20ABI))
	if err != nil {
		log.Fatalf("Failed to parse ERC-20 ABI: %v", err)
		return err
	}

	erc20Address := common.HexToAddress(erc20Address)
	spenderAddress := common.HexToAddress(approveAddress)
	amount := big.NewInt(9000000000000000000) // 1 token, assuming 18 decimal places

	for _, account := range allAccounts {
		privKey, err := crypto.HexToECDSA(account.PrivateKey)
		if err != nil {
			log.Fatalf("Failed to convert private key: %v", err)
			return err
		}

		auth, err := bind.NewKeyedTransactorWithChainID(privKey, big.NewInt(11155111)) // Mainnet chain ID is 1
		if err != nil {
			log.Fatalf("Failed to create transactor: %v", err)
			return err
		}

		nonce, err := client.PendingNonceAt(context.Background(), auth.From)
		if err != nil {
			log.Fatalf("Failed to get account nonce: %v", err)
			return err
		}

		gasPrice, err := client.SuggestGasPrice(context.Background())
		if err != nil {
			log.Fatalf("Failed to get gas price: %v", err)
			return err
		}

		auth.Nonce = big.NewInt(int64(nonce))
		auth.Value = big.NewInt(0)     // in wei
		auth.GasLimit = uint64(300000) // in units
		auth.GasPrice = gasPrice

		data, err := parsedABI.Pack("approve", spenderAddress, amount)
		if err != nil {
			log.Fatalf("Failed to pack ABI data: %v", err)
			return err
		}

		tx := types.NewTransaction(nonce, erc20Address, big.NewInt(0), auth.GasLimit, auth.GasPrice, data)

		signedTx, err := auth.Signer(auth.From, tx)
		if err != nil {
			log.Fatalf("Failed to sign transaction: %v", err)
			return err
		}

		err = client.SendTransaction(context.Background(), signedTx)
		if err != nil {
			log.Fatalf("Failed to send transaction: %v", err)
			return err
		}

		fmt.Printf("Transaction sent: %s\n", signedTx.Hash().Hex())
	}
	return err
}
func readAccounts(filename string) ([]Account, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	var accounts []Account
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&accounts); err != nil {
		return nil, fmt.Errorf("failed to decode json: %v", err)
	}

	return accounts, nil
}

func readaddress(filename string) ([]Address, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	var addresss []Address
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&addresss); err != nil {
		return nil, fmt.Errorf("failed to decode json: %v", err)
	}

	return addresss, nil
}
