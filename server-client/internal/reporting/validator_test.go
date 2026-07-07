package reporting

import (
	"server-client/internal/models"
	"testing"
)

func TestValidate(t *testing.T) {
	catalog := map[string]models.Item{
		"I001": {ItemID: "I001", Name: "apple", UnitPrice: 10},
	}

	valid := models.Order{
		OrderID:     "O1",
		CustomerID:  "C1",
		Items:       []models.OrderItem{{ItemID: "I001", Quantity: 1}},
		Destination: "Bangalore",
		Date:        "2026-07-07",
	}

	tests := []struct {
		name    string
		order   models.Order
		wantErr bool
	}{
		{name: "valid", order: valid, wantErr: false},
		{name: "missing order id", order: models.Order{CustomerID: "C1", Items: valid.Items, Destination: "Bangalore", Date: "2026-07-07"}, wantErr: true},
		{name: "missing customer id", order: models.Order{OrderID: "O1", Items: valid.Items, Destination: "Bangalore", Date: "2026-07-07"}, wantErr: true},
		{name: "missing items", order: models.Order{OrderID: "O1", CustomerID: "C1", Destination: "Bangalore", Date: "2026-07-07"}, wantErr: true},
		{name: "invalid quantity", order: models.Order{OrderID: "O1", CustomerID: "C1", Items: []models.OrderItem{{ItemID: "I001", Quantity: 0}}, Destination: "Bangalore", Date: "2026-07-07"}, wantErr: true},
		{name: "invalid item id", order: models.Order{OrderID: "O1", CustomerID: "C1", Items: []models.OrderItem{{ItemID: "I999", Quantity: 1}}, Destination: "Bangalore", Date: "2026-07-07"}, wantErr: true},
		{name: "missing destination", order: models.Order{OrderID: "O1", CustomerID: "C1", Items: valid.Items, Date: "2026-07-07"}, wantErr: true},
		{name: "missing date", order: models.Order{OrderID: "O1", CustomerID: "C1", Items: valid.Items, Destination: "Bangalore"}, wantErr: true},
		{name: "invalid date", order: models.Order{OrderID: "O1", CustomerID: "C1", Items: valid.Items, Destination: "Bangalore", Date: "07-07-2026"}, wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := Validate(tc.order, catalog)
			if tc.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}
