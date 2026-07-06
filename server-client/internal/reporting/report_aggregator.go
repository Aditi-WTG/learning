package reporting

import (
	"maps"
	"server-client/internal/models"
	"sort"
	"sync"
)

type ReportAggregator struct {
	mu            sync.RWMutex
	seenOrderIDs  map[string]struct{}
	reportsByDate map[string]*models.ShippingReport
	catalogMap    map[string]models.Item
}

func NewReportAggregator(catalog []models.Item) *ReportAggregator {
	catalogMap := make(map[string]models.Item, len(catalog))
	for _, item := range catalog {
		catalogMap[item.ItemID] = item
	}

	return &ReportAggregator{
		seenOrderIDs:  make(map[string]struct{}),
		reportsByDate: make(map[string]*models.ShippingReport),
		catalogMap:    catalogMap,
	}
}

func (a *ReportAggregator) ProcessOrder(order models.Order) (models.ShippingReport, error) {
	if err := Validate(order, a.catalogMap); err != nil {
		return models.ShippingReport{}, err
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	report := a.getReport(order.Date)
	if _, exists := a.seenOrderIDs[order.OrderID]; exists {
		a.recordDuplicateOrder(report, order)
		return a.cloneReport(report), nil
	}

	a.seenOrderIDs[order.OrderID] = struct{}{}
	a.addOrderToReport(report, order)

	return a.cloneReport(report), nil
}

func (a *ReportAggregator) Snapshot(date string) (models.ShippingReport, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	report, ok := a.reportsByDate[date]
	if !ok {
		return models.ShippingReport{}, false
	}

	return a.cloneReport(report), true
}

func (a *ReportAggregator) SnapshotAll() []models.ShippingReport {
	a.mu.RLock()
	defer a.mu.RUnlock()

	dates := make([]string, 0, len(a.reportsByDate))
	for date := range a.reportsByDate {
		dates = append(dates, date)
	}

	sort.Strings(dates)

	reports := make([]models.ShippingReport, 0, len(dates))
	for _, date := range dates {
		reports = append(reports, a.cloneReport(a.reportsByDate[date]))
	}

	return reports
}

func (a *ReportAggregator) getReport(date string) *models.ShippingReport {
	report, ok := a.reportsByDate[date]
	if ok {
		return report
	}

	newReport := models.NewShippingReport()
	newReport.Date = date
	a.reportsByDate[date] = &newReport

	return &newReport
}

func (a *ReportAggregator) addOrderToReport(report *models.ShippingReport, order models.Order) {
	report.TotalOrders++

	for _, orderItem := range order.Items {
		item := a.catalogMap[orderItem.ItemID]

		report.TotalCost += float64(orderItem.Quantity) * item.UnitPrice

		itemSummary := report.Items[item.ItemID]
		itemSummary.Name = item.Name
		itemSummary.Quantity += orderItem.Quantity
		report.Items[item.ItemID] = itemSummary
	}

	report.OrdersByDestination[order.Destination]++
	report.OrdersByCustomer[order.CustomerID]++
}

func (a *ReportAggregator) recordDuplicateOrder(report *models.ShippingReport, order models.Order) {
	report.TotalOrders++
	report.DuplicateOrders++
	report.DuplicateOrdersByCustomer[order.CustomerID]++
}

func (a *ReportAggregator) cloneReport(report *models.ShippingReport) models.ShippingReport {
	if report == nil {
		return models.NewShippingReport()
	}

	clone := models.NewShippingReport()
	clone.Date = report.Date
	clone.TotalOrders = report.TotalOrders
	clone.TotalCost = report.TotalCost
	clone.DuplicateOrders = report.DuplicateOrders

	maps.Copy(clone.Items, report.Items)
	maps.Copy(clone.OrdersByDestination, report.OrdersByDestination)
	maps.Copy(clone.OrdersByCustomer, report.OrdersByCustomer)
	maps.Copy(clone.DuplicateOrdersByCustomer, report.DuplicateOrdersByCustomer)

	return clone
}
