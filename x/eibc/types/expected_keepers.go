package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	delayedacktypes "github.com/dymensionxyz/dymension/x/delayedack/types"
)

// AccountKeeper defines the expected account keeper used for simulations (noalias)
type AccountKeeper interface {
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) types.AccountI
}

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
}

type DelayedAckKeeper interface {
	GetRollappPacket(ctx sdk.Context, rollappPacketKey string) *delayedacktypes.RollappPacket
	UpdateRollappPacketRecipient(
		ctx sdk.Context,
		rollappPacketKey string,
		newRecipient string,
	) error
}
