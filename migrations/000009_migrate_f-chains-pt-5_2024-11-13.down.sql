-- This DOWN migration deactivates all the F-chains and reactivates all the old chains.

-- Deactivate all F-chains
WITH f_chain_mapping AS (
    VALUES 
    ('A0CA', ARRAY ['A0CA']),
    ('A0CB', ARRAY ['A0CB']),
    ('A0CC', ARRAY ['A0CC']),
    ('A0CD', ARRAY ['A0CD']),
    ('F000', ARRAY ['0001', 'A001']),
    ('F002', ARRAY ['A086']),
    ('F003', ARRAY ['0003', 'A003']),
    ('F004', ARRAY ['03DF']),
    ('F008', ARRAY ['A090']),
    ('F009', ARRAY ['0004', '0010']),
    ('F00B', ARRAY ['0065']),
    ('F00D', ARRAY ['0081']),
    ('F00F', ARRAY ['0046']),
    ('F013', ARRAY ['0027', '000C']),
    ('F014', ARRAY ['0040']),
    ('F015', ARRAY ['0044']),
    ('F019', ARRAY ['0050']),
    ('F01B', ARRAY ['0052']),
    ('F01D', ARRAY ['0053', 'A053']),
    ('F01E', ARRAY ['A087']),
    ('F01F', ARRAY ['A089']),
    ('F020', ARRAY ['0054']),
    ('F022', ARRAY ['A085']),
    ('F023', ARRAY ['0083']),
    ('F027', ARRAY ['7A00']),
    ('F028', ARRAY ['7A10']),
    ('F029', ARRAY ['0074']),
    ('F02A', ARRAY ['A091']),
    ('F02B', ARRAY ['0084'])
) AS f_chain_mapping(f_chain_id, old_chain_ids)
UPDATE chains
SET active = false
WHERE id IN (
    SELECT f_chain_id
    FROM f_chain_mapping
);

-- Reactivate all old chains
WITH f_chain_mapping AS (
    VALUES 
    ('A0CA', ARRAY ['A0CA']),
    ('A0CB', ARRAY ['A0CB']),
    ('A0CC', ARRAY ['A0CC']),
    ('A0CD', ARRAY ['A0CD']),
    ('F000', ARRAY ['0001', 'A001']),
    ('F002', ARRAY ['A086']),
    ('F003', ARRAY ['0003', 'A003']),
    ('F004', ARRAY ['03DF']),
    ('F008', ARRAY ['A090']),
    ('F009', ARRAY ['0004', '0010']),
    ('F00B', ARRAY ['0065']),
    ('F00D', ARRAY ['0081']),
    ('F00F', ARRAY ['0046']),
    ('F013', ARRAY ['0027', '000C']),
    ('F014', ARRAY ['0040']),
    ('F015', ARRAY ['0044']),
    ('F019', ARRAY ['0050']),
    ('F01B', ARRAY ['0052']),
    ('F01D', ARRAY ['0053', 'A053']),
    ('F01E', ARRAY ['A087']),
    ('F01F', ARRAY ['A089']),
    ('F020', ARRAY ['0054']),
    ('F022', ARRAY ['A085']),
    ('F023', ARRAY ['0083']),
    ('F027', ARRAY ['7A00']),
    ('F028', ARRAY ['7A10']),
    ('F029', ARRAY ['0074']),
    ('F02A', ARRAY ['A091']),
    ('F02B', ARRAY ['0084'])
) AS f_chain_mapping(f_chain_id, old_chain_ids)
UPDATE chains
SET active = true
FROM f_chain_mapping
WHERE chains.id = ANY(f_chain_mapping.old_chain_ids);