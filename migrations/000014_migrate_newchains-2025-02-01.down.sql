-- Update Chains <-> Gigastakes Mapping for February 2025 new chains
BEGIN;
-- Update the chain_id in chains_gigastake_applications based on the mapping
UPDATE chains_gigastake_applications AS cga
SET cga.chain_id = CASE
    WHEN ga.address = '3c50afb94de9d9719f9c23bd2976858c74ad44eb' THEN 'F028'
    WHEN ga.address = '5266fa808ed9a6a5e915894570d46e3d451c1267' THEN 'F022'
    WHEN ga.address = '62e344a58a141935e4f00766bce8f8c1c2736d61' THEN 'F028'
    WHEN ga.address = '6a54693f8169ef5915215abfc3c0f1a44401f688' THEN 'F02B'
    WHEN ga.address = '832db8020cdfdbf65194f2276905950682b69121' THEN 'F000'
    WHEN ga.address = '875d2e01a07c92f32a10e63cec0985fe6a9eb758' THEN 'F01A'
    WHEN ga.address = 'a2e4fee5687a6fa4012f7ba1449a2b4fd3bb32b9' THEN 'F028'
    WHEN ga.address = 'a712d116669653f019b2773c8d4ca47956fa5492' THEN 'F01E'
    WHEN ga.address = 'b728fb333dfdc215aee85c99cebe6386895c743a' THEN 'F000'
    WHEN ga.address = 'b90d686e3c20ca41d795fede44603f17eed04ca3' THEN 'F000'
    WHEN ga.address = 'bd78e8bc24b73fb7977e4db3a781d55e531e73e4' THEN 'F02B'
    WHEN ga.address = 'ccd7e9e95ba790db57a9134124c0e728f42c6e6e' THEN 'F000'
    WHEN ga.address = 'd009929986304b7beb2e90fb560827678c39d056' THEN 'F000'
    WHEN ga.address = 'd5f279c2641e61d9404bb32c070dd3c798cef402' THEN 'F000'
    WHEN ga.address = 'daa46b19faa46e699f07e135bd2211d7ea32345f' THEN 'F000'
    WHEN ga.address = 'dbbf3693f4c2a13460a84c1c848878a727a34a93' THEN 'F000'
    WHEN ga.address = 'e0855387204668b75a7127996f2b52cbaa33786c' THEN 'F000'
    WHEN ga.address = 'ec335e16cf90e5d37ef02123b2fd476802cdad06' THEN 'F00A'
    WHEN ga.address = 'f271965de3d862e63e2258d4618687f30c9cf5b9' THEN 'F000'
    WHEN ga.address = 'f743db5998ba07ae4cbbacb60e62ed194bedd084' THEN 'F000'
    WHEN ga.address = 'f9051e866a8185c0a25e7536ad54f8a12e99d312' THEN 'F000'
    WHEN ga.address = 'f929087a312d0d111118209f814a2bab03179bb6' THEN 'F000'
    WHEN ga.address = 'f950687c021ae45f2df1e5481d22217ea3073d37' THEN 'F000'
    WHEN ga.address = 'fec14897e8427bc025fe4ebc6f7ee6098eac85b8' THEN 'F000'
    ELSE cga.chain_id -- Keep current value if not matched
END
FROM gigastake_applications AS ga
WHERE ga.id = cga.gigastake_application_id
  AND ga.address IN (
    '3c50afb94de9d9719f9c23bd2976858c74ad44eb',
    '5266fa808ed9a6a5e915894570d46e3d451c1267',
    '62e344a58a141935e4f00766bce8f8c1c2736d61',
    '6a54693f8169ef5915215abfc3c0f1a44401f688',
    '832db8020cdfdbf65194f2276905950682b69121',
    '875d2e01a07c92f32a10e63cec0985fe6a9eb758',
    'a2e4fee5687a6fa4012f7ba1449a2b4fd3bb32b9',
    'a712d116669653f019b2773c8d4ca47956fa5492',
    'b728fb333dfdc215aee85c99cebe6386895c743a',
    'b90d686e3c20ca41d795fede44603f17eed04ca3',
    'bd78e8bc24b73fb7977e4db3a781d55e531e73e4',
    'ccd7e9e95ba790db57a9134124c0e728f42c6e6e',
    'd009929986304b7beb2e90fb560827678c39d056',
    'd5f279c2641e61d9404bb32c070dd3c798cef402',
    'daa46b19faa46e699f07e135bd2211d7ea32345f',
    'dbbf3693f4c2a13460a84c1c848878a727a34a93',
    'e0855387204668b75a7127996f2b52cbaa33786c',
    'ec335e16cf90e5d37ef02123b2fd476802cdad06',
    'f271965de3d862e63e2258d4618687f30c9cf5b9',
    'f743db5998ba07ae4cbbacb60e62ed194bedd084',
    'f9051e866a8185c0a25e7536ad54f8a12e99d312',
    'f929087a312d0d111118209f814a2bab03179bb6',
    'f950687c021ae45f2df1e5481d22217ea3073d37',
    'fec14897e8427bc025fe4ebc6f7ee6098eac85b8'
);
COMMIT;
