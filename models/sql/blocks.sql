CREATE TABLE blocks (
  height INT PRIMARY KEY UNIQUE NOT NULL,
  block_hash TEXT NOT NULL,
  proposer_address TEXT NOT NULL,
  txn_count INT NOT NULL,
  timestamp TIMESTAMP NOT NULL,
);