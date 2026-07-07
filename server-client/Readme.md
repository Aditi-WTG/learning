# Simple Go gRPC Publisher-Consumer

This project is a gRPC-based event service for online-store reporting.

Current flow:
1. Publisher sends `order.created` events.
2. Service validates and processes orders into daily aggregates.
3. Service emits `report.daily` events.
4. Subscribers consume `report.daily` updates.
5. Clients can query reports using pull RPCs (`GetReportByDate`, `GetAllReports`).

## Available Client Modes

1. `-mode=pub` requires `-topic` and `-message`
2. `-mode=sub` requires `-topic`
3. `-mode=report` requires `-date`
4. `-mode=report-all` no extra flag required

## Run

Terminal 1 (server):

```powershell
go run ./cmd/server
```

Terminal 2 (subscribe to daily reports):

```powershell
go run ./cmd/client -mode=sub -topic='report.daily'
```

Terminal 3 (publish an order):

```powershell
go run ./cmd/client -mode=pub -topic='order.created' -message='{"orderId":"O1001","customerId":"C001","items":[{"itemId":"I001","quantity":2}],"destination":"Bangalore","date":"2026-07-06"}'
```

Query one day report:

```powershell
go run ./cmd/client -mode=report -date=2026-07-06
```

Query all reports:

```powershell
go run ./cmd/client -mode=report-all
```