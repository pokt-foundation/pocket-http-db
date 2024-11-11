-- Down Migration: Remove all newly added F-chain rows to revert to previous state
-- Ensures the database is restored to its original state before the new chains were added

-- Remove chain checks
DELETE FROM chain_checks WHERE chain_id IN (
    'F001', 'F002', 'F003', 'F005', 'F006', 'F008', 'F009', 'F00A', 'F00B', 'F00C',
    'F00D', 'F00E', 'F00F', 'F010', 'F011', 'F012', 'F013', 'F014', 'F015', 'F016',
    'F017', 'F018', 'F019', 'F01A', 'F01B', 'F01C', 'F01D', 'F01E', 'F01F', 'F020',
    'F021', 'F022', 'F024', 'F027', 'F028', 'F029', 'F02A', 'F02B'
);

-- Remove sync checks
DELETE FROM chain_checks WHERE chain_id IN (
    'F001', 'F002', 'F003', 'F004', 'F005', 'F006', 'F007', 'F008', 'F009', 'F00A',
    'F00B', 'F00C', 'F00D', 'F00E', 'F00F', 'F010', 'F011', 'F012', 'F013', 'F014',
    'F015', 'F016', 'F017', 'F018', 'F019', 'F01A', 'F01B', 'F01C', 'F01D', 'F01E',
    'F01F', 'F020', 'F021', 'F022', 'F024', 'F025', 'F026', 'F027', 'F028', 'F029',
    'F02A', 'F02B'
);

-- Remove chain alias domains
DELETE FROM chain_alias_domains WHERE chain_id IN (
    'F000', 'F001', 'F002', 'F003', 'F004', 'F005', 'F006', 'F007', 'F008', 'F009',
    'F00A', 'F00B', 'F00C', 'F00D', 'F00E', 'F00F', 'F010', 'F011', 'F012', 'F013',
    'F014', 'F015', 'F016', 'F017', 'F018', 'F019', 'F01A', 'F01B', 'F01C', 'F01D',
    'F01E', 'F01F', 'F020', 'F021', 'F022', 'F023', 'F024', 'F025', 'F026', 'F027',
    'F028', 'F029', 'F02A', 'F02B'
);

-- Remove chain config
DELETE FROM chain_config WHERE chain_id IN (
    'F000', 'F001', 'F002', 'F003', 'F004', 'F005', 'F006', 'F007', 'F008', 'F009',
    'F00A', 'F00B', 'F00C', 'F00D', 'F00E', 'F00F', 'F010', 'F011', 'F012', 'F013',
    'F014', 'F015', 'F016', 'F017', 'F018', 'F019', 'F01A', 'F01B', 'F01C', 'F01D',
    'F01E', 'F01F', 'F020', 'F021', 'F022', 'F023', 'F024', 'F025', 'F026', 'F027',
    'F028', 'F029', 'F02A', 'F02B'
);

-- Remove chains
DELETE FROM chains WHERE id IN (
    'F000', 'F001', 'F002', 'F003', 'F004', 'F005', 'F006', 'F007', 'F008', 'F009',
    'F00A', 'F00B', 'F00C', 'F00D', 'F00E', 'F00F', 'F010', 'F011', 'F012', 'F013',
    'F014', 'F015', 'F016', 'F017', 'F018', 'F019', 'F01A', 'F01B', 'F01C', 'F01D',
    'F01E', 'F01F', 'F020', 'F021', 'F022', 'F023', 'F024', 'F025', 'F026', 'F027',
    'F028', 'F029', 'F02A', 'F02B'
);
