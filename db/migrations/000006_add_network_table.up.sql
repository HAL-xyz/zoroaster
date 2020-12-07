BEGIN;

ALTER TABLE triggers DROP COLUMN network;

CREATE TABLE IF NOT EXISTS public.networks (
    network_id_name text PRIMARY KEY,
    friendly_name text NOT NULL,
    technology text NOT NULL,
    network_name text NOT NULL,
    endpoint text NOT NULL
);

INSERT INTO networks(network_id_name, friendly_name, technology, network_name, endpoint)
VALUES ('1_eth_mainnet', 'Ethereum Mainnet', 'ETH', 'Mainnet', 'foobar');

ALTER TABLE triggers ADD COLUMN network_id text NOT NULL DEFAULT '1_eth_mainnet';
ALTER TABLE triggers ADD CONSTRAINT fk_networks FOREIGN KEY (network_id) REFERENCES networks(network_id_name);

ALTER TABLE state ADD COLUMN network_id text NOT NULL DEFAULT '1_eth_mainnet';
ALTER TABLE state ADD CONSTRAINT fk_state_network FOREIGN KEY (network_id) REFERENCES networks(network_id_name);

COMMIT;
