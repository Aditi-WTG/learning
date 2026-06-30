package models

type OrderProcessor struct {
	ProcessedOrders map[string]bool
	UniqueOrders    []Order
	DuplicateOrders []Order
}

func NewOrderProcessor() *OrderProcessor {
	return &OrderProcessor{
		ProcessedOrders: make(map[string]bool),
		UniqueOrders:    []Order{},
		DuplicateOrders: []Order{},
	}
}

func (op *OrderProcessor) Process(order Order) {
	_, exists := op.ProcessedOrders[order.OrderID]
	if !exists {
		op.ProcessedOrders[order.OrderID] = true
		op.UniqueOrders = append(op.UniqueOrders, order)
	} else {
		op.DuplicateOrders = append(op.DuplicateOrders, order)
	}
}
