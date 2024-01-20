package ibctesting_test

import (
	"encoding/json"
	"strconv"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	transfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	ibctesting "github.com/cosmos/ibc-go/v6/testing"
	eibckeeper "github.com/dymensionxyz/dymension/x/eibc/keeper"
	eibctypes "github.com/dymensionxyz/dymension/x/eibc/types"
)

type EIBCTestSuite struct {
	KeeperTestSuite

	msgServer   eibctypes.MsgServer
	ctx         sdk.Context
	queryClient eibctypes.QueryClient
}

func TestEIBCTestSuite(t *testing.T) {
	suite.Run(t, new(EIBCTestSuite))
}

func (suite *EIBCTestSuite) SetupTest() {
	suite.KeeperTestSuite.SetupTest()
	eibcKeeper := ConvertToApp(suite.hubChain).EIBCKeeper
	suite.msgServer = eibckeeper.NewMsgServerImpl(eibcKeeper)
}

func (suite *EIBCTestSuite) TestEIBCDemandOrderCreation() {
	// Create rollapp only once
	suite.CreateRollapp()
	// Create cases
	cases := []struct {
		name                string
		amount              string
		fee                 string
		recipient           string
		demandOrdersCreated int
		isAckError          bool
	}{
		{
			"valid demand order",
			"1000000000",
			"150",
			suite.hubChain.SenderAccount.GetAddress().String(),
			1,
			false,
		},
		{
			"invalid demand order - negative fee",
			"1000000000",
			"-150",
			suite.hubChain.SenderAccount.GetAddress().String(),
			0,
			true,
		},
		{
			"invalid demand order - fee > amount",
			"1000",
			"1001",
			suite.hubChain.SenderAccount.GetAddress().String(),
			0,
			true,
		},
		{
			"invalid demand order - fee is 0",
			"1",
			"0",
			suite.hubChain.SenderAccount.GetAddress().String(),
			0,
			true,
		},
		{
			"invalid demand order - fee > max uint64",
			"10000",
			"100000000000000000000000000000",
			suite.hubChain.SenderAccount.GetAddress().String(),
			0,
			true,
		},
	}
	totalDemandOrdersCreated := 0
	for _, tc := range cases {
		suite.Run(tc.name, func() {

			// Send the EIBC Packet
			eibc := map[string]map[string]string{
				"eibc": {
					"fee": tc.fee,
				},
			}
			eibcJson, _ := json.Marshal(eibc)
			memo := string(eibcJson)
			suite.TransferRollappToHub(tc.recipient, tc.amount, memo, tc.isAckError)

			// Validate demand orders results
			eibcKeeper := ConvertToApp(suite.hubChain).EIBCKeeper
			demandOrders := eibcKeeper.GetAllDemandOrders(suite.hubChain.GetContext())
			suite.Require().Equal(tc.demandOrdersCreated, len(demandOrders)-totalDemandOrdersCreated)
			totalDemandOrdersCreated = len(demandOrders)
			amountInt, ok := sdk.NewIntFromString(tc.amount)
			suite.Require().True(ok)
			feeInt, ok := sdk.NewIntFromString(tc.fee)
			suite.Require().True(ok)
			if tc.demandOrdersCreated > 0 {
				lastDemandOrder := demandOrders[len(demandOrders)-1]
				suite.Require().Equal(tc.recipient, lastDemandOrder.Recipient)
				suite.Require().Equal(amountInt.Sub(feeInt), lastDemandOrder.Price)
				suite.Require().Equal(tc.fee, lastDemandOrder.Fee)
			}

		})
	}
}

