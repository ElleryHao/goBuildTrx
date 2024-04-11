package SpecialSatsOp

import (
	"fmt"
	"math/big"
	"sort"
	"strconv"
	"strings"

	"github.com/mrtnetwork/bitcoin/address"

	"github.com/mrtnetwork/bitcoin/provider"

	"github.com/mrtnetwork/bitcoin/scripts"
)

var G_MaxTrxOutsNum int
var G_MaxBtcUtxo BTCUTXO

type BTCUTXO struct {
	Id     string
	Index  int
	Value  uint64
	Offset uint64
}

func FromStringToBtcUtxo(from string) *BTCUTXO {
	if from == string("") {
		return nil
	}
	templist := strings.Split(from, ":")
	if len(templist) != 2 {
		return nil
	}
	index, _ := strconv.Atoi(templist[1])
	return &BTCUTXO{
		Id:     templist[0],
		Index:  index,
		Value:  0,
		Offset: 0,
	}
}

func fixBtcUxtoArray(utxos []BTCUTXO) map[BTCUTXO][]uint64 {
	fmt.Println("fixBtcUxtoArray come")
	ret := map[BTCUTXO][]uint64{}
	for _, utxo := range utxos {
		btcUtxo := BTCUTXO{
			Id:     utxo.Id,
			Index:  utxo.Index,
			Value:  utxo.Value,
			Offset: 0,
		}
		ret[btcUtxo] = append(ret[btcUtxo], uint64(utxo.Offset))
	}
	fmt.Println("fixBtcUxtoArray end")
	return ret
}

func constructTrxOutByOffsets(offsetList []uint64, value uint64, from string, to string) []*scripts.TxOutput {
	fmt.Println("constructTrxOutByOffsets come")
	var maxOffset = value - 1
	fromAddr, _err := address.P2WPKHAddresssFromAddress(from)
	if _err != nil {
		fmt.Printf("address [%s] is invalid \n", from)
		return nil
	}
	toAddr, _ := address.P2WPKHAddresssFromAddress(to)
	sort.Slice(offsetList, func(i, j int) bool { return offsetList[i] < offsetList[j] })
	var trxOuts = []*scripts.TxOutput{}
	//slide windows,right not include but left value > G_MinimumSplitUtxo
	var left uint64 = 0
	var right uint64 = 0
	for _, offset := range offsetList {
		for {
			if right <= offset {
				right = right + uint64(G_MinimumSplitUtxo)
				continue
			}
			// three conditions
			//1.
			if right >= maxOffset {
				amount := maxOffset - left + 1
				trxOut := scripts.NewTxOutput(big.NewInt(int64(amount)), toAddr.Program().ToScriptPubKey())
				trxOuts = append(trxOuts, trxOut)
				return trxOuts
			}

			//2.
			if right+uint64(G_MinimumSplitUtxo) > maxOffset {
				//two conditions
				if left+uint64(G_MinimumSplitUtxo) < right {
					amount := right - uint64(G_MinimumSplitUtxo) - left
					trxOut1 := scripts.NewTxOutput(big.NewInt(int64(amount)), fromAddr.Program().ToScriptPubKey())

					left = right - uint64(G_MinimumSplitUtxo)
					amount = maxOffset - left + 1
					trxOut2 := scripts.NewTxOutput(big.NewInt(int64(amount)), toAddr.Program().ToScriptPubKey())

					trxOuts = append(trxOuts, trxOut1, trxOut2)
				} else {
					amount := maxOffset - left + 1
					trxOut := scripts.NewTxOutput(big.NewInt(int64(amount)), toAddr.Program().ToScriptPubKey())
					trxOuts = append(trxOuts, trxOut)
				}
				return trxOuts
			}

			//3.
			if right > offset && left <= offset {
				// build two txouts
				//1.
				amount := right - uint64(G_MinimumSplitUtxo) - left
				trxOut1 := scripts.NewTxOutput(big.NewInt(int64(amount)), fromAddr.Program().ToScriptPubKey())
				//2.
				trxOut2 := scripts.NewTxOutput(big.NewInt(int64(G_MinimumSplitUtxo)), toAddr.Program().ToScriptPubKey())
				left = right
				trxOuts = append(trxOuts, trxOut1, trxOut2)
				break
			}
		}

	}
	amount := maxOffset - left + 1
	trxOut := scripts.NewTxOutput(big.NewInt(int64(amount)), fromAddr.Program().ToScriptPubKey())
	trxOuts = append(trxOuts, trxOut)
	fmt.Println("constructTrxOutByOffsets end")
	return trxOuts
}

