package reporting

import "online-store/models"

type Deduplicator struct {
	seenOrderIDs    map[string]bool
	UniqueOrders    []models.Order
	DuplicateOrders []models.Order
}

func NewDeduplicator() *Deduplicator {
	return &Deduplicator{
		seenOrderIDs:    make(map[string]bool),
		UniqueOrders:    []models.Order{},
		DuplicateOrders: []models.Order{},
	}
}

func (d *Deduplicator) Add(order models.Order) {
	_, exists := d.seenOrderIDs[order.OrderID]
	if !exists {
		d.seenOrderIDs[order.OrderID] = true
		d.UniqueOrders = append(d.UniqueOrders, order)
	} else {
		d.DuplicateOrders = append(d.DuplicateOrders, order)
	}
}
