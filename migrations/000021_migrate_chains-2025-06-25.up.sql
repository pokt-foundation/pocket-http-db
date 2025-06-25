-- This is a mega cleanup script designed to refactor the Portal DB significantly and bring it into a PATH-ready state.
BEGIN;

-- Drop unused tables
DROP TABLE chain_aliases;
DROP TABLE chain_altruists;
DROP TABLE chain_checks;
DROP TABLE chain_config;
DROP TABLE chains_gigastake_applications;
DROP TABLE config_gateway;
DROP TABLE config_openapi;
DROP TABLE method_validators;
DROP TABLE portal_application_aats;
DROP TABLE protocols;

-- Clean up chains table
/* Here we are surgically removing all chains that are legacy (predate F-chains
   and some duplicate or deprecated F chains. 
   Notable removals: 
   - F02C - xrplevm devnet (removed from Portal and no longer supporting)
   - F031 - Berachain bArtio Testnet (deprecated by Berachain) */
DELETE FROM chains 
WHERE id IN (
    '0001', '0003', '0004', '0005', '0006', '0009', '000A', '000B', 
    '000C', '000D', '000F', '0010', '0021', '0022', '0023', '0024', 
    '0025', '0026', '0027', '0028', '0029', '0040', '0044', '0046', 
    '0047', '0048', '0049', '0050', '0051', '0052', '0053', '0054', 
    '0055', '0056', '0057', '0058', '0059', '0060', '0061', '0062', 
    '0063', '0065', '0066', '0067', '0068', '0069', '0070', '0071', 
    '0072', '0073', '0074', '0075', '0076', '0077', '0078', '0079', 
    '0080', '0081', '0082', '0083', '0084', '03CB', '03DF', 'A001', 
    'A003', 'A053', 'A085', 'A086', 'A087', 'A088', 'A089', 'A090', 
    'A091', 'B021', 'B081', 'BE2A', 'C006', 'E021', 'F02C', 'F031'
);

-- Clean up chain_alias_domains table
DELETE FROM chain_alias_domains 
WHERE id IN (
    '0001', '0003', '0004', '0005', '0006', '0009', '000A', '000B', 
    '000C', '000D', '000F', '0010', '0021', '0022', '0023', '0024', 
    '0025', '0026', '0027', '0028', '0029', '0040', '0044', '0046', 
    '0047', '0048', '0049', '0050', '0051', '0052', '0053', '0054', 
    '0055', '0056', '0057', '0058', '0059', '0060', '0061', '0062', 
    '0063', '0065', '0066', '0067', '0068', '0069', '0070', '0071', 
    '0072', '0074', '0075', '0076', '0077', '0078', '0079', '0080', 
    '0081', '0082', '0083', '0084', '03CB', '03DF', 'A001', 'A003', 
    'A053', 'A085', 'A086', 'A087', 'A088', 'A089', 'A090', 'A091', 
    'B081', 'BE2A', 'C006', 'E021', 'F02C', 'F031'
);

-- Update tables to have new 42 char limit on ID and 169 on Name
ALTER TABLE chains 
ALTER COLUMN id TYPE varchar(42),
ALTER COLUMN blockchain TYPE varchar(169);

ALTER TABLE chain_alias_domains
ALTER COLUMN chain_id TYPE varchar(42);

-- Clean up all other chains to meet new onchain IDs
-- Clean up chains table first
UPDATE chains 
SET id = v.new_id,
    blockchain = v.new_description,
    description = v.new_description
FROM (VALUES 
    ('F000', 'pocket', 'Pocket Network'),
    ('F001', 'arb-one', 'Arbitrum One'),
    ('F002', 'arb-sepolia-testnet', 'Arbitrum Sepolia Testnet'),
    ('F003', 'avax', 'Avalanche'),
    ('F004', 'avax-dfk', 'DeFi Kingdoms'),
    ('F005', 'base', 'Base'),
    ('F006', 'base-sepolia-testnet', 'Base Sepolia Testnet'),
    ('F007', 'bitcoin', 'Bitcoin'),
    ('F008', 'blast', 'Blast'),
    ('F009', 'bsc', 'BNB Smart Chain'),
    ('F00A', 'boba', 'Boba'),
    ('F00B', 'celo', 'Celo'),
    ('F00C', 'eth', 'Ethereum'),
    ('F00D', 'eth-holesky-testnet', 'Ethereum Holesky Testnet'),
    ('F00E', 'eth-sepolia-testnet', 'Ethereum Sepolia Testnet'),
    ('F00F', 'evmos', 'Evmos'),
    ('F010', 'fantom', 'Fantom'),
    ('F011', 'fraxtal', 'Fraxtal'),
    ('F012', 'fuse', 'Fuse'),
    ('F013', 'gnosis', 'Gnosis'),
    ('F014', 'harmony', 'Harmony Shard 0'),
    ('F015', 'iotex', 'IoTeX'),
    ('F016', 'kaia', 'Kaia'),
    ('F017', 'kava', 'Kava'),
    ('F018', 'metis', 'Metis'),
    ('F019', 'moonbeam', 'Moonbeam'),
    ('F01A', 'moonriver', 'Moonriver'),
    ('F01B', 'near', 'NEAR'),
    ('F01C', 'oasys', 'Oasys'),
    ('F01D', 'op', 'Optimism'),
    ('F01E', 'op-sepolia-testnet', 'Optimism Sepolia Testnet'),
    ('F01F', 'opbnb', 'opBNB'),
    ('F020', 'osmosis', 'Osmosis'),
    ('F021', 'poly', 'Polygon'),
    ('F022', 'poly-amoy-testnet', 'Polygon Amoy Testnet'),
    ('F023', 'radix', 'Radix'),
    ('F024', 'scroll', 'Scroll'),
    ('F025', 'solana', 'Solana'),
    ('F026', 'sui', 'Sui'),
    ('F027', 'taiko', 'Taiko'),
    ('F028', 'taiko-hekla-testnet', 'Taiko Hekla Testnet'),
    ('F029', 'poly-zkevm', 'Polygon zkEVM'),
    ('F02A', 'zklink-nova', 'zkLink Nova'),
    ('F02B', 'zksync-era', 'zkSync Era'),
    ('F02D', 'sonic', 'Sonic'),
    ('F02E', 'tron', 'TRON'),
    ('F030', 'linea', 'Linea'),
    ('F031', 'bera', 'Berachain'),
    ('F032', 'ink', 'Ink'),
    ('F033', 'mantle', 'Mantle'),
    ('F034', 'sei', 'Sei'),
    ('F035', 'bera', 'Berachain'),
    ('F036', 'xrplevm-testnet', 'XRPLEVM Testnet'),
    ('1234', 'xrplevm', 'XRPLEVM'),
    ('1235', 'pocket-alpha-testnet', 'Pocket Alpha Testnet'),
    ('1236', 'pocket-beta-testnet', 'Pocket Beta Testnet');
) AS v(current_id, new_id, new_description)
WHERE chains.id = v.current_id;

