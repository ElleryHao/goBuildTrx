package SpecialSatsOp

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"gopkg.in/ini.v1"
)

var G_Config *ini.File
var G_SatsTypes []string
var G_MinimumSplitUtxo int

type SATS struct {
	Sats   []uint64 `json:"sats"`
	Name   string   `json:"name"`
	Block  int      `json:"block"`
	Time   string   `json:"time"`
	Offset uint64   `json:"offset"`
	Types  []string `json:"types"`
}

type UTXOS struct {
	Is_safe bool   `json:"is_safe"`
	Sats    []SATS `json:"sats"`
	Id      string `json:"id"`
	Value   uint64 `json:"value"`
}

var request_url = "https://gw.sating.io/api/account/sats/"

func GetRspFromScanReq(addr string) []byte {
	var url = request_url + addr
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println(err)
		return []byte{}
	}
	c := http.Client{
		Timeout: time.Second * 2 * 60,
	}
	res, err := c.Do(req)
	if err != nil {
		fmt.Println(err)
		return []byte{}
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return []byte{}
	}
	fmt.Println(res)
	return body
}

func ScanAddrForSpecial(addr string) []BTCUTXO {
	fmt.Println(addr)
	G_MaxBtcUtxo = BTCUTXO{
		Id:    "",
		Index: -1,
		Value: 0,
	}
	data := GetRspFromScanReq(addr)
	if len(data) == 0 {
		return nil
	}
	var utxos []UTXOS
	err := json.Unmarshal(data, &utxos)
	if err != nil {
		fmt.Println("Error:", err)
		return nil
	}
	var btcUtxos []BTCUTXO
	for _, utxo := range utxos {
		for _, sat := range utxo.Sats {
			sort.Strings(sat.Types)
			for _, sats_type := range G_SatsTypes {
				index := sort.SearchStrings(sat.Types, sats_type)
				if index < len(sat.Types) && sat.Types[index] == sats_type {
					btcUtxo := FromStringToBtcUtxo(utxo.Id)
					btcUtxo.Offset = sat.Offset
					btcUtxo.Value = utxo.Value
					btcUtxos = append(btcUtxos, *btcUtxo)
					fmt.Printf("address:%s,utxo:%s,offset:%d,value:%d, %s\n",
						addr, utxo.Id, sat.Offset, utxo.Value, sats_type)
				}
			}
		}
		//TODO the max utxo not has type
		if len(utxo.Sats) == 0 {
			if G_MaxBtcUtxo.Index == -1 {
				utxo_temp := strings.Split(utxo.Id, ":")
				G_MaxBtcUtxo.Id = utxo_temp[0]
				index, _ := strconv.Atoi(utxo_temp[1])
				G_MaxBtcUtxo.Index = index
				G_MaxBtcUtxo.Value = utxo.Value
			} else {
				if G_MaxBtcUtxo.Value < utxo.Value {
					utxo_temp := strings.Split(utxo.Id, ":")
					G_MaxBtcUtxo.Id = utxo_temp[0]
					index, _ := strconv.Atoi(utxo_temp[1])
					G_MaxBtcUtxo.Index = index
					G_MaxBtcUtxo.Value = utxo.Value
				}
			}
		}
	}
	return btcUtxos

}
