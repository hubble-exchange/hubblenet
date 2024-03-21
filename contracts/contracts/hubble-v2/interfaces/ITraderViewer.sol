// SPDX-License-Identifier: MIT

pragma solidity ^0.8.0;

import { IClearingHouse } from "./IClearingHouse.sol";
import { ILimitOrderBook } from "./IJuror.sol";
import { IOrderHandler } from "./IOrderHandler.sol";

interface ITraderViewer {
    function getNotionalPositionAndMargin(address trader, bool includeFundingPayments, IClearingHouse.Mode mode) external view returns(uint256 notionalPosition, int256 margin, uint256 requiredMargin);

    function getTraderDataForMarket(address trader, uint256 ammIndex, IClearingHouse.Mode mode) external view returns(
        bool isIsolated, uint256 notionalPosition, int256 unrealizedPnl, uint256 requiredMargin, int256 pendingFunding
    );

    function getCrossMarginAccountData(address trader, IClearingHouse.Mode mode)
        external
        view
        returns(uint256 notionalPosition, uint256 requiredMargin, int256 unrealizedPnl, int256 pendingFunding);

    function getTotalFundingForCrossMarginPositions(address trader) external view returns(int256 totalFunding);

    function validateCancelLimitOrderV2(ILimitOrderBook.Order memory order, address sender, bool assertLowMargin, bool assertOverPositionCap) external view returns (string memory err, bytes32 orderHash, IOrderHandler.CancelOrderRes memory res);

    function getRequiredMargin(int256 baseAssetQuantity, uint256 price, uint ammIndex, address trader) external view returns(uint256 requiredMargin);
}
