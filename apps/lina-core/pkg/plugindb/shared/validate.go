// This file validates typed plugindb query plans and mutation payloads.

package shared

import (
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
)

// ValidateDataQueryPlan validates one structured data query plan.
func ValidateDataQueryPlan(plan *DataQueryPlan) error {
	if plan == nil {
		return gerror.New("plugindb query plan cannot be nil")
	}
	if strings.TrimSpace(plan.Table) == "" {
		return gerror.New("plugindb query plan table cannot be empty")
	}
	if !plan.Action.IsValid() {
		return gerror.Newf("plugindb query plan action is invalid: %s", plan.Action)
	}
	for _, field := range plan.Fields {
		if strings.TrimSpace(field) == "" {
			return gerror.New("plugindb selected field cannot be empty")
		}
	}
	for _, filter := range plan.Filters {
		if err := ValidateDataFilter(filter); err != nil {
			return err
		}
	}
	for _, order := range plan.Orders {
		if err := ValidateDataOrder(order); err != nil {
			return err
		}
	}
	if plan.Action == DataPlanActionTransaction && plan.Transaction == nil {
		return gerror.New("plugindb transaction action requires transaction payload")
	}
	if plan.Action != DataPlanActionTransaction && plan.Transaction != nil {
		return gerror.Newf("plugindb action %s does not accept transaction payload", plan.Action)
	}
	if plan.Transaction != nil {
		if err := ValidateDataTransactionPlan(plan.Transaction); err != nil {
			return err
		}
	}
	return nil
}

// ValidateDataFilter validates one filter clause.
func ValidateDataFilter(filter *DataFilter) error {
	if filter == nil {
		return gerror.New("plugindb filter cannot be nil")
	}
	if strings.TrimSpace(filter.Field) == "" {
		return gerror.New("plugindb filter field cannot be empty")
	}
	if !filter.Operator.IsValid() {
		return gerror.Newf("plugindb filter operator is invalid: %s", filter.Operator)
	}
	switch filter.Operator {
	case DataFilterOperatorEQ, DataFilterOperatorLike:
		if len(filter.ValueJSON) == 0 {
			return gerror.Newf("plugindb filter %s requires valueJson", filter.Operator)
		}
	case DataFilterOperatorIN:
		if len(filter.ValuesJSON) == 0 {
			return gerror.New("plugindb filter in requires valuesJson")
		}
	}
	return nil
}

// ValidateDataOrder validates one order-by clause.
func ValidateDataOrder(order *DataOrder) error {
	if order == nil {
		return gerror.New("plugindb order cannot be nil")
	}
	if strings.TrimSpace(order.Field) == "" {
		return gerror.New("plugindb order field cannot be empty")
	}
	if !order.Direction.IsValid() {
		return gerror.Newf("plugindb order direction is invalid: %s", order.Direction)
	}
	return nil
}

// ValidateDataTransactionPlan validates one transaction payload.
func ValidateDataTransactionPlan(plan *DataTransactionPlan) error {
	if plan == nil {
		return gerror.New("plugindb transaction plan cannot be nil")
	}
	if len(plan.Operations) == 0 {
		return gerror.New("plugindb transaction plan requires at least one operation")
	}
	for _, operation := range plan.Operations {
		if operation == nil {
			return gerror.New("plugindb transaction operation cannot be nil")
		}
		if !operation.Action.IsValid() {
			return gerror.Newf("plugindb transaction action is invalid: %s", operation.Action)
		}
		if (operation.Action == DataMutationActionUpdate || operation.Action == DataMutationActionDelete) && len(operation.KeyJSON) == 0 {
			return gerror.Newf("plugindb transaction %s requires keyJson", operation.Action)
		}
		if (operation.Action == DataMutationActionCreate || operation.Action == DataMutationActionUpdate) && len(operation.RecordJSON) == 0 {
			return gerror.Newf("plugindb transaction %s requires recordJson", operation.Action)
		}
	}
	return nil
}
