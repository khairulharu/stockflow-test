## Create Migrations File Using Golang Migrate
```
migrate create -ext sql -dir db/migrations create_table_products
migrate create -ext sql -dir db/migrations create_table_stock_movements
migrate create -ext sql -dir db/migrations create_table_sales_orders
migrate create -ext sql -dir db/migrations create_table_sales_order_items

```