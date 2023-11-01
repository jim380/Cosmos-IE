CREATE TABLE absent_validators (
  block_height INT REFERENCES blocks(height),
  cons_pub_address TEXT REFERENCES validators(cons_pub_address),
  PRIMARY KEY (block_height, cons_pub_address)
);