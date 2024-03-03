CREATE TABLE akash_provider_auditors (
  id SERIAL PRIMARY KEY,
  provider_owner TEXT NOT NULL,
  auditor TEXT NOT NULL,
  FOREIGN KEY (provider_owner) REFERENCES akash_providers(owner)
);