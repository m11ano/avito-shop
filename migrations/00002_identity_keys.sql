-- +goose Up
ALTER TABLE shop_purchase
    ADD COLUMN identity_key uuid;

CREATE INDEX idx_shop_purchase_identity_key
    ON shop_purchase (identity_key);

ALTER TABLE coin_transfer
    ADD COLUMN identity_key uuid;

CREATE INDEX idx_coin_transfer_identity_key
    ON coin_transfer (identity_key);

ALTER TABLE coin_transfer
    ADD COLUMN created_at TIMESTAMP NOT NULL DEFAULT NOW();

-- +goose Down
DROP INDEX IF EXISTS idx_shop_purchase_identity_key;
ALTER TABLE shop_purchase
    DROP COLUMN IF EXISTS identity_key;

DROP INDEX IF EXISTS idx_coin_transfer_identity_key;
ALTER TABLE coin_transfer
    DROP COLUMN IF EXISTS identity_key;

ALTER TABLE coin_transfer
    DROP COLUMN IF EXISTS created_at;
