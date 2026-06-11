// This file validates typed record store capability query plans and mutation payloads.

package plan

import (
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
)

// ValidateQueryPlan validates one structured record store query plan.
func ValidateQueryPlan(plan *QueryPlan) error {
	if plan == nil {
		return gerror.New("record store capability query plan cannot be nil")
	}
	if strings.TrimSpace(plan.Table) == "" {
		return gerror.New("record store capability query plan table cannot be empty")
	}
	if !plan.Action.IsValid() {
		return gerror.Newf("record store capability query plan action is invalid: %s", plan.Action)
	}
	for _, field := range plan.Fields {
		if strings.TrimSpace(field) == "" {
			return gerror.New("record store capability selected field cannot be empty")
		}
	}
	for _, filter := range plan.Filters {
		if err := ValidateFilter(filter); err != nil {
			return err
		}
	}
	for _, order := range plan.Orders {
		if err := ValidateOrder(order); err != nil {
			return err
		}
	}
	if plan.Action == PlanActionTransaction && plan.Transaction == nil {
		return gerror.New("record store capability transaction action requires transaction payload")
	}
	if plan.Action != PlanActionTransaction && plan.Transaction != nil {
		return gerror.Newf("record store capability action %s does not accept transaction payload", plan.Action)
	}
	if plan.Transaction != nil {
		if err := ValidateTransactionPlan(plan.Transaction); err != nil {
			return err
		}
	}
	return nil
}

// ValidateFilter validates one filter clause.
func ValidateFilter(filter *Filter) error {
	if filter == nil {
		return gerror.New("record store capability filter cannot be nil")
	}
	if strings.TrimSpace(filter.Field) == "" {
		return gerror.New("record store capability filter field cannot be empty")
	}
	if !filter.Operator.IsValid() {
		return gerror.Newf("record store capability filter operator is invalid: %s", filter.Operator)
	}
	switch filter.Operator {
	case FilterOperatorEQ, FilterOperatorLike:
		if len(filter.ValueJSON) == 0 {
			return gerror.Newf("record store capability filter %s requires valueJson", filter.Operator)
		}
	case FilterOperatorIN:
		if len(filter.ValuesJSON) == 0 {
			return gerror.New("record store capability filter in requires valuesJson")
		}
	}
	return nil
}

// ValidateOrder validates one order-by clause.
func ValidateOrder(order *Order) error {
	if order == nil {
		return gerror.New("record store capability order cannot be nil")
	}
	if strings.TrimSpace(order.Field) == "" {
		return gerror.New("record store capability order field cannot be empty")
	}
	if !order.Direction.IsValid() {
		return gerror.Newf("record store capability order direction is invalid: %s", order.Direction)
	}
	return nil
}

// ValidateTransactionPlan validates one transaction payload.
func ValidateTransactionPlan(plan *TransactionPlan) error {
	if plan == nil {
		return gerror.New("record store capability transaction plan cannot be nil")
	}
	if len(plan.Operations) == 0 {
		return gerror.New("record store capability transaction plan requires at least one operation")
	}
	for _, operation := range plan.Operations {
		if operation == nil {
			return gerror.New("record store capability transaction operation cannot be nil")
		}
		if !operation.Action.IsValid() {
			return gerror.Newf("record store capability transaction action is invalid: %s", operation.Action)
		}
		if (operation.Action == MutationActionUpdate || operation.Action == MutationActionDelete) && len(operation.KeyJSON) == 0 {
			return gerror.Newf("record store capability transaction %s requires keyJson", operation.Action)
		}
		if (operation.Action == MutationActionCreate || operation.Action == MutationActionUpdate) && len(operation.RecordJSON) == 0 {
			return gerror.Newf("record store capability transaction %s requires recordJson", operation.Action)
		}
	}
	return nil
}
