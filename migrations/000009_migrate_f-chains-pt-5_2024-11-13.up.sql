-- F000
INSERT INTO chain_alias_domains (chain_id, alias, domains)
SELECT 'F000', alias, ARRAY [alias || '.gateway.grove.city']
FROM chain_alias_domains WHERE chain_id = ANY(ARRAY['0001', 'A001'])
ON CONFLICT (chain_id, alias) DO NOTHING;
UPDATE chains SET active = false WHERE id = ANY(ARRAY['0001', 'A001']);

-- F002
INSERT INTO chain_alias_domains (chain_id, alias, domains)
SELECT 'F002', alias, ARRAY [alias || '.gateway.grove.city']
FROM chain_alias_domains WHERE chain_id = 'A086'
ON CONFLICT (chain_id, alias) DO NOTHING;
UPDATE chains SET active = false WHERE id = 'A086';

-- F003
INSERT INTO chain_alias_domains (chain_id, alias, domains)
SELECT 'F003', alias, ARRAY [alias || '.gateway.grove.city']
FROM chain_alias_domains WHERE chain_id = ANY(ARRAY['0003', 'A003'])
ON CONFLICT (chain_id, alias) DO NOTHING;
UPDATE chains SET active = false WHERE id = ANY(ARRAY['0003', 'A003']);

-- F004
INSERT INTO chain_alias_domains (chain_id, alias, domains)
SELECT 'F004', alias, ARRAY [alias || '.gateway.grove.city']
FROM chain_alias_domains WHERE chain_id = '03DF'
ON CONFLICT (chain_id, alias) DO NOTHING;
UPDATE chains SET active = false WHERE id = '03DF';

-- F008
INSERT INTO chain_alias_domains (chain_id, alias, domains)
SELECT 'F008', alias, ARRAY [alias || '.gateway.grove.city']
FROM chain_alias_domains WHERE chain_id = 'A090'
ON CONFLICT (chain_id, alias) DO NOTHING;
UPDATE chains SET active = false WHERE id = 'A090';

-- F009
INSERT INTO chain_alias_domains (chain_id, alias, domains)
SELECT 'F009', alias, ARRAY [alias || '.gateway.grove.city']
FROM chain_alias_domains WHERE chain_id = ANY(ARRAY['0004', '0010'])
ON CONFLICT (chain_id, alias) DO NOTHING;
UPDATE chains SET active = false WHERE id = ANY(ARRAY['0004', '0010']);

-- F00B
INSERT INTO chain_alias_domains (chain_id, alias, domains)
SELECT 'F00B', alias, ARRAY [alias || '.gateway.grove.city']
FROM chain_alias_domains WHERE chain_id = '0065'
ON CONFLICT (chain_id, alias) DO NOTHING;
UPDATE chains SET active = false WHERE id = '0065';

-- F00D
INSERT INTO chain_alias_domains (chain_id, alias, domains)
SELECT 'F00D', alias, ARRAY [alias || '.gateway.grove.city']
FROM chain_alias_domains WHERE chain_id = '0081'
ON CONFLICT (chain_id, alias) DO NOTHING;
UPDATE chains SET active = false WHERE id = '0081';

-- F00F
INSERT INTO chain_alias_domains (chain_id, alias, domains)
SELECT 'F00F', alias, ARRAY [alias || '.gateway.grove.city']
FROM chain_alias_domains WHERE chain_id = '0046'
ON CONFLICT (chain_id, alias) DO NOTHING;
UPDATE chains SET active = false WHERE id = '0046';

-- F013
INSERT INTO chain_alias_domains (chain_id, alias, domains)
SELECT 'F013', alias, ARRAY [alias || '.gateway.grove.city']
FROM chain_alias_domains WHERE chain_id = ANY(ARRAY['0027', '000C'])
ON CONFLICT (chain_id, alias) DO NOTHING;
UPDATE chains SET active = false WHERE id = ANY(ARRAY['0027', '000C']);

-- F014
INSERT INTO chain_alias_domains (chain_id, alias, domains)
SELECT 'F014', alias, ARRAY [alias || '.gateway.grove.city']
FROM chain_alias_domains WHERE chain_id = '0040'
ON CONFLICT (chain_id, alias) DO NOTHING;
UPDATE chains SET active = false WHERE id = '0040';

-- F015
INSERT INTO chain_alias_domains (chain_id, alias, domains)
SELECT 'F015', alias, ARRAY [alias || '.gateway.grove.city']
FROM chain_alias_domains WHERE chain_id = '0044'
ON CONFLICT (chain_id, alias) DO NOTHING;
UPDATE chains SET active = false WHERE id = '0044';

