CREATE TABLE mempool (
    txid CHAR(64) PRIMARY KEY NOT NULL,
    tx BLOB NOT NULL,
    fee INT
);


CREATE INDEX fee_idx ON mempool(fee);