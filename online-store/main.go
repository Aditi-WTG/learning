package main

import (
	"encoding/json"
	"fmt"
	"online-store/models"
	"path/filepath"
	"sort"
)

func main() {

	orders, err := Load[models.Order](filepath.Join("data", "orders.json"))
	if err != nil {
		panic("Error loading the orders: " + err.Error())
	}

	catalog, err := Load[models.Item](filepath.Join("data", "catalog.json"))
	if err != nil {
		panic("Error loading the catalog: " + err.Error())
	}

	catalogMap := make(map[string]models.Item)
	for _, item := range catalog {
		catalogMap[item.ItemID] = item
	}

	processor := models.NewOrderProcessor()
	for _, order := range orders {
		err = Validate(order, catalogMap)
		if err != nil {
			fmt.Println("Rejected order ", order.OrderID, ": ", err)
			continue
		}

		processor.Process(order)
	}

	Sort(processor.UniqueOrders)

	groupedUniqueOrders := GroupByDate(processor.UniqueOrders)
	groupedDuplicateOrders := GroupByDate(processor.DuplicateOrders)

	dates := make([]string, 0, len(groupedUniqueOrders)+len(groupedDuplicateOrders))

	for date := range groupedUniqueOrders {
		dates = append(dates, date)
	}

	for date := range groupedDuplicateOrders {
		_, exists := groupedUniqueOrders[date]
		if !exists {
			dates = append(dates, date)
		}
	}

	sort.Strings(dates)

	for _, date := range dates {
		orders := groupedUniqueOrders[date]
		duplicateOrders := groupedDuplicateOrders[date]

		report := generateReport(orders, duplicateOrders, catalogMap)
		report.Date = date

		jsonReport, err := json.Marshal(report)
		if err != nil {
			panic("Error marshalling the report: " + err.Error())
		}

		fmt.Println("Report for: ", date)
		fmt.Println(string(jsonReport))
	}
}

func GroupByDate(orders []models.Order) map[string][]models.Order {
	groupedOrders := make(map[string][]models.Order)
	for _, order := range orders {
		groupedOrders[order.Date] = append(groupedOrders[order.Date], order)
	}
	return groupedOrders
}
