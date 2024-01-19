package keeper

import (
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/dymensionxyz/dymension/x/eibc/types"
)

type (
	Keeper struct {
		cdc        codec.BinaryCodec
		storeKey   storetypes.StoreKey
		memKey     storetypes.StoreKey
		paramstore paramtypes.Subspace
		types.DelayedAckKeeper
		types.AccountKeeper
		types.BankKeeper
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	ps paramtypes.Subspace,
	delayedAckKeeper types.DelayedAckKeeper,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,

) *Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		cdc:              cdc,
		storeKey:         storeKey,
		memKey:           memKey,
		paramstore:       ps,
		DelayedAckKeeper: delayedAckKeeper,
		AccountKeeper:    accountKeeper,
		BankKeeper:       bankKeeper,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) SetDemandOrder(ctx sdk.Context, order *types.DemandOrder) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetDemandOrderKey(order.Id), k.cdc.MustMarshal(order))

	// Emit events
	eventAttributes := []sdk.Attribute{
		sdk.NewAttribute(types.AttributeKeyPacketKey, order.Id),
		sdk.NewAttribute(types.AttributeKeyPrice, order.Price),
		sdk.NewAttribute(types.AttributeKeyFee, order.Fee),
		sdk.NewAttribute(types.AttributeKeyDenom, order.Denom),
		sdk.NewAttribute(types.AttributeIsFulfilled, strconv.FormatBool(order.IsFulfilled)),
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeEIBC,
			eventAttributes...,
		),
	)
}

func (k Keeper) GetDemandOrder(ctx sdk.Context, id string) *types.DemandOrder {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetDemandOrderKey(id))
	if bz == nil {
		return nil
	}
	var order types.DemandOrder
	k.cdc.MustUnmarshal(bz, &order)
	return &order
}

// GetAllDemandOrders returns all demand orders. Shouldn't be exposed to the client.
func (k Keeper) GetAllDemandOrders(
	ctx sdk.Context,
) (list []types.DemandOrder) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.DemandOrderKeyPrefix))

	// Build the prefix which is composed of the rollappID and the status
	var prefix []byte

	// Iterate over the range from lastProofHeight to proofHeight
	iterator := sdk.KVStorePrefixIterator(store, prefix)
	defer iterator.Close() // nolint: errcheck

	for ; iterator.Valid(); iterator.Next() {
		var val types.DemandOrder
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return list
}
