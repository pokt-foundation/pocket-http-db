-- deactivate F-chain chains
UPDATE chains
SET active = FALSE
WHERE id IN (
    'F000', 'F001', 'F002', 'F003', 'F004', 'F005', 'F006', 'F007', 'F008', 'F009',
    'F00A', 'F00B', 'F00C', 'F00D', 'F00E', 'F00F', 'F010', 'F011', 'F012', 'F013',
    'F014', 'F015', 'F016', 'F017', 'F018', 'F019', 'F01A', 'F01B', 'F01C', 'F01D',
    'F01E', 'F01F', 'F020', 'F021', 'F022', 'F023', 'F024', 'F025', 'F026', 'F027',
    'F028', 'F029', 'F02A', 'F02B'
);