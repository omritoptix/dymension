package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/eibc module sentinel errors
var (
	ErrInvalidDemandOrderPrice       = sdkerrors.Register(ModuleName, 1, "Price must be greater than 0")
	ErrInvalidDemandOrderFee         = sdkerrors.Register(ModuleName, 2, "Fee must be greater than 0 and less than or equal to the total amount")
	ErrInvalidOrderID                = sdkerrors.Register(ModuleName, 3, "Invalid order ID")
	ErrInvalidAmount                 = sdkerrors.Register(ModuleName, 4, "invalid amount")
	ErrDemandOrderDoesNotExist       = sdkerrors.Register(ModuleName, 5, "demand order does not exist")
	ErrDemandOrderAlreadyFullfilled  = sdkerrors.Register(ModuleName, 6, "demand order already fulfilled")
	ErrFullfillerAddressDoesNotExist = sdkerrors.Register(ModuleName, 7, "fullfiller address does not exist")
	ErrFullfillerInsufficientBalance = sdkerrors.Register(ModuleName, 8, "fullfiller does not have enough balance")
	ErrInvalidRecipientAddress       = sdkerrors.Register(ModuleName, 9, "invalid recipient address")
)