-- Update chain_alias_domains
-- Update to new chain_ids
UPDATE chain_alias_domains 
SET chain_id = v.new_id
FROM (VALUES 
    ('F000', 'pocket'),
    ('F001', 'arb-one'),
    ('F002', 'arb-sepolia-testnet'),
    ('F003', 'avax'),
    ('F004', 'avax-dfk'),
    ('F005', 'base'),
    ('F006', 'base-sepolia-testnet'),
    ('F007', 'bitcoin'),
    ('F008', 'blast'),
    ('F009', 'bsc'),
    ('F00A', 'boba'),
    ('F00B', 'celo'),
    ('F00C', 'eth'),
    ('F00D', 'eth-holesky-testnet'),
    ('F00E', 'eth-sepolia-testnet'),
    ('F00F', 'evmos'),
    ('F010', 'fantom'),
    ('F011', 'fraxtal'),
    ('F012', 'fuse'),
    ('F013', 'gnosis'),
    ('F014', 'harmony'),
    ('F015', 'iotex'),
    ('F016', 'kaia'),
    ('F017', 'kava'),
    ('F018', 'metis'),
    ('F019', 'moonbeam'),
    ('F01A', 'moonriver'),
    ('F01B', 'near'),
    ('F01C', 'oasys'),
    ('F01D', 'op'),
    ('F01E', 'op-sepolia-testnet'),
    ('F01F', 'opbnb'),
    ('F020', 'osmosis'),
    ('F021', 'poly'),
    ('F022', 'poly-amoy-testnet'),
    ('F023', 'radix'),
    ('F024', 'scroll'),
    ('F025', 'solana'),
    ('F026', 'sui'),
    ('F027', 'taiko'),
    ('F028', 'taiko-hekla-testnet'),
    ('F029', 'poly-zkevm'),
    ('F02A', 'zklink-nova'),
    ('F02B', 'zksync-era'),
    ('F02D', 'sonic'),
    ('F02E', 'tron'),
    ('F030', 'linea'),
    ('F031', 'bera'),
    ('F032', 'ink'),
    ('F033', 'mantle'),
    ('F034', 'sei'),
    ('F035', 'bera'),
    ('F036', 'xrplevm-testnet'),
    ('1234', 'xrplevm'),
    ('1235', 'pocket-alpha-testnet'),
    ('1236', 'pocket-beta-testnet');
) AS v(current_id, new_id)
WHERE chain_alias_domains.chain_id = v.current_id;

-- Update aliases where appropriate
INSERT INTO chain_alias_domains (chain_id, alias, domain)
SELECT 
    v.new_id,
    v.new_id,
    ARRAY[v.new_id || '.gateway.grove.city']
FROM (VALUES 
    ('pocket'),
    ('arb-one'),
    ('arb-sepolia-testnet'),
    ('avax'),
    ('avax-dfk'),
    ('base'),
    ('base-sepolia-testnet'),
    ('bitcoin'),
    ('blast'),
    ('bsc'),
    ('boba'),
    ('celo'),
    ('eth'),
    ('eth-holesky-testnet'),
    ('eth-sepolia-testnet'),
    ('evmos'),
    ('fantom'),
    ('fraxtal'),
    ('fuse'),
    ('gnosis'),
    ('harmony'),
    ('iotex'),
    ('kaia'),
    ('kava'),
    ('metis'),
    ('moonbeam'),
    ('moonriver'),
    ('near'),
    ('oasys'),
    ('op'),
    ('op-sepolia-testnet'),
    ('opbnb'),
    ('osmosis'),
    ('poly'),
    ('poly-amoy-testnet'),
    ('radix'),
    ('scroll'),
    ('solana'),
    ('sui'),
    ('taiko'),
    ('taiko-hekla-testnet'),
    ('poly-zkevm'),
    ('zklink-nova'),
    ('zksync-era'),
    ('sonic'),
    ('tron'),
    ('linea'),
    ('bera'),
    ('ink'),
    ('mantle'),
    ('sei'),
    ('xrplevm-testnet'),
    ('xrplevm'),
    ('pocket-alpha-testnet'),
    ('pocket-beta-testnet');
) AS v(new_id)
WHERE NOT EXISTS (
    SELECT 1 FROM chain_alias_domains 
    WHERE chain_id = v.new_id AND alias = v.new_id
);

COMMIT;
