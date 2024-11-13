-- F001
DELETE FROM chain_alias_domains WHERE chain_id = 'F001';
UPDATE chains SET active = true WHERE id = '0066';

-- F00E
DELETE FROM chain_alias_domains WHERE chain_id = 'F00E';
UPDATE chains SET active = true WHERE id = ANY(ARRAY['0077', '0078']);

-- F00F
DELETE FROM chain_alias_domains WHERE chain_id = 'F00F';
UPDATE chains SET active = true WHERE id = '0046';

-- F011
DELETE FROM chain_alias_domains WHERE chain_id = 'F011';
UPDATE chains SET active = true WHERE id = 'A088';

-- F012
DELETE FROM chain_alias_domains WHERE chain_id = 'F012';
UPDATE chains SET active = true WHERE id = '0005';

-- F016
DELETE FROM chain_alias_domains WHERE chain_id = 'F016';
UPDATE chains SET active = true WHERE id = '0056';

-- F01A
DELETE FROM chain_alias_domains WHERE chain_id = 'F01A';
UPDATE chains SET active = true WHERE id = '0051';

-- F021
DELETE FROM chain_alias_domains WHERE chain_id = 'F021';
UPDATE chains SET active = true WHERE id = ANY(ARRAY['0009', '000B']);

-- F026
DELETE FROM chain_alias_domains WHERE chain_id = 'F026';
UPDATE chains SET active = true WHERE id = '0076';

-- F029
DELETE FROM chain_alias_domains WHERE chain_id = 'F029';
UPDATE chains SET active = true WHERE id = '0074';