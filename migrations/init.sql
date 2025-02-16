CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE users (
                       user_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                       name TEXT NOT NULL,
                       password_hash TEXT NOT NULL,
                       coin_balance INTEGER NOT NULL DEFAULT 1000 CHECK (coin_balance >= 0)
);

CREATE TABLE merch (
                       merch_id SERIAL PRIMARY KEY,
                       name TEXT UNIQUE NOT NULL,
                       price INTEGER NOT NULL CHECK (price > 0)
);

CREATE TABLE user_inventory (
                                user_id UUID NOT NULL,
                                merch_id INTEGER NOT NULL,
                                quantity INTEGER NOT NULL CHECK (quantity >= 0),
                                PRIMARY KEY (user_id, merch_id),
                                FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE,
                                FOREIGN KEY (merch_id) REFERENCES merch(merch_id) ON DELETE CASCADE
);

CREATE TABLE coin_transfers (
                                transfer_id SERIAL PRIMARY KEY,
                                from_user_id UUID NOT NULL,
                                to_user_id UUID NOT NULL,
                                amount INTEGER NOT NULL CHECK (amount > 0),
                                transfer_date TIMESTAMP DEFAULT NOW(),
                                FOREIGN KEY (from_user_id) REFERENCES users(user_id) ON DELETE CASCADE,
                                FOREIGN KEY (to_user_id) REFERENCES users(user_id) ON DELETE CASCADE
);

CREATE TABLE purchases (
                           purchase_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                           user_id UUID NOT NULL,
                           total_price INTEGER NOT NULL CHECK (total_price > 0),
                           purchase_date TIMESTAMP DEFAULT NOW(),
                           FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);

CREATE TABLE purchase_items (
                                purchase_item_id SERIAL PRIMARY KEY,
                                purchase_id UUID NOT NULL,
                                merch_id INTEGER NOT NULL,
                                quantity INTEGER NOT NULL CHECK (quantity > 0),
                                price_at_purchase INTEGER NOT NULL CHECK (price_at_purchase > 0),
                                FOREIGN KEY (purchase_id) REFERENCES purchases(purchase_id) ON DELETE CASCADE,
                                FOREIGN KEY (merch_id) REFERENCES merch(merch_id)
);

INSERT INTO merch (name, price) VALUES
                                    ('t-shirt', 80),
                                    ('cup', 20),
                                    ('book', 50),
                                    ('pen', 10),
                                    ('powerbank', 200),
                                    ('hoody', 300),
                                    ('umbrella', 200),
                                    ('socks', 10),
                                    ('wallet', 50),
                                    ('pink-hoody', 500);

CREATE INDEX idx_coin_transfers_from_user ON coin_transfers (from_user_id);
CREATE INDEX idx_coin_transfers_to_user ON coin_transfers (to_user_id);
CREATE INDEX idx_user_inventory_user ON user_inventory (user_id);
CREATE INDEX idx_purchases_user ON purchases (user_id);
