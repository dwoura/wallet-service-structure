-- 1. Users Table
CREATE TABLE IF NOT EXISTS users (
    id bigserial PRIMARY KEY,
    username varchar(255) NOT NULL UNIQUE,
    email varchar(255) NOT NULL UNIQUE,
    password_hash varchar(255) NOT NULL,
    created_at timestamptz,
    updated_at timestamptz,
    deleted_at timestamptz
);
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at);

-- 2. Accounts Table
CREATE TABLE IF NOT EXISTS accounts (
    id bigserial PRIMARY KEY,
    user_id bigint NOT NULL,
    currency varchar(10) NOT NULL,
    balance numeric(32,18) NOT NULL DEFAULT 0,
    locked_balance numeric(32,18) NOT NULL DEFAULT 0,
    version bigint NOT NULL DEFAULT 0,
    created_at timestamptz,
    updated_at timestamptz,
    CONSTRAINT fk_users_accounts FOREIGN KEY(user_id) REFERENCES users(id)
);
CREATE INDEX IF NOT EXISTS idx_accounts_user_id ON accounts(user_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_user_currency ON accounts(user_id, currency);

-- 3. Addresses Table
CREATE TABLE IF NOT EXISTS addresses (
    id bigserial PRIMARY KEY,
    user_id bigint NOT NULL,
    chain varchar(20) NOT NULL,
    address varchar(255) NOT NULL,
    hd_path_index integer NOT NULL,
    created_at timestamptz,
    CONSTRAINT fk_users_addresses FOREIGN KEY(user_id) REFERENCES users(id)
);
CREATE INDEX IF NOT EXISTS idx_addresses_user_id ON addresses(user_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_chain_address ON addresses(chain, address);
CREATE UNIQUE INDEX IF NOT EXISTS idx_chain_path ON addresses(chain, hd_path_index);

-- 4. Deposits Table
CREATE TABLE IF NOT EXISTS deposits (
    id bigserial PRIMARY KEY,
    user_id bigint NOT NULL,
    block_app_id bigint NOT NULL,
    tx_hash varchar(255) NOT NULL,
    amount numeric(32,18) NOT NULL,
    block_height bigint NOT NULL,
    status varchar(20) NOT NULL,
    created_at timestamptz,
    confirmed_at timestamptz
);
CREATE INDEX IF NOT EXISTS idx_deposits_user_id ON deposits(user_id);
CREATE INDEX IF NOT EXISTS idx_deposits_block_app_id ON deposits(block_app_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_tx_app ON deposits(tx_hash, block_app_id); -- Assuming unique per tx per app-address

-- 5. Withdrawals Table
CREATE TABLE IF NOT EXISTS withdrawals (
    id bigserial PRIMARY KEY,
    user_id bigint NOT NULL,
    to_address varchar(255) NOT NULL,
    amount numeric(32,18) NOT NULL,
    chain varchar(20) NOT NULL,
    tx_hash varchar(255),
    status varchar(20) NOT NULL,
    created_at timestamptz,
    updated_at timestamptz
);
CREATE INDEX IF NOT EXISTS idx_withdrawals_user_id ON withdrawals(user_id);

-- 6. Collections Table
CREATE TABLE IF NOT EXISTS collections (
    id bigserial PRIMARY KEY,
    deposit_id bigint NOT NULL,
    tx_hash varchar(66) NOT NULL,
    from_address varchar(42) NOT NULL,
    to_address varchar(42) NOT NULL,
    amount numeric(30,0) NOT NULL,
    gas_fee numeric(30,0) NOT NULL,
    status varchar(20) NOT NULL DEFAULT 'pending',
    created_at timestamptz,
    updated_at timestamptz
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_collections_deposit_id ON collections(deposit_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_collections_tx_hash ON collections(tx_hash);

-- 7. OutboxMessages Table
CREATE TABLE IF NOT EXISTS outbox_messages (
    id bigserial PRIMARY KEY,
    topic varchar(255) NOT NULL,
    payload text NOT NULL,
    status varchar(50) NOT NULL DEFAULT 'PENDING',
    created_at timestamptz,
    updated_at timestamptz,
    deleted_at timestamptz
);
CREATE INDEX IF NOT EXISTS idx_outbox_messages_status ON outbox_messages(status);
CREATE INDEX IF NOT EXISTS idx_outbox_messages_deleted_at ON outbox_messages(deleted_at);
