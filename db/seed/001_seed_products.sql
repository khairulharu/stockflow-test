BEGIN;

INSERT INTO products (sku, name, price, stock, is_active) VALUES
    ('BRS-001', 'Beras Premium 5kg',           68000,  120, TRUE),
    ('MNY-001', 'Minyak Goreng 2L',            38000,   80, TRUE),
    ('GLA-001', 'Gula Pasir 1kg',              17500,  150, TRUE),
    ('TLR-001', 'Telur Ayam 1kg',              29000,   60, TRUE),
    ('KOP-001', 'Kopi Bubuk 200g',             22000,   45, TRUE),
    ('TEH-001', 'Teh Celup isi 25',            11500,  200, TRUE),
    ('MIE-001', 'Mie Instan Goreng',            3500,  500, TRUE),
    ('SUS-001', 'Susu UHT Coklat 1L',          19500,   70, TRUE),
    ('SBN-001', 'Sabun Mandi Cair 400ml',      24000,   35, TRUE),
    ('DTR-001', 'Deterjen Bubuk 800g',         21000,   55, TRUE),
    ('AIR-001', 'Air Mineral 600ml',            3000,  300, TRUE),
    ('SNK-001', 'Keripik Kentang 70g',         10500,    0, TRUE),
    ('OLD-001', 'Permen Jahe (discontinued)',   5000,   10, FALSE)
ON CONFLICT (sku) DO NOTHING;

INSERT INTO stock_movements (product_id, type, qty, note)
SELECT p.id, 'in', p.stock, 'initial stock'
FROM products p
WHERE p.stock > 0
  AND NOT EXISTS (
      SELECT 1 FROM stock_movements sm
      WHERE sm.product_id = p.id AND sm.note = 'initial stock'
  );

COMMIT;