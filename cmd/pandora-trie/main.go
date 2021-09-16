package main

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/prysmaticlabs/prysm/beacon-chain/cache/depositcache"
	"github.com/prysmaticlabs/prysm/beacon-chain/db"
	"github.com/prysmaticlabs/prysm/beacon-chain/db/kv"
	"github.com/prysmaticlabs/prysm/beacon-chain/powchain"
	"github.com/prysmaticlabs/prysm/shared/params"
	"time"
)

func main() {
	boltDB, err := db.NewDB(context.Background(), "./", &kv.Config{
		InitialMMapSize: 536870912,
	})

	if nil != err {
		panic(err)
	}

	depositCache, err := depositcache.New()

	if nil != err {
		panic(err)
	}

	config := params.BeaconConfig()
	oldConfig := params.BeaconConfig().Copy()
	defer func() {
		params.OverrideBeaconConfig(oldConfig)
	}()
	config.DepositChainID = 808081
	config.DepositNetworkID = 808081
	params.OverrideBeaconConfig(config)

	remoteEndpoint := "http://34.141.25.249:8545"
	s1, err := powchain.NewService(context.Background(), &powchain.Web3ServiceConfig{
		HTTPEndpoints:   []string{remoteEndpoint},
		DepositContract: common.HexToAddress("0x000000000000000000000000000000000000cafe"),
		BeaconDB:        boltDB,
		DepositCache:    depositCache,
	})

	if nil != err {
		panic(err)
	}

	s1.Start()
	fmt.Printf("I am waiting 5 secs to estabilish connection to Pandora remote \n")
	time.Sleep(time.Second * 5)

	err, _ = s1.RecreateDepositTrieFromRemote()

	if nil != err {
		panic(fmt.Sprintf("got error during recretaion of remote: %s", err.Error()))
	}
}