func constructTrxOutsByTrxIns(btcTrxInMap map[BTCUTXO][]uint64, from string, to string) map[BTCUTXO][]*scripts.TxOutput {
	btcTrxInTrxOutMap := map[BTCUTXO][]*scripts.TxOutput{}
	var outNums = 0
	for key, value := range btcTrxInMap {
		trxOuts := constructTrxOutByOffsets(value, key.Value, from, to)
		outNums += len(trxOuts)
		if outNums <= G_MaxTrxOutsNum {
			btcTrxInTrxOutMap[key] = append(btcTrxInTrxOutMap[key], trxOuts...)
		} else {
			break
		}
	}
	return btcTrxInTrxOutMap
}

func BuildBtcTrx(from string, to string, utxos []BTCUTXO) string {
	if G_MaxBtcUtxo.Index != -1 {
		fmt.Printf("utxo used to pay fees is [%s:%d]\n", G_MaxBtcUtxo.Id, G_MaxBtcUtxo.Index)
	} else {
		fmt.Println("there is no utox used to pay fees,pls wait")
		return ""
	}

	btcUtxoMap := fixBtcUxtoArray(utxos)
	btcFilterUtxoMap := constructTrxOutsByTrxIns(btcUtxoMap, from, to)

	var txIns []*scripts.TxInput
	var txOuts []*scripts.TxOutput
	for utxo, vouts := range btcFilterUtxoMap {
		txin := scripts.NewDefaultTxInput(utxo.Id, utxo.Index)
		txIns = append(txIns, txin)
		txOuts = append(txOuts, vouts...)
	}
	txIns = append(txIns, scripts.NewDefaultTxInput(G_MaxBtcUtxo.Id, G_MaxBtcUtxo.Index))
	tempTxOuts := txOuts[:]
	fromAddr, _ := address.P2WPKHAddresssFromAddress(from)
	tempTxOuts = append(tempTxOuts,
		scripts.NewTxOutput(big.NewInt(int64(G_MaxBtcUtxo.Value)), fromAddr.Program().ToScriptPubKey()))
	temp_tx := scripts.NewBtcTransaction(txIns, tempTxOuts, false)
	transactionSize := temp_tx.GetSize()

	network := address.MainnetNetwork
	api := provider.SelectApi(provider.BlockCyperApi, &network)
	networkFee, err := api.GetNetworkFee()
	var realFee *big.Int = big.NewInt(0)
	if err != nil {
		fmt.Println("cannot read network fee: ", err)
		fee_str := G_Config.Section("").Key("network_fee").Value()
		fee_int, _ := strconv.Atoi(fee_str)
		realFee.SetInt64(int64(fee_int))
	} else {
		realFee = networkFee.Low
	}

	total_fee := GetEstimate(transactionSize, realFee)
	exchange := int64(G_MaxBtcUtxo.Value) - total_fee.Int64()
	if exchange < 0 {
		fmt.Printf("fee is [%d] there is no enough exchange,pls wait", total_fee)
		return ""
	} else if exchange > 0 {
		fmt.Printf("total fee is [%d]", total_fee.Int64())
		txOuts = append(txOuts,
			scripts.NewTxOutput(big.NewInt(exchange), fromAddr.Program().ToScriptPubKey()))
	}
	tx := scripts.NewBtcTransaction(txIns, txOuts, false)
	fmt.Println(tx.GetTransactionDigest(0, fromAddr.Program().ToScriptPubKey()))

	return tx.Serialize()
}

func GetEstimate(trSize int, feeRate *big.Int) *big.Int {
	trSizeBigInt := new(big.Int).SetInt64(int64(trSize))
	return new(big.Int).Div(new(big.Int).Mul(trSizeBigInt, feeRate), big.NewInt(1024))
}