func (suite *EIBCTestSuite) TestEIBCDemandOrderFulfillment() {
	// Create rollapp only once
	suite.CreateRollapp()
	// Create demand order to fulfill
	ibcTransferAmount := "200"
	eibcTransferFee := "150"
	eibcKeeper := ConvertToApp(suite.hubChain).EIBCKeeper
	delayedAckKeeper := ConvertToApp(suite.hubChain).DelayedAckKeeper
	// Create cases
	cases := []struct {
		name           string
		isOrderIdExist bool
		isReFulfill    bool
	}{
		{
			"fulfill demand order successfully",
			true,
			false,
		},
	}
	totalDemandOrdersCreated := 0
	for _, tc := range cases {
		suite.Run(tc.name, func() {
			// Create a valid demand order
			recipient := suite.hubChain.SenderAccount.GetAddress()
			initialRecipientAccountBalance := eibcKeeper.BankKeeper.SpendableCoins(suite.hubChain.GetContext(), recipient)
			eibc := map[string]map[string]string{
				"eibc": {
					"fee": eibcTransferFee,
				},
			}
			eibcJson, _ := json.Marshal(eibc)
			memo := string(eibcJson)
			suite.TransferRollappToHub(recipient.String(), ibcTransferAmount, memo, false)

			// Validate demand order created
			demandOrders := eibcKeeper.GetAllDemandOrders(suite.hubChain.GetContext())
			suite.Require().Greater(len(demandOrders), totalDemandOrdersCreated)
			totalDemandOrdersCreated = len(demandOrders)
			lastDemandOrder := demandOrders[len(demandOrders)-1]

			// Fulfill the demand order
			fullfillerAccount := suite.hubChain.SenderAccounts[len(suite.hubChain.SenderAccounts)-1].SenderAccount
			initialFulfillerAccountBalance := eibcKeeper.BankKeeper.SpendableCoins(suite.hubChain.GetContext(), fullfillerAccount.GetAddress())
			msgFulfillDemandOrder := &eibctypes.MsgFulfillOrder{
				FullfillerAddress: fullfillerAccount.GetAddress().String(),
				OrderId:           lastDemandOrder.Id,
			}
			// Validate demand order fulfilled
			_, err := suite.msgServer.FullfillOrder(suite.hubChain.GetContext(), msgFulfillDemandOrder)
			suite.Require().NoError(err)
			// validate eibc packet recipient has been updated
			rollappPacket := delayedAckKeeper.GetRollappPacket(suite.hubChain.GetContext(), lastDemandOrder.TrackingPacketKey)
			var data transfertypes.FungibleTokenPacketData
			err = transfertypes.ModuleCdc.UnmarshalJSON(rollappPacket.Packet.GetData(), &data)
			suite.Require().NoError(err)
			suite.Require().Equal(msgFulfillDemandOrder.FullfillerAddress, data.Receiver)
			// validate balances of fullfiller and recipient
			fullfillerAccountBalance := eibcKeeper.BankKeeper.SpendableCoins(suite.hubChain.GetContext(), fullfillerAccount.GetAddress())
			recipientAccountBalance := eibcKeeper.BankKeeper.SpendableCoins(suite.hubChain.GetContext(), recipient)
			ibcTransferAmountInt, _ := strconv.ParseInt(ibcTransferAmount, 10, 64)
			eibcTransferFeeInt, _ := strconv.ParseInt(eibcTransferFee, 10, 64)
			demandOrderPriceInt := ibcTransferAmountInt - eibcTransferFeeInt
			suite.Require().True(fullfillerAccountBalance.IsEqual(initialFulfillerAccountBalance.Sub(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(demandOrderPriceInt)))))
			suite.Require().True(recipientAccountBalance.IsEqual(initialRecipientAccountBalance.Add(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(demandOrderPriceInt)))))

			// validate demand order fulfilled
			demandOrder := eibcKeeper.GetDemandOrder(suite.hubChain.GetContext(), lastDemandOrder.Id)
			suite.Require().True(demandOrder.Status == eibctypes.DemandOrder_FULFILLED)

			// finalize rollapp and check balance was updated with fee
			suite.FinalizeRollapp()
			fullfillerAccountBalanceAfterFinalization := eibcKeeper.BankKeeper.SpendableCoins(suite.hubChain.GetContext(), fullfillerAccount.GetAddress())
			suite.Require().True(fullfillerAccountBalanceAfterFinalization.IsEqual(initialFulfillerAccountBalance.Add(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(eibcTransferFeeInt)))))

		})
	}
}

/* -------------------------------------------------------------------------- */
/*                                    Utils                                   */
/* -------------------------------------------------------------------------- */

func (suite *EIBCTestSuite) TransferRollappToHub(receiver string, amount string, memo string, isAckError bool) {
	path := suite.NewTransferPath(suite.hubChain, suite.rollappChain)
	suite.coordinator.Setup(path)
	hubEndpoint := path.EndpointA
	rollappEndpoint := path.EndpointB

	hubIBCKeeper := suite.hubChain.App.GetIBCKeeper()

	timeoutHeight := clienttypes.NewHeight(100, 110)
	amountInt, ok := sdk.NewIntFromString(amount)
	suite.Require().True(ok)
	coinToSendToB := sdk.NewCoin(sdk.DefaultBondDenom, amountInt)

	msg := types.NewMsgTransfer(rollappEndpoint.ChannelConfig.PortID, rollappEndpoint.ChannelID,
		coinToSendToB, suite.rollappChain.SenderAccount.GetAddress().String(), receiver, timeoutHeight, 0, memo)
	res, err := suite.rollappChain.SendMsgs(msg)
	suite.Require().NoError(err) // message committed

	packet, err := ibctesting.ParsePacketFromEvents(res.GetEvents())
	suite.Require().NoError(err)

	err = path.RelayPacket(packet)

	// If ack error that an ack is retuned immediately hence found. The reason we get err in the relay packet is
	// beacuse no ack can be parsed from events
	if isAckError {
		suite.Require().NoError(err)
		found := hubIBCKeeper.ChannelKeeper.HasPacketAcknowledgement(hubEndpoint.Chain.GetContext(), packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence())
		suite.Require().True(found)
	} else {
		suite.Require().Error(err)
		found := hubIBCKeeper.ChannelKeeper.HasPacketAcknowledgement(hubEndpoint.Chain.GetContext(), packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence())
		suite.Require().False(found)
	}

}
