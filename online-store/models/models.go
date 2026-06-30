package models

type Item struct {
	ItemID    string  `json:"itemId"`
	Name      string  `json:"name"`
	UnitPrice float64 `json:"unitPrice"`
}

type OrderItem struct {
	ItemID   string `json:"itemId"`
	Quantity int    `json:"quantity"`
}

type Order struct {
	OrderID     string      `json:"orderId"`
	CustomerID  string      `json:"customerId"`
	Items       []OrderItem `json:"items"`
	Destination string      `json:"destination"`
	Date        string      `json:"date"`
}
