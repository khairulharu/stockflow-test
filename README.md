## Create Migrations File Using Golang Migrate

```
migrate create -ext sql -dir db/migrations create_table_products
migrate create -ext sql -dir db/migrations create_table_stock_movements
migrate create -ext sql -dir db/migrations create_table_sales_orders
migrate create -ext sql -dir db/migrations create_table_sales_order_items

```

## Run Docker Compose for PostgresDB

```
docker compose up -d db
```

## Run Migration file using golang migrate

```
migrate -database "postgres://khai:secret@localhost:5432/mydb?sslmode=disable" -path db/migrations up
```

## Run Seed

```
docker exec -i local-postgres psql -U khai -d mydb < ./db/seed/001_seed_products.sql
```
--stop