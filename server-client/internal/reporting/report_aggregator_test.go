package reporting

import (
	"server-client/internal/models"
	"testing"
)

func testCatalog() []models.Item {
	return []models.Item{
		{ItemID: "I001", Name: "apple", UnitPrice: 10},
		{ItemID: "I002", Name: "banana", UnitPrice: 5},
	}
}

func testOrder(id, customerID, itemID, destination, date string, qty int) models.Order {
	return models.Order{
		OrderID:     id,
		CustomerID:  customerID,
		Items:       []models.OrderItem{{ItemID: itemID, Quantity: qty}},
		Destination: destination,
		Date:        date,
	}
}

func TestProcessOrderAndDuplicate(t *testing.T) {
	a := NewReportAggregator(testCatalog())

	order := testOrder("O1001", "C001", "I001", "Bangalore", "2026-07-07", 2)
	report, err := a.ProcessOrder(order)
	if err != nil {
		t.Fatalf("unexpected error for valid order: %v", err)
	}

	if report.TotalOrders != 1 {
		t.Fatalf("expected TotalOrders=1, got %d", report.TotalOrders)
	}
	if report.DuplicateOrders != 0 {
		t.Fatalf("expected DuplicateOrders=0, got %d", report.DuplicateOrders)
	}
	if report.TotalCost != 20 {
		t.Fatalf("expected TotalCost=20, got %v", report.TotalCost)
	}

	report, err = a.ProcessOrder(order)
	if err != nil {
		t.Fatalf("unexpected error for duplicate order: %v", err)
	}

	if report.TotalOrders != 1 {
		t.Fatalf("expected TotalOrders=1 after duplicate order, got %d", report.TotalOrders)
	}
	if report.DuplicateOrders != 1 {
		t.Fatalf("expected DuplicateOrders=1, got %d", report.DuplicateOrders)
	}
	if report.TotalCost != 20 {
		t.Fatalf("duplicate order should not change TotalCost, got %v", report.TotalCost)
	}
	if report.DuplicateOrdersByCustomer["C001"] != 1 {
		t.Fatalf("expected duplicate count for customer C001 to be 1")
	}
}

func TestSnapshotAllSortedByDate(t *testing.T) {
	a := NewReportAggregator(testCatalog())

	_, err := a.ProcessOrder(testOrder("O2001", "C001", "I001", "Bangalore", "2026-07-08", 1))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = a.ProcessOrder(testOrder("O2002", "C002", "I002", "Mumbai", "2026-07-05", 2))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	all := a.SnapshotAll()
	if len(all) != 2 {
		t.Fatalf("expected 2 reports, got %d", len(all))
	}

	if all[0].Date != "2026-07-05" || all[1].Date != "2026-07-08" {
		t.Fatalf("reports are not sorted by date: got %s then %s", all[0].Date, all[1].Date)
	}
}

func TestSnapshotReturnsClone(t *testing.T) {
	a := NewReportAggregator(testCatalog())

	_, err := a.ProcessOrder(testOrder("O3001", "C001", "I001", "Bangalore", "2026-07-07", 1))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	snap1, ok := a.Snapshot("2026-07-07")
	if !ok {
		t.Fatal("expected snapshot to exist")
	}

	snap1.Items["I001"] = models.ItemSummary{Name: "tampered", Quantity: 999}
	snap1.OrdersByCustomer["C001"] = 999

	snap2, ok := a.Snapshot("2026-07-07")
	if !ok {
		t.Fatal("expected snapshot to exist")
	}

	if snap2.Items["I001"].Name == "tampered" || snap2.Items["I001"].Quantity == 999 {
		t.Fatal("snapshot mutation leaked into internal state")
	}
	if snap2.OrdersByCustomer["C001"] == 999 {
		t.Fatal("snapshot map mutation leaked into internal state")
	}
}

func TestProcessOrderAggregatesMultipleItemsAndCounters(t *testing.T) {
	a := NewReportAggregator(testCatalog())

	order := models.Order{
		OrderID:    "O4001",
		CustomerID: "C010",
		Items: []models.OrderItem{
			{ItemID: "I001", Quantity: 2},
			{ItemID: "I002", Quantity: 3},
		},
		Destination: "Delhi",
		Date:        "2026-07-10",
	}

	report, err := a.ProcessOrder(order)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if report.TotalOrders != 1 {
		t.Fatalf("expected TotalOrders=1, got %d", report.TotalOrders)
	}

	if report.TotalCost != 35 {
		t.Fatalf("expected TotalCost=35, got %v", report.TotalCost)
	}

	if report.Items["I001"].Quantity != 2 || report.Items["I002"].Quantity != 3 {
		t.Fatal("unexpected item quantities in report")
	}

	if report.OrdersByDestination["Delhi"] != 1 {
		t.Fatal("expected destination counter for Delhi to be 1")
	}

	if report.OrdersByCustomer["C010"] != 1 {
		t.Fatal("expected customer counter for C010 to be 1")
	}
}
