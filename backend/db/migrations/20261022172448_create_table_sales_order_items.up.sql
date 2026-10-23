CREATE TABLE sales_order_items (
    id          SERIAL PRIMARY KEY,
    order_id    INTEGER NOT NULL REFERENCES sales_orders(id),
    sku         TEXT    NOT NULL REFERENCES products(sku),
    qty         INTEGER NOT NULL CHECK (qty > 0),
    unit_price  INTEGER NOT NULL
);

CREATE INDEX idx_sales_order_items_order_id ON sales_order_items(order_id);
CREATE INDEX idx_sales_order_items_sku      ON sales_order_items(sku);