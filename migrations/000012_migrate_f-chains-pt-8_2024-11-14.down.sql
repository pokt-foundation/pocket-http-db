-- F00E
DELETE FROM chain_alias_domains WHERE chain_id = 'F00E';
UPDATE chains SET active = true WHERE id = ANY(ARRAY['0077', '0078']);

-- F012
DELETE FROM chain_alias_domains WHERE chain_id = 'F012';
UPDATE chains SET active = true WHERE id = ANY(ARRAY['0005', '000A']);

-- F026
DELETE FROM chain_alias_domains WHERE chain_id = 'F026';
UPDATE chains SET active = true WHERE id = '0076';

-- F029
DELETE FROM chain_alias_domains WHERE chain_id = 'F029';
UPDATE chains SET active = true WHERE id = '0074';