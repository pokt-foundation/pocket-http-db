-- F000
DELETE FROM chain_alias_domains WHERE chain_id = 'F000';
UPDATE chains SET active = true WHERE id = ANY(ARRAY['0001', 'A001']);

-- F002
DELETE FROM chain_alias_domains WHERE chain_id = 'F002';
UPDATE chains SET active = true WHERE id = 'A086';

-- F003
DELETE FROM chain_alias_domains WHERE chain_id = 'F003';
UPDATE chains SET active = true WHERE id = ANY(ARRAY['0003', 'A003']);

-- F004
DELETE FROM chain_alias_domains WHERE chain_id = 'F004';
UPDATE chains SET active = true WHERE id = '03DF';

-- F008
DELETE FROM chain_alias_domains WHERE chain_id = 'F008';
UPDATE chains SET active = true WHERE id = 'A090';

-- F009
DELETE FROM chain_alias_domains WHERE chain_id = 'F009';
UPDATE chains SET active = true WHERE id = ANY(ARRAY['0004', '0010']);

-- F00B
DELETE FROM chain_alias_domains WHERE chain_id = 'F00B';
UPDATE chains SET active = true WHERE id = '0065';

-- F00D
DELETE FROM chain_alias_domains WHERE chain_id = 'F00D';
UPDATE chains SET active = true WHERE id = '0081';

-- F00F
DELETE FROM chain_alias_domains WHERE chain_id = 'F00F';
UPDATE chains SET active = true WHERE id = '0046';

-- F013
DELETE FROM chain_alias_domains WHERE chain_id = 'F013';
UPDATE chains SET active = true WHERE id = ANY(ARRAY['0027', '000C']);

-- F014
DELETE FROM chain_alias_domains WHERE chain_id = 'F014';
UPDATE chains SET active = true WHERE id = '0040';

-- F015
DELETE FROM chain_alias_domains WHERE chain_id = 'F015';
UPDATE chains SET active = true WHERE id = '0044';

-- F019
DELETE FROM chain_alias_domains WHERE chain_id = 'F019';
UPDATE chains SET active = true WHERE id = '0050';

-- F01B
DELETE FROM chain_alias_domains WHERE chain_id = 'F01B';
UPDATE chains SET active = true WHERE id = '0052';

-- F01D
DELETE FROM chain_alias_domains WHERE chain_id = 'F01D';
UPDATE chains SET active = true WHERE id = ANY(ARRAY['0053', 'A053']);

-- F01E
DELETE FROM chain_alias_domains WHERE chain_id = 'F01E';
UPDATE chains SET active = true WHERE id = 'A087';

-- F01F
DELETE FROM chain_alias_domains WHERE chain_id = 'F01F';
UPDATE chains SET active = true WHERE id = 'A089';

-- F020
DELETE FROM chain_alias_domains WHERE chain_id = 'F020';
UPDATE chains SET active = true WHERE id = '0054';

-- F022
DELETE FROM chain_alias_domains WHERE chain_id = 'F022';
UPDATE chains SET active = true WHERE id = 'A085';

-- F023
DELETE FROM chain_alias_domains WHERE chain_id = 'F023';
UPDATE chains SET active = true WHERE id = '0083';

-- F027
DELETE FROM chain_alias_domains WHERE chain_id = 'F027';
UPDATE chains SET active = true WHERE id = '7A00';

-- F028
DELETE FROM chain_alias_domains WHERE chain_id = 'F028';
UPDATE chains SET active = true WHERE id = '7A10';

-- F02A
DELETE FROM chain_alias_domains WHERE chain_id = 'F02A';
UPDATE chains SET active = true WHERE id = 'A091';

-- F02B
DELETE FROM chain_alias_domains WHERE chain_id = 'F02B';
UPDATE chains SET active = true WHERE id = '0084';

-- F024
DELETE FROM chain_alias_domains WHERE chain_id = 'F024';
UPDATE chains SET active = true WHERE id = '0072';
