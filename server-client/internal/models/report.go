package models

type ItemSummary struct {
	Name     string
	Quantity int
}

type ShippingReport struct {
	Date                      string
	TotalOrders               int
	TotalCost                 float64
	Items                     map[string]ItemSummary
	OrdersByDestination       map[string]int
	DuplicateOrders           int
	OrdersByCustomer          map[string]int
	DuplicateOrdersByCustomer map[string]int
}

func NewShippingReport() ShippingReport {
	return ShippingReport{
		Items:                     make(map[string]ItemSummary),
		OrdersByDestination:       make(map[string]int),
		OrdersByCustomer:          make(map[string]int),
		DuplicateOrdersByCustomer: make(map[string]int),
	}
}
