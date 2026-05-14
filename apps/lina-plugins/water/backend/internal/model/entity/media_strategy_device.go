// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// MediaStrategyDevice is the golang structure for table media_strategy_device.
type MediaStrategyDevice struct {
	DeviceId   string `json:"deviceId"   orm:"device_id"   description:"设备国标ID"`
	StrategyId int64  `json:"strategyId" orm:"strategy_id" description:"策略ID"`
}
