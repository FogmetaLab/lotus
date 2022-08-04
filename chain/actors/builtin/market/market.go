package market

import (
	"unicode/utf8"

	"github.com/ipfs/go-cid"
	cbg "github.com/whyrusleeping/cbor-gen"
	"golang.org/x/xerrors"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	builtin8 "github.com/filecoin-project/go-state-types/builtin"
	market8 "github.com/filecoin-project/go-state-types/builtin/v8/market"
	"github.com/filecoin-project/go-state-types/cbor"
	"github.com/filecoin-project/go-state-types/network"
	builtin0 "github.com/filecoin-project/specs-actors/actors/builtin"
	builtin2 "github.com/filecoin-project/specs-actors/v2/actors/builtin"
	builtin3 "github.com/filecoin-project/specs-actors/v3/actors/builtin"
	builtin4 "github.com/filecoin-project/specs-actors/v4/actors/builtin"
	builtin5 "github.com/filecoin-project/specs-actors/v5/actors/builtin"
	builtin6 "github.com/filecoin-project/specs-actors/v6/actors/builtin"
	builtin7 "github.com/filecoin-project/specs-actors/v7/actors/builtin"

	"github.com/filecoin-project/lotus/chain/actors"
	"github.com/filecoin-project/lotus/chain/actors/adt"
	"github.com/filecoin-project/lotus/chain/types"
)

var (
	Address = builtin8.StorageMarketActorAddr
	Methods = builtin8.MethodsMarket
)

func Load(store adt.Store, act *types.Actor) (State, error) {
	if name, av, ok := actors.GetActorMetaByCode(act.Code); ok {
		if name != actors.MarketKey {
			return nil, xerrors.Errorf("actor code is not market: %s", name)
		}

		switch av {

		case actors.Version8:
			return load8(store, act.Head)

		}
	}

	switch act.Code {

	case builtin0.StorageMarketActorCodeID:
		return load0(store, act.Head)

	case builtin2.StorageMarketActorCodeID:
		return load2(store, act.Head)

	case builtin3.StorageMarketActorCodeID:
		return load3(store, act.Head)

	case builtin4.StorageMarketActorCodeID:
		return load4(store, act.Head)

	case builtin5.StorageMarketActorCodeID:
		return load5(store, act.Head)

	case builtin6.StorageMarketActorCodeID:
		return load6(store, act.Head)

	case builtin7.StorageMarketActorCodeID:
		return load7(store, act.Head)

	}

	return nil, xerrors.Errorf("unknown actor code %s", act.Code)
}

func MakeState(store adt.Store, av actors.Version) (State, error) {
	switch av {

	case actors.Version0:
		return make0(store)

	case actors.Version2:
		return make2(store)

	case actors.Version3:
		return make3(store)

	case actors.Version4:
		return make4(store)

	case actors.Version5:
		return make5(store)

	case actors.Version6:
		return make6(store)

	case actors.Version7:
		return make7(store)

	case actors.Version8:
		return make8(store)

	}
	return nil, xerrors.Errorf("unknown actor version %d", av)
}

type State interface {
	cbor.Marshaler
	BalancesChanged(State) (bool, error)
	EscrowTable() (BalanceTable, error)
	LockedTable() (BalanceTable, error)
	TotalLocked() (abi.TokenAmount, error)
	StatesChanged(State) (bool, error)
	VerifyDealsForActivation(
		minerAddr address.Address, deals []abi.DealID, currEpoch, sectorExpiry abi.ChainEpoch,
	) (weight, verifiedWeight abi.DealWeight, err error)
	NextID() (abi.DealID, error)
	GetState() interface{}

	Proposals() (DealProposals, error)
	ProposalsChanged(State) (bool, error)
	DealProposalsAmtBitWidth() int

	States() (DealStates, error)
	DealStatesAmtBitWidth() int
}

type BalanceTable interface {
	ForEach(cb func(address.Address, abi.TokenAmount) error) error
	Get(key address.Address) (abi.TokenAmount, error)
}

type DealStates interface {
	ForEach(cb func(id abi.DealID, ds DealState) error) error
	Get(id abi.DealID) (*DealState, bool, error)

	StatesArray() adt.Array
	Decode(*cbg.Deferred) (*DealState, error)
}

