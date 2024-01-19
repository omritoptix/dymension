package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	delayedacktypes "github.com/dymensionxyz/dymension/x/delayedack/types"
	"github.com/dymensionxyz/dymension/x/eibc/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

func (m msgServer) FullfillOrder(goCtx context.Context, msg *types.MsgFulfillOrder) (*types.MsgFulfillOrderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	logger := ctx.Logger()
	// Check that the msg is valid
	err := msg.ValidateBasic()
	if err != nil {
		return nil, err
	}
	// Check that the order exists
	demandOrder := m.GetDemandOrder(ctx, msg.OrderId)
	if demandOrder == nil {
		return nil, types.ErrDemandOrderDoesNotExist
	}
	// Check that the order is not yet fulfilled
	if demandOrder.IsFulfilled {
		return nil, types.ErrDemandOrderAlreadyFullfilled
	}
	// Check that the underlying packet is still in pending status
	rollappPacket := m.GetRollappPacket(ctx, msg.OrderId)
	if rollappPacket == nil {
		return nil, delayedacktypes.ErrRollappPacketDoesNotExist
	}
	if rollappPacket.Status != delayedacktypes.RollappPacket_PENDING {
		return nil, delayedacktypes.ErrCanOnlyUpdatePendingPacket
	}
	// Check that the fullfiller has enough balance to fulfill the order
	fullfillerAccount := m.GetAccount(ctx, msg.GetFullfillerBech32Address())
	if fullfillerAccount == nil {
		return nil, types.ErrFullfillerAddressDoesNotExist
	}
	fullfillerBalance := m.BankKeeper.SpendableCoins(ctx, fullfillerAccount.GetAddress())
	requiredBalance := demandOrder.GetPriceInCoins()
	// Iterate through the coins and check if the fullfiller has enough balance
	for _, coin := range fullfillerBalance {
		if coin.Denom == requiredBalance.Denom {
			if coin.Amount.LT(requiredBalance.Amount) {
				return nil, types.ErrFullfillerInsufficientBalance
			}
		}
	}
	// Update the receiver address to the market maker address
	err = m.UpdateRollappPacketRecipient(ctx, msg.OrderId, msg.FullfillerAddress)
	if err != nil {
		logger.Error("Failed to update rollapp packet recipient", "error", err)
		return nil, err
	}
	// Send the funds from the fullfiller to the eibc packet owner
	err = m.BankKeeper.SendCoins(ctx, fullfillerAccount.GetAddress(), demandOrder.GetRecipientBech32Address(), sdk.Coins{requiredBalance})
	if err != nil {
		logger.Error("Failed to send coins", "error", err)
		return nil, err
	}
	// Mark the order as fulfilled
	demandOrder.IsFulfilled = true
	m.SetDemandOrder(ctx, demandOrder)

	return &types.MsgFulfillOrderResponse{}, nil
}
