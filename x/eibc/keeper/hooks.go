package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	delayeacktypes "github.com/dymensionxyz/dymension/x/delayedack/types"
	types "github.com/dymensionxyz/dymension/x/eibc/types"
)

var _ delayeacktypes.DelayedAckHooks = delayedAckHooks{}

type delayedAckHooks struct {
	delayeacktypes.BaseDelayedAckHook
	Keeper
}

func (k Keeper) GetDelayedAckHooks() delayeacktypes.DelayedAckHooks {
	return delayedAckHooks{
		BaseDelayedAckHook: delayeacktypes.BaseDelayedAckHook{},
		Keeper:             k,
	}
}

// AfterPacketStatusUpdated is called every time the underlying IBC packet is updated.
// There are 2 assumptions here:
// 1. The packet status can change only once hence the oldPacketKey should always represent the order ID as it was created from it.
// 2. The packet status can only change from PENDING
func (k delayedAckHooks) AfterPacketStatusUpdated(ctx sdk.Context, packet *delayeacktypes.RollappPacket,
	oldPacketKey string, newPacketKey string) error {
	// Update the demand order tracking packet key
	demandOrderID := types.BuildDemandIDFromPacketKey(oldPacketKey)
	demandOrder := k.GetDemandOrder(ctx, demandOrderID)
	demandOrder.TrackingPacketKey = newPacketKey
	// Update the demand order status according to the underlying packet status
	// If the demand order is already fulfilled updating it's status is not necessary
	if demandOrder.Status != types.DemandOrder_FULFILLED {
		switch packet.Status {
		case delayeacktypes.RollappPacket_ACCEPTED:
			demandOrder.Status = types.DemandOrder_EXPIRED
		case delayeacktypes.RollappPacket_REJECTED:
			demandOrder.Status = types.DemandOrder_REVERTED
		}
	}
	k.SetDemandOrder(ctx, demandOrder)
	return nil
}
