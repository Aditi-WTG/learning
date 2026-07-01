package reporting

import "online-store/models"

func GenerateReport(uniqueOrders []models.Order, duplicateOrders []models.Order, catalogMap map[string]models.Item) models.ShippingReport {
	report := models.NewShippingReport()

	report.TotalOrders = len(uniqueOrders) + len(duplicateOrders)
	report.DuplicateOrders = len(duplicateOrders)

	for _, order := range uniqueOrders {
		for _, orderItem := range order.Items {
			item := catalogMap[orderItem.ItemID]

			report.TotalCost += float64(orderItem.Quantity) * item.UnitPrice

			itemSummary := report.Items[item.ItemID]
			itemSummary.Name = item.Name
			itemSummary.Quantity += orderItem.Quantity
			report.Items[item.ItemID] = itemSummary
		}

		report.OrdersByDestination[order.Destination]++
		report.OrdersByCustomer[order.CustomerID]++
	}

	for _, duplicateOrder := range duplicateOrders {
		report.DuplicateOrdersByCustomer[duplicateOrder.CustomerID]++
	}

	return report
}
