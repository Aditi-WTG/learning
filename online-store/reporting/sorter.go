package reporting

import (
	"online-store/models"
	"sort"
)

func Sort(orders []models.Order) {
	sort.Slice(orders, func(i, j int) bool {
		if orders[i].Date != orders[j].Date {
			return orders[i].Date < orders[j].Date
		}

		if orders[i].CustomerID != orders[j].CustomerID {
			return orders[i].CustomerID < orders[j].CustomerID
		}

		return orders[i].OrderID < orders[j].OrderID
	})
}
