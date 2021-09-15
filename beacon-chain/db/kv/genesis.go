package kv

import (
	"bytes"
	"context"
	"fmt"
	ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
	"github.com/prysmaticlabs/prysm/beacon-chain/state/stateV0"
	protodb "github.com/prysmaticlabs/prysm/proto/beacon/db"
	"github.com/prysmaticlabs/prysm/shared/trieutil"
	"io"
	"io/ioutil"

	"github.com/pkg/errors"
	"github.com/prysmaticlabs/prysm/beacon-chain/core/blocks"
	dbIface "github.com/prysmaticlabs/prysm/beacon-chain/db/iface"
	iface "github.com/prysmaticlabs/prysm/beacon-chain/state/interface"
	state "github.com/prysmaticlabs/prysm/beacon-chain/state/stateV0"
	pbp2p "github.com/prysmaticlabs/prysm/proto/beacon/p2p/v1"
	"github.com/prysmaticlabs/prysm/shared/params"
)

// SaveGenesisData bootstraps the beaconDB with a given genesis state.
func (s *Store) SaveGenesisData(ctx context.Context, genesisState iface.BeaconState) error {
	stateRoot, err := genesisState.HashTreeRoot(ctx)
	if err != nil {
		return err
	}
	genesisBlk := blocks.NewGenesisBlock(stateRoot[:])
	genesisBlkRoot, err := genesisBlk.Block.HashTreeRoot()
	if err != nil {
		return errors.Wrap(err, "could not get genesis block root")
	}
	if err := s.SaveBlock(ctx, genesisBlk); err != nil {
		return errors.Wrap(err, "could not save genesis block")
	}
	if err := s.SaveState(ctx, genesisState, genesisBlkRoot); err != nil {
		return errors.Wrap(err, "could not save genesis state")
	}
	if err := s.SaveStateSummary(ctx, &pbp2p.StateSummary{
		Slot: 0,
		Root: genesisBlkRoot[:],
	}); err != nil {
		return err
	}

	if err := s.SaveHeadBlockRoot(ctx, genesisBlkRoot); err != nil {
		return errors.Wrap(err, "could not save head block root")
	}
	if err := s.SaveGenesisBlockRoot(ctx, genesisBlkRoot); err != nil {
		return errors.Wrap(err, "could not save genesis block root")
	}

	return nil
}

// LoadGenesisFromFile loads a genesis state from a given file path, if no genesis exists already.
func (s *Store) LoadGenesis(ctx context.Context, r io.Reader) error {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	st := &pbp2p.BeaconState{}
	if err := st.UnmarshalSSZ(b); err != nil {
		return err
	}
	gs, err := state.InitializeFromProtoUnsafe(st)
	if err != nil {
		return err
	}
	existing, err := s.GenesisState(ctx)
	if err != nil {
		return err
	}
	// If some different genesis state existed already, return an error. The same genesis state is
	// considered a no-op.
	if existing != nil {
		a, err := existing.HashTreeRoot(ctx)
		if err != nil {
			return err
		}
		b, err := gs.HashTreeRoot(ctx)
		if err != nil {
			return err
		}
		if a == b {
			return nil
		}
		return dbIface.ErrExistingGenesisState
	}

	if !bytes.Equal(gs.Fork().CurrentVersion, params.BeaconConfig().GenesisForkVersion) {
		return fmt.Errorf("loaded genesis fork version (%#x) does not match config genesis "+
			"fork version (%#x)", gs.Fork().CurrentVersion, params.BeaconConfig().GenesisForkVersion)
	}
	//
	powChainData, err := s.PowchainData(ctx)

	if nil != err {
		return err
	}

	depositTrie, err := trieutil.NewTrie(params.BeaconConfig().DepositContractTreeDepth)

	if nil != err {
		return err
	}

	// TODO:
	// Recreate deposit trie
	for index := uint64(0); index < gs.Eth1DepositIndex(); index++ {
	}

	// Recreate deposit containers
	// Insert them into database as eth1data

	if nil == powChainData {
		powChainData = &protodb.ETH1ChainData{
			CurrentEth1Data: &protodb.LatestETH1Data{
				BlockHeight: 0,
				BlockTime:   gs.GenesisTime(),
				// This is missing, can we recreate it somehow or pass it via flag?
				BlockHash:          st.Eth1Data.BlockHash,
				LastRequestedBlock: 0,
			},
			ChainstartData: &protodb.ChainStartData{
				Chainstarted: false,
				GenesisTime:  gs.GenesisTime(),
				GenesisBlock: 0,
				Eth1Data:     gs.Eth1Data(),
				// TODO: how to achieve chain start deposits?
				ChainstartDeposits: nil,
			},
			BeaconState:       st,
			Trie:              depositTrie.ToProto(),
			DepositContainers: s.cfg.DepositCache.AllDepositContainers(ctx),
		}
	}
	//
	//
	//
	//
	//err = s.SavePowchainData(ctx, eth1DataFromGs)
	//
	//if nil != err {
	//	return err
	//}

	return s.SaveGenesisData(ctx, gs)
}

// EnsureEmbeddedGenesis checks that a genesis block has been generated when an embedded genesis
// state is used. If a genesis block does not exist, but a genesis state does, then we should call
// SaveGenesisData on the existing genesis state.
func (s *Store) EnsureEmbeddedGenesis(ctx context.Context) error {
	gb, err := s.GenesisBlock(ctx)
	if err != nil {
		return err
	}
	if gb != nil {
		return nil
	}
	gs, err := s.GenesisState(ctx)
	if err != nil {
		return err
	}
	if gs != nil {
		return s.SaveGenesisData(ctx, gs)
	}
	return nil
}
