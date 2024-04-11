package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	getSpecialSats "wallet/specialSats/src"
	satsTransfer "wallet/specialSats/src"

	"gopkg.in/ini.v1"
)

func main() {
	args := os.Args
	arg_num := len(os.Args)

	var config string
	if arg_num > 1 {
		config = args[1]
		fmt.Println(config)
	} else {
		config = "config.ini"
	}
	err := readText(config)

	if err != nil {
		return
	}

	types := getSpecialSats.G_Config.Section("").Key("types").Value()
	if types == "" {
		getSpecialSats.G_SatsTypes = []string{"uncommon", "rare", "epic", "legendary"}
	} else {
		getSpecialSats.G_SatsTypes = strings.Split(types, ",")
	}
	satsTransfer.G_MinimumSplitUtxo = 12000
	minimum_split_utxo := getSpecialSats.G_Config.Section("").Key("minimum_split_utxo").Value()
	if minimum_split_utxo != "" {
		value, _ := strconv.Atoi(minimum_split_utxo)
		satsTransfer.G_MinimumSplitUtxo = value
	}
	addrs := getSpecialSats.G_Config.Section("").Key("build_trx_address").Value()
	addrs_list := strings.Split(addrs, ",")

	to := getSpecialSats.G_Config.Section("").Key("to_uncommon_address").Value()
	satsTransfer.G_MaxTrxOutsNum = 200
	for _, addr := range addrs_list {
		buildBtxo := getSpecialSats.ScanAddrForSpecial(addr)
		/*
			addr = "tb1q32qy8l502dv54l9ekyvxjxw5nkcd8th0vg4ugl"
			buildBtxo := []satsTransfer.BTCUTXO{
				{
					Id:     "326a778ebde5c6926f0507bdc9ebc84962588efc094c734f7f58e5d7b633fb0c",
					Index:  2,
					Value:  11971,
					Offset: 3000,
				},
			}

			satsTransfer.G_MaxBtcUtxo = satsTransfer.BTCUTXO{
				Id:     "3aed61370646df7d92722cdb182039f0d6f38700697e84bfd7b956047e2a87f8",
				Index:  0,
				Value:  100000,
				Offset: 0,
			}
		*/
		tx := satsTransfer.BuildBtcTrx(addr, to, buildBtxo)
		fmt.Println(tx)
	}

}

func readText(cfg string) error {
	var err error
	getSpecialSats.G_Config, err = ini.Load(cfg)
	if err != nil {
		fmt.Println("err:", err)
		return err
	}
	return nil
}
