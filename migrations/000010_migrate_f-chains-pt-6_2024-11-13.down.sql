-- F005
DELETE FROM chain_alias_domains WHERE chain_id = 'F005';
UPDATE chains SET active = true WHERE id = '0079';

-- F006
DELETE FROM chain_alias_domains WHERE chain_id = 'F006';
UPDATE chains SET active = true WHERE id = '0080';

-- F00A
DELETE FROM chain_alias_domains WHERE chain_id = 'F00A';
UPDATE chains SET active = true WHERE id = '0048';

-- F00B
DELETE FROM chain_alias_domains WHERE chain_id = 'F00B';
UPDATE chains SET active = true WHERE id = '0065';

-- F010
DELETE FROM chain_alias_domains WHERE chain_id = 'F010';
UPDATE chains SET active = true WHERE id = '0049';

-- F017
DELETE FROM chain_alias_domains WHERE chain_id = 'F017';
UPDATE chains SET active = true WHERE id = ANY(ARRAY['0071', '0072']);

-- F018
DELETE FROM chain_alias_domains WHERE chain_id = 'F018';
UPDATE chains SET active = true WHERE id = '0058';

-- F01C
DELETE FROM chain_alias_domains WHERE chain_id = 'F01C';
UPDATE chains SET active = true WHERE id = ANY(ARRAY['0069', '0070']);

-- F025
DELETE FROM chain_alias_domains WHERE chain_id = 'F025';
UPDATE chains SET active = true WHERE id = ANY(ARRAY['0006', 'C006']);