pragma solidity 0.8.9;

import "../IAllowList.sol";

interface IHubbleConfigManager is IAllowList{
  //getSpreadRatioThreshold returns the spreadRatioThreshold stored in evm state
  function getSpreadRatioThreshold() external view returns (uint256 result);

  //setSpreadRatioThreshold stores the spreadRatioThreshold in evm state
  function setSpreadRatioThreshold(uint256 response) external;
}