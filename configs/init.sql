CREATE TABLE IF NOT EXISTS users (
    user_id SERIAL PRIMARY KEY,
    username VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL
);
INSERT INTO users (username, email)
VALUES ('john', 'john@example.com');
INSERT INTO users (username, email)
VALUES ('tom', 'tom@example.com');

CREATE TABLE IF NOT EXISTS wallets (
    wallet_id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(user_id) ON DELETE CASCADE,
    balance DECIMAL(15, 2) DEFAULT 0.00
);
INSERT INTO wallets (user_id, balance)
VALUES
    (1, 100.00),
    (2, 50.00);

CREATE TABLE IF NOT EXISTS transactions (
    transaction_id SERIAL PRIMARY KEY,
    from_user_id INT REFERENCES users(user_id) ON DELETE CASCADE,
    to_user_id INT REFERENCES users(user_id),
    amount DECIMAL(15, 2),
    transaction_type VARCHAR(20), -- 'deposit', 'withdrawal', 'transfer'
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
