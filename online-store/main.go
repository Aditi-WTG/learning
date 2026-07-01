package main

import (
	"encoding/json"
	"fmt"
	"online-store/models"
	"online-store/reporting"
	"path/filepath"
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

	reportingService := reporting.NewReportingService()
	reports, err := reportingService.BuildDailyReports(orders, catalog)
	if err != nil {
		panic("Error generating reports: " + err.Error())
	}

	for _, report := range reports {
		jsonReport, err := json.Marshal(report)
		if err != nil {
			panic("Error marshalling the report: " + err.Error())
		}

		fmt.Println("Report for: ", report.Date)
		fmt.Println(string(jsonReport))
	}
}
