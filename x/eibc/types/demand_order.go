package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
)

func NewDemandOrder(id string, price string, fee string, denom string, recipient string) *DemandOrder {
	return &DemandOrder{
		Id:          id,
		Price:       price,
		Fee:         fee,
		Denom:       denom,
		Recipient:   recipient,
		IsFulfilled: false,
	}
}

func (m *DemandOrder) ValidateBasic() error {
	price, ok := sdk.NewIntFromString(m.Price)
	if !ok {
		return ErrInvalidAmount
	}
	if !price.IsPositive() {
		return ErrInvalidDemandOrderFee
	}
	fee, ok := sdk.NewIntFromString(m.Fee)
	if !ok {
		return ErrInvalidAmount
	}
	if !fee.IsPositive() {
		return ErrInvalidDemandOrderFee
	}
	if fee.GT(price) {
		return ErrInvalidDemandOrderFee
	}
	_, err := sdk.AccAddressFromBech32(m.Recipient)
	if err != nil {
		return ErrInvalidRecipientAddress
	}
	return ibctransfertypes.ValidatePrefixedDenom(m.Denom)
}

func (m *DemandOrder) Validate(totalAmount string) error {
	if err := m.ValidateBasic(); err != nil {
		return err
	}
	return nil
}

// GetPriceMathInt returns the price as a math.Int. Should
// be called after ValidateBasic hence should not panic.
func (m *DemandOrder) GetPriceInCoins() sdk.Coin {
	price, ok := sdk.NewIntFromString(m.Price)
	if !ok {
		panic("invalid price")
	}
	return sdk.NewCoin(m.Denom, price)
}

// GetRecipientBech32Address returns the recipient address as a string.
// Should be called after ValidateBasic hence should not panic.
func (m *DemandOrder) GetRecipientBech32Address() sdk.AccAddress {
	recipientBech32, err := sdk.AccAddressFromBech32(m.Recipient)
	if err != nil {
		panic(ErrInvalidRecipientAddress)
	}
	return recipientBech32
}
