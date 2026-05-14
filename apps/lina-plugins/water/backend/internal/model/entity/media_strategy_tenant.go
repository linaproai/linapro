// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// MediaStrategyTenant is the golang structure for table media_strategy_tenant.
type MediaStrategyTenant struct {
	TenantId   string `json:"tenantId"   orm:"tenant_id"   description:"租户ID"`
	StrategyId int64  `json:"strategyId" orm:"strategy_id" description:"策略ID"`
}
