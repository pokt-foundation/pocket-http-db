-- F001
DELETE FROM chain_alias_domains WHERE chain_id = 'F001';
UPDATE chains SET active = true WHERE id = '0066';

-- F011
DELETE FROM chain_alias_domains WHERE chain_id = 'F011';
UPDATE chains SET active = true WHERE id = 'A088';

-- F016
DELETE FROM chain_alias_domains WHERE chain_id = 'F016';
UPDATE chains SET active = true WHERE id = '0056';

-- F01A
DELETE FROM chain_alias_domains WHERE chain_id = 'F01A';
UPDATE chains SET active = true WHERE id = '0051';

-- F021
DELETE FROM chain_alias_domains WHERE chain_id = 'F021';
UPDATE chains SET active = true WHERE id = ANY(ARRAY['0009', '000B']);
