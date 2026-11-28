-- up
-- Rebuild sales so the CHECK accepts the 'facebook' channel.
-- The migration runner wraps this file in a transaction with foreign_keys=ON,
-- so children that reference sales are detached first (sale_lines) or have
-- their sale_id nulled (cash_entries, receivables) and restored afterwards.
CREATE TEMP TABLE _sale_lines_backup AS SELECT * FROM sale_lines;
CREATE TEMP TABLE _cash_entries_sale AS
  SELECT id, sale_id FROM cash_entries WHERE sale_id IS NOT NULL;
CREATE TEMP TABLE _receivables_sale AS
  SELECT id, sale_id FROM receivables WHERE sale_id IS NOT NULL;

CREATE TABLE sales_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    item_id INTEGER NOT NULL REFERENCES items(id),
    sold_at TEXT NOT NULL,
    channel TEXT NOT NULL CHECK (channel IN ('direct', 'mercadolivre', 'shopee', 'olx', 'facebook', 'other')),
    gross_cents INTEGER NOT NULL,
    fee_cents INTEGER NOT NULL DEFAULT 0,
    shipping_cents INTEGER NOT NULL DEFAULT 0,
    net_cents INTEGER NOT NULL,
    payment_status TEXT NOT NULL CHECK (payment_status IN ('received', 'pending', 'cancelled')),
    unit_cost_cents_at_sale INTEGER NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    client_id INTEGER REFERENCES clients(id)
);

INSERT INTO sales_new
  (id, item_id, sold_at, channel, gross_cents, fee_cents, shipping_cents,
   net_cents, payment_status, unit_cost_cents_at_sale, created_at, client_id)
SELECT id, item_id, sold_at, channel, gross_cents, fee_cents, shipping_cents,
       net_cents, payment_status, unit_cost_cents_at_sale, created_at, client_id
FROM sales;

DELETE FROM sale_lines;
UPDATE cash_entries SET sale_id = NULL WHERE sale_id IS NOT NULL;
UPDATE receivables SET sale_id = NULL WHERE sale_id IS NOT NULL;

DROP TABLE sales;
ALTER TABLE sales_new RENAME TO sales;

INSERT INTO sale_lines SELECT * FROM _sale_lines_backup;
UPDATE cash_entries
   SET sale_id = (SELECT b.sale_id FROM _cash_entries_sale b WHERE b.id = cash_entries.id)
 WHERE id IN (SELECT id FROM _cash_entries_sale);
UPDATE receivables
   SET sale_id = (SELECT b.sale_id FROM _receivables_sale b WHERE b.id = receivables.id)
 WHERE id IN (SELECT id FROM _receivables_sale);

CREATE INDEX IF NOT EXISTS idx_sales_client_id ON sales(client_id);

DROP TABLE _sale_lines_backup;
DROP TABLE _cash_entries_sale;
DROP TABLE _receivables_sale;

-- down
-- SQLite cannot restore the narrower CHECK without dropping facebook rows.
