# Online store

You're building the backend for an online store. Every day, thousands of customer orders arrive. Your job is to:

- Maintain a backend catalog of items(Item ID,Name, Unit Price). we can maintain a master data file.
- Receive orders from customers(OrderID, CustomerID, List of Items with quantity, Destination, date). We can use a file as input(assume that this file contains orders from multiple sources).
- Process them (validate, deduplicate, sort) and generate report.

Generate a daily shipping manifest report per day(one per date)

- total Orders received
- total cost of orders combined
- list of Items and respective quantities ordered
- total orders per destination
- total duplicate orders(by order ID)
- total orders per customer
- total duplicate orders per customer
