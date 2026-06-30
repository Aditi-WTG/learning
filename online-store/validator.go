package main

import (
	"fmt"
	"online-store/models"
	"time"
)

func Validate(order models.Order, catalogMap map[string]models.Item) error {
	if order.OrderID == "" {
		return fmt.Errorf("Missing Order Id")
	}

	if order.CustomerID == "" {
		return fmt.Errorf("Missing Customer Id")
	}

	if len(order.Items) == 0 {
		return fmt.Errorf("Missing Order Items")
	}

	for _, orderItem := range order.Items {

		if orderItem.Quantity <= 0 {
			return fmt.Errorf("Invalid Order Item Quantity")
		}

		_, ok := catalogMap[orderItem.ItemID]
		if !ok {
			return fmt.Errorf("Invalid Order Item Id")
		}
	}

	if order.Destination == "" {
		return fmt.Errorf("Missing Destination")
	}

	if order.Date == "" {
		return fmt.Errorf("Missing Date")
	}

	_, err := time.Parse("2006-01-02", order.Date)
	if err != nil {
		return fmt.Errorf("Invalid Date")
	}

	return nil
}