-- F019
INSERT INTO chain_alias_domains (chain_id, alias, domains)
SELECT 'F019', alias, ARRAY [alias || '.gateway.grove.city']
FROM chain_alias_domains WHERE chain_id = '0050'
ON CONFLICT (chain_id, alias) DO NOTHING;
UPDATE chains SET active = false WHERE id = '0050';

-- F01B
INSERT INTO chain_alias_domains (chain_id, alias, domains)
SELECT 'F01B', alias, ARRAY [alias || '.gateway.grove.city']
FROM chain_alias_domains WHERE chain_id = '0052'
ON CONFLICT (chain_id, alias) DO NOTHING;
UPDATE chains SET active = false WHERE id = '0052';

-- F01D
INSERT INTO chain_alias_domains (chain_id, alias, domains)
SELECT 'F01D', alias, ARRAY [alias || '.gateway.grove.city']
FROM chain_alias_domains WHERE chain_id = ANY(ARRAY['0053', 'A053'])
ON CONFLICT (chain_id, alias) DO NOTHING;
UPDATE chains SET active = false WHERE id = ANY(ARRAY['0053', 'A053']);

-- F01E
INSERT INTO chain_alias_domains (chain_id, alias, domains)
SELECT 'F01E', alias, ARRAY [alias || '.gateway.grove.city']
FROM chain_alias_domains WHERE chain_id = 'A087'
ON CONFLICT (chain_id, alias) DO NOTHING;
UPDATE chains SET active = false WHERE id = 'A087';

-- F01F
INSERT INTO chain_alias_domains (chain_id, alias, domains)
SELECT 'F01F', alias, ARRAY [alias || '.gateway.grove.city']
FROM chain_alias_domains WHERE chain_id = 'A089'
ON CONFLICT (chain_id, alias) DO NOTHING;
UPDATE chains SET active = false WHERE id = 'A089';

-- F020
INSERT INTO chain_alias_domains (chain_id, alias, domains)
SELECT 'F020', alias, ARRAY [alias || '.gateway.grove.city']
FROM chain_alias_domains WHERE chain_id = '0054'
ON CONFLICT (chain_id, alias) DO NOTHING;
UPDATE chains SET active = false WHERE id = '0054';

-- F022
INSERT INTO chain_alias_domains (chain_id, alias, domains)
SELECT 'F022', alias, ARRAY [alias || '.gateway.grove.city']
FROM chain_alias_domains WHERE chain_id = 'A085'
ON CONFLICT (chain_id, alias) DO NOTHING;
UPDATE chains SET active = false WHERE id = 'A085';

-- F023
INSERT INTO chain_alias_domains (chain_id, alias, domains)
SELECT 'F023', alias, ARRAY [alias || '.gateway.grove.city']
FROM chain_alias_domains WHERE chain_id = '0083'
ON CONFLICT (chain_id, alias) DO NOTHING;
UPDATE chains SET active = false WHERE id = '0083';

-- F027
INSERT INTO chain_alias_domains (chain_id, alias, domains)
SELECT 'F027', alias, ARRAY [alias || '.gateway.grove.city']
FROM chain_alias_domains WHERE chain_id = '7A00'
ON CONFLICT (chain_id, alias) DO NOTHING;
UPDATE chains SET active = false WHERE id = '7A00';

-- F028
INSERT INTO chain_alias_domains (chain_id, alias, domains)
SELECT 'F028', alias, ARRAY [alias || '.gateway.grove.city']
FROM chain_alias_domains WHERE chain_id = '7A10'
ON CONFLICT (chain_id, alias) DO NOTHING;
UPDATE chains SET active = false WHERE id = '7A10';

-- F02A
INSERT INTO chain_alias_domains (chain_id, alias, domains)
SELECT 'F02A', alias, ARRAY [alias || '.gateway.grove.city']
FROM chain_alias_domains WHERE chain_id = 'A091'
ON CONFLICT (chain_id, alias) DO NOTHING;
UPDATE chains SET active = false WHERE id = 'A091';

-- F02B
INSERT INTO chain_alias_domains (chain_id, alias, domains)
SELECT 'F02B', alias, ARRAY [alias || '.gateway.grove.city']
FROM chain_alias_domains WHERE chain_id = '0084'
ON CONFLICT (chain_id, alias) DO NOTHING;
UPDATE chains SET active = false WHERE id = '0084';

-- F024
INSERT INTO chain_alias_domains (chain_id, alias, domains)
SELECT 'F024', alias, ARRAY [alias || '.gateway.grove.city']
FROM chain_alias_domains WHERE chain_id = '0072'
ON CONFLICT (chain_id, alias) DO NOTHING;
UPDATE chains SET active = false WHERE id = '0072';
