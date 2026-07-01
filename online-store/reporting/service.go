package reporting

import (
	"fmt"
	"online-store/models"
	"sort"
)

type ReportingService struct{}

func NewReportingService() *ReportingService {
	return &ReportingService{}
}

func (s *ReportingService) BuildDailyReports(orders []models.Order, catalog []models.Item) ([]models.ShippingReport, error) {
	catalogMap := make(map[string]models.Item)
	for _, item := range catalog {
		catalogMap[item.ItemID] = item
	}

	deduplicator := NewDeduplicator()
	for _, order := range orders {
		err := Validate(order, catalogMap)
		if err != nil {
			fmt.Println("Rejected order ", order.OrderID, ": ", err)
			continue
		}

		deduplicator.Add(order)
	}

	Sort(deduplicator.UniqueOrders)

	groupedUniqueOrders := groupByDate(deduplicator.UniqueOrders)
	groupedDuplicateOrders := groupByDate(deduplicator.DuplicateOrders)

	dates := make([]string, 0, len(groupedUniqueOrders)+len(groupedDuplicateOrders))
	for date := range groupedUniqueOrders {
		dates = append(dates, date)
	}

	for date := range groupedDuplicateOrders {
		if _, exists := groupedUniqueOrders[date]; !exists {
			dates = append(dates, date)
		}
	}

	sort.Strings(dates)

	reports := make([]models.ShippingReport, 0, len(dates))
	for _, date := range dates {
		uniqueOrders := groupedUniqueOrders[date]
		duplicateOrders := groupedDuplicateOrders[date]

		report := GenerateReport(uniqueOrders, duplicateOrders, catalogMap)
		report.Date = date
		reports = append(reports, report)
	}

	return reports, nil
}

func groupByDate(orders []models.Order) map[string][]models.Order {
	groupedOrders := make(map[string][]models.Order)
	for _, order := range orders {
		groupedOrders[order.Date] = append(groupedOrders[order.Date], order)
	}
	return groupedOrders
}
