-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE account (
    account_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(66),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE operation (
    operation_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    operation_type SMALLINT NOT NULL,
    account_id UUID NOT NULL REFERENCES account(account_id) ON DELETE RESTRICT,  -- Запрещаем удаление пользователя, если есть записи
    amount BIGINT NOT NULL,
    source_type SMALLINT NOT NULL,
    source_id UUID NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Добавляем индекс для ускорения агрегации баланса + подключаем amount, чтобы индекс не заглядывал в кучу таблицы
CREATE INDEX idx_operations_account_id ON operation(account_id) INCLUDE (amount);

CREATE TABLE shop_item (
    item_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    item_name VARCHAR(100) NOT NULL UNIQUE,
    price BIGINT NOT NULL
);

INSERT INTO shop_item (item_id, item_name, price) VALUES
(uuid_generate_v4(), 't-shirt', 80),
(uuid_generate_v4(), 'cup', 20),
(uuid_generate_v4(), 'book', 50),
(uuid_generate_v4(), 'pen', 10),
(uuid_generate_v4(), 'powerbank', 200),
(uuid_generate_v4(), 'hoody', 300),
(uuid_generate_v4(), 'umbrella', 200),
(uuid_generate_v4(), 'socks', 10),
(uuid_generate_v4(), 'wallet', 50),
(uuid_generate_v4(), 'pink-hoody', 500);

CREATE TABLE shop_purchase (
    purchase_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    item_id UUID NOT NULL REFERENCES shop_item(item_id) ON DELETE RESTRICT, -- Запрещаем удаление товара, если есть записи
    account_id UUID NOT NULL REFERENCES account(account_id) ON DELETE RESTRICT, -- Запрещаем удаление пользователя, если есть записи
    quantity BIGINT NOT NULL CHECK (quantity > 0),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_shop_purchase_account_id ON shop_purchase(account_id);
CREATE INDEX idx_shop_purchase_item_id ON shop_purchase(item_id);

CREATE TABLE coin_transfer (
    transfer_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    transfer_type SMALLINT NOT NULL,
    owner_account_id UUID NOT NULL REFERENCES account(account_id) ON DELETE RESTRICT, -- Запрещаем удаление пользователя, если есть записи
    counterparty_account_id UUID NOT NULL REFERENCES account(account_id) ON DELETE RESTRICT, -- Запрещаем удаление пользователя, если есть записи
    amount BIGINT NOT NULL CHECK (amount > 0)
);

CREATE INDEX idx_coin_transfer_owner_account_id ON coin_transfer(owner_account_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS coin_transfer;
DROP TABLE IF EXISTS shop_purchase;
DROP TABLE IF EXISTS shop_item;
DROP TABLE IF EXISTS operation;
DROP TABLE IF EXISTS account;
-- +goose StatementEnd