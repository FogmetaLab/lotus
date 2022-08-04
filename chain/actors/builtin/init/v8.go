package init

import (
	"crypto/sha256"
	"fmt"

	"github.com/ipfs/go-cid"
	cbg "github.com/whyrusleeping/cbor-gen"
	"golang.org/x/xerrors"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	builtin8 "github.com/filecoin-project/go-state-types/builtin"
	init8 "github.com/filecoin-project/go-state-types/builtin/v8/init"
	adt8 "github.com/filecoin-project/go-state-types/builtin/v8/util/adt"

	"github.com/filecoin-project/lotus/chain/actors"
	"github.com/filecoin-project/lotus/chain/actors/adt"
	"github.com/filecoin-project/lotus/node/modules/dtypes"
)

var _ State = (*state8)(nil)

func load8(store adt.Store, root cid.Cid) (State, error) {
	out := state8{store: store}
	err := store.Get(store.Context(), root, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func make8(store adt.Store, networkName string) (State, error) {
	out := state8{store: store}

	s, err := init8.ConstructState(store, networkName)
	if err != nil {
		return nil, err
	}

	out.State = *s

	return &out, nil
}

type state8 struct {
	init8.State
	store adt.Store
}

func (s *state8) ResolveAddress(address address.Address) (address.Address, bool, error) {
	return s.State.ResolveAddress(s.store, address)
}

func (s *state8) MapAddressToNewID(address address.Address) (address.Address, error) {
	return s.State.MapAddressToNewID(s.store, address)
}

func (s *state8) ForEachActor(cb func(id abi.ActorID, address address.Address) error) error {
	addrs, err := adt8.AsMap(s.store, s.State.AddressMap, builtin8.DefaultHamtBitwidth)
	if err != nil {
		return err
	}
	var actorID cbg.CborInt
	return addrs.ForEach(&actorID, func(key string) error {
		addr, err := address.NewFromBytes([]byte(key))
		if err != nil {
			return err
		}
		return cb(abi.ActorID(actorID), addr)
	})
}

func (s *state8) NetworkName() (dtypes.NetworkName, error) {
	return dtypes.NetworkName(s.State.NetworkName), nil
}

func (s *state8) SetNetworkName(name string) error {
	s.State.NetworkName = name
	return nil
}

func (s *state8) SetNextID(id abi.ActorID) error {
	s.State.NextID = id
	return nil
}

func (s *state8) Remove(addrs ...address.Address) (err error) {
	m, err := adt8.AsMap(s.store, s.State.AddressMap, builtin8.DefaultHamtBitwidth)
	if err != nil {
		return err
	}
	for _, addr := range addrs {
		if err = m.Delete(abi.AddrKey(addr)); err != nil {
			return xerrors.Errorf("failed to delete entry for address: %s; err: %w", addr, err)
		}
	}
	amr, err := m.Root()
	if err != nil {
		return xerrors.Errorf("failed to get address map root: %w", err)
	}
	s.State.AddressMap = amr
	return nil
}

func (s *state8) SetAddressMap(mcid cid.Cid) error {
	s.State.AddressMap = mcid
	return nil
}

func (s *state8) GetState() interface{} {
	return &s.State
}

func (s *state8) ActorKey() string {
	return actors.InitKey
}

func (s *state8) ActorVersion() actors.Version {
	return actors.Version8
}

func (s *state8) Code() cid.Cid {
	code, ok := actors.GetActorCodeID(s.ActorVersion(), s.ActorKey())
	if !ok {
		panic(fmt.Errorf("didn't find actor %v code id for actor version %d", s.ActorKey(), s.ActorVersion()))
	}

	return code
}

func (s *state8) AddressMap() (adt.Map, error) {
	return adt8.AsMap(s.store, s.State.AddressMap, builtin8.DefaultHamtBitwidth)
}

func (s *state8) AddressMapBitWidth() int {
	return builtin8.DefaultHamtBitwidth
}

func (s *state8) AddressMapHashFunction() func(input []byte) []byte {
	return func(input []byte) []byte {
		res := sha256.Sum256(input)
		return res[:]
	}
}