type DealProposals interface {
	ForEach(cb func(id abi.DealID, dp market8.DealProposal) error) error
	Get(id abi.DealID) (*market8.DealProposal, bool, error)

	ProposalsArray() adt.Array
	Decode(*cbg.Deferred) (*market8.DealProposal, error)
}

type PublishStorageDealsReturn interface {
	DealIDs() ([]abi.DealID, error)
	// Note that this index is based on the batch of deals that were published, NOT the DealID
	IsDealValid(index uint64) (bool, int, error)
}

func DecodePublishStorageDealsReturn(b []byte, nv network.Version) (PublishStorageDealsReturn, error) {
	av, err := actors.VersionForNetwork(nv)
	if err != nil {
		return nil, err
	}

	switch av {

	case actors.Version0:
		return decodePublishStorageDealsReturn0(b)

	case actors.Version2:
		return decodePublishStorageDealsReturn2(b)

	case actors.Version3:
		return decodePublishStorageDealsReturn3(b)

	case actors.Version4:
		return decodePublishStorageDealsReturn4(b)

	case actors.Version5:
		return decodePublishStorageDealsReturn5(b)

	case actors.Version6:
		return decodePublishStorageDealsReturn6(b)

	case actors.Version7:
		return decodePublishStorageDealsReturn7(b)

	case actors.Version8:
		return decodePublishStorageDealsReturn8(b)

	}
	return nil, xerrors.Errorf("unknown actor version %d", av)
}

type DealProposal = market8.DealProposal

type DealState = market8.DealState

type DealStateChanges struct {
	Added    []DealIDState
	Modified []DealStateChange
	Removed  []DealIDState
}

type DealIDState struct {
	ID   abi.DealID
	Deal DealState
}

// DealStateChange is a change in deal state from -> to
type DealStateChange struct {
	ID   abi.DealID
	From *DealState
	To   *DealState
}

type DealProposalChanges struct {
	Added   []ProposalIDState
	Removed []ProposalIDState
}

type ProposalIDState struct {
	ID       abi.DealID
	Proposal market8.DealProposal
}

func EmptyDealState() *DealState {
	return &DealState{
		SectorStartEpoch: -1,
		SlashEpoch:       -1,
		LastUpdatedEpoch: -1,
	}
}

// returns the earned fees and pending fees for a given deal
func GetDealFees(deal market8.DealProposal, height abi.ChainEpoch) (abi.TokenAmount, abi.TokenAmount) {
	tf := big.Mul(deal.StoragePricePerEpoch, big.NewInt(int64(deal.EndEpoch-deal.StartEpoch)))

	ef := big.Mul(deal.StoragePricePerEpoch, big.NewInt(int64(height-deal.StartEpoch)))
	if ef.LessThan(big.Zero()) {
		ef = big.Zero()
	}

	if ef.GreaterThan(tf) {
		ef = tf
	}

	return ef, big.Sub(tf, ef)
}

func IsDealActive(state market8.DealState) bool {
	return state.SectorStartEpoch > -1 && state.SlashEpoch == -1
}

func labelFromGoString(s string) (market8.DealLabel, error) {
	if utf8.ValidString(s) {
		return market8.NewLabelFromString(s)
	} else {
		return market8.NewLabelFromBytes([]byte(s))
	}
}

func AllCodes() []cid.Cid {
	return []cid.Cid{
		(&state0{}).Code(),
		(&state2{}).Code(),
		(&state3{}).Code(),
		(&state4{}).Code(),
		(&state5{}).Code(),
		(&state6{}).Code(),
		(&state7{}).Code(),
		(&state8{}).Code(),
	}
}

func VersionCodes() map[actors.Version]cid.Cid {
	return map[actors.Version]cid.Cid{
		actors.Version0: (&state0{}).Code(),
		actors.Version2: (&state2{}).Code(),
		actors.Version3: (&state3{}).Code(),
		actors.Version4: (&state4{}).Code(),
		actors.Version5: (&state5{}).Code(),
		actors.Version6: (&state6{}).Code(),
		actors.Version7: (&state7{}).Code(),
		actors.Version8: (&state8{}).Code(),
	}
}
