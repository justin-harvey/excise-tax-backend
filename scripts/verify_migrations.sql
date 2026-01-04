-- =====================================================
-- Migration Verification Script
-- Run this after applying all migrations to verify schema
-- =====================================================

\echo '==========================================';
\echo 'Excise Tax Portal - Migration Verification';
\echo '==========================================';
\echo '';

-- =====================================================
-- 1. Check Migration Version
-- =====================================================
\echo '1. Migration Version:';
\echo '--------------------';
SELECT version, dirty FROM schema_migrations;
\echo '';

-- =====================================================
-- 2. Count Database Objects
-- =====================================================
\echo '2. Database Object Counts:';
\echo '-------------------------';

-- Count tables
SELECT 'Tables:' as object_type, COUNT(*) as count
FROM information_schema.tables
WHERE table_schema = 'public' AND table_type = 'BASE TABLE'
UNION ALL
-- Count views
SELECT 'Views:', COUNT(*)
FROM information_schema.views
WHERE table_schema = 'public'
UNION ALL
-- Count functions
SELECT 'Functions:', COUNT(*)
FROM pg_proc p
JOIN pg_namespace n ON p.pronamespace = n.oid
WHERE n.nspname = 'public' AND p.prokind = 'f'
UNION ALL
-- Count triggers
SELECT 'Triggers:', COUNT(*)
FROM pg_trigger t
JOIN pg_class c ON t.tgrelid = c.oid
JOIN pg_namespace n ON c.relnamespace = n.oid
WHERE n.nspname = 'public' AND NOT t.tgisinternal;

\echo '';

-- =====================================================
-- 3. List All Tables
-- =====================================================
\echo '3. Tables (should be 11):';
\echo '------------------------';
SELECT
    schemaname,
    tablename,
    tableowner
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY tablename;
\echo '';

-- =====================================================
-- 4. List All Views
-- =====================================================
\echo '4. Views (should be 7):';
\echo '----------------------';
SELECT
    schemaname,
    viewname,
    viewowner
FROM pg_views
WHERE schemaname = 'public'
ORDER BY viewname;
\echo '';

-- =====================================================
-- 5. Verify Table Structure
-- =====================================================
\echo '5. Table Column Counts:';
\echo '----------------------';
SELECT
    table_name,
    COUNT(*) as column_count
FROM information_schema.columns
WHERE table_schema = 'public'
GROUP BY table_name
ORDER BY table_name;
\echo '';

-- =====================================================
-- 6. Check Foreign Keys
-- =====================================================
\echo '6. Foreign Key Constraints:';
\echo '--------------------------';
SELECT
    tc.table_name,
    kcu.column_name,
    ccu.table_name AS foreign_table_name,
    ccu.column_name AS foreign_column_name,
    rc.delete_rule
FROM information_schema.table_constraints AS tc
JOIN information_schema.key_column_usage AS kcu
    ON tc.constraint_name = kcu.constraint_name
    AND tc.table_schema = kcu.table_schema
JOIN information_schema.constraint_column_usage AS ccu
    ON ccu.constraint_name = tc.constraint_name
    AND ccu.table_schema = tc.table_schema
JOIN information_schema.referential_constraints AS rc
    ON rc.constraint_name = tc.constraint_name
WHERE tc.constraint_type = 'FOREIGN KEY'
    AND tc.table_schema = 'public'
ORDER BY tc.table_name, kcu.column_name;
\echo '';

-- =====================================================
-- 7. Check Indexes
-- =====================================================
\echo '7. Index Count by Table:';
\echo '------------------------';
SELECT
    tablename,
    COUNT(*) as index_count
FROM pg_indexes
WHERE schemaname = 'public'
GROUP BY tablename
ORDER BY tablename;
\echo '';

-- =====================================================
-- 8. Verify Critical Indexes
-- =====================================================
\echo '8. Critical Indexes Check:';
\echo '-------------------------';
SELECT
    tablename,
    indexname,
    indexdef
FROM pg_indexes
WHERE schemaname = 'public'
    AND (
        indexname LIKE '%_email%'
        OR indexname LIKE '%_tx_hash%'
        OR indexname LIKE '%_status%'
        OR indexname LIKE '%_user_id%'
        OR indexname LIKE '%_manufacturer_id%'
    )
ORDER BY tablename, indexname;
\echo '';

-- =====================================================
-- 9. Check Triggers
-- =====================================================
\echo '9. Triggers:';
\echo '-----------';
SELECT
    tgname as trigger_name,
    relname as table_name,
    proname as function_name
FROM pg_trigger t
JOIN pg_class c ON t.tgrelid = c.oid
JOIN pg_proc p ON t.tgfoid = p.oid
WHERE NOT t.tgisinternal
    AND c.relnamespace = (SELECT oid FROM pg_namespace WHERE nspname = 'public')
ORDER BY relname, tgname;
\echo '';

-- =====================================================
-- 10. Check Functions
-- =====================================================
\echo '10. Database Functions:';
\echo '----------------------';
SELECT
    p.proname as function_name,
    pg_get_function_arguments(p.oid) as arguments,
    pg_get_function_result(p.oid) as return_type
FROM pg_proc p
JOIN pg_namespace n ON p.pronamespace = n.oid
WHERE n.nspname = 'public' AND p.prokind = 'f'
ORDER BY p.proname;
\echo '';

-- =====================================================
-- 11. Check Check Constraints
-- =====================================================
\echo '11. Check Constraints:';
\echo '---------------------';
SELECT
    tc.table_name,
    tc.constraint_name,
    cc.check_clause
FROM information_schema.table_constraints tc
JOIN information_schema.check_constraints cc
    ON tc.constraint_name = cc.constraint_name
WHERE tc.table_schema = 'public'
    AND tc.constraint_type = 'CHECK'
ORDER BY tc.table_name, tc.constraint_name;
\echo '';

-- =====================================================
-- 12. Check Unique Constraints
-- =====================================================
\echo '12. Unique Constraints:';
\echo '----------------------';
SELECT
    tc.table_name,
    kcu.column_name,
    tc.constraint_name
FROM information_schema.table_constraints tc
JOIN information_schema.key_column_usage kcu
    ON tc.constraint_name = kcu.constraint_name
WHERE tc.table_schema = 'public'
    AND tc.constraint_type = 'UNIQUE'
ORDER BY tc.table_name, kcu.column_name;
\echo '';

-- =====================================================
-- 13. Verify Default Values
-- =====================================================
\echo '13. Columns with Default Values:';
\echo '-------------------------------';
SELECT
    table_name,
    column_name,
    column_default
FROM information_schema.columns
WHERE table_schema = 'public'
    AND column_default IS NOT NULL
ORDER BY table_name, ordinal_position;
\echo '';

-- =====================================================
-- 14. Check Table Comments
-- =====================================================
\echo '14. Table Comments:';
\echo '------------------';
SELECT
    c.relname as table_name,
    pg_catalog.obj_description(c.oid, 'pg_class') as comment
FROM pg_catalog.pg_class c
JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace
WHERE n.nspname = 'public'
    AND c.relkind = 'r'
    AND pg_catalog.obj_description(c.oid, 'pg_class') IS NOT NULL
ORDER BY c.relname;
\echo '';

-- =====================================================
-- 15. Database Size
-- =====================================================
\echo '15. Database Size:';
\echo '-----------------';
SELECT
    pg_size_pretty(pg_database_size(current_database())) as database_size;
\echo '';

-- =====================================================
-- 16. Table Sizes
-- =====================================================
\echo '16. Table Sizes:';
\echo '---------------';
SELECT
    schemaname,
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;
\echo '';

-- =====================================================
-- 17. Expected vs Actual Validation
-- =====================================================
\echo '17. Validation Summary:';
\echo '----------------------';

WITH expected AS (
    SELECT 11 as tables, 7 as views, 4 as functions, 5 as triggers
),
actual AS (
    SELECT
        (SELECT COUNT(*) FROM information_schema.tables
         WHERE table_schema = 'public' AND table_type = 'BASE TABLE') as tables,
        (SELECT COUNT(*) FROM information_schema.views
         WHERE table_schema = 'public') as views,
        (SELECT COUNT(*) FROM pg_proc p
         JOIN pg_namespace n ON p.pronamespace = n.oid
         WHERE n.nspname = 'public' AND p.prokind = 'f') as functions,
        (SELECT COUNT(*) FROM pg_trigger t
         JOIN pg_class c ON t.tgrelid = c.oid
         JOIN pg_namespace n ON c.relnamespace = n.oid
         WHERE n.nspname = 'public' AND NOT t.tgisinternal) as triggers
)
SELECT
    'Tables' as object_type,
    e.tables as expected,
    a.tables as actual,
    CASE WHEN e.tables = a.tables THEN '✓ PASS' ELSE '✗ FAIL' END as status
FROM expected e, actual a
UNION ALL
SELECT
    'Views',
    e.views,
    a.views,
    CASE WHEN e.views = a.views THEN '✓ PASS' ELSE '✗ FAIL' END
FROM expected e, actual a
UNION ALL
SELECT
    'Functions',
    e.functions,
    a.functions,
    CASE WHEN e.functions = a.functions THEN '✓ PASS' ELSE '✗ FAIL' END
FROM expected e, actual a
UNION ALL
SELECT
    'Triggers',
    e.triggers,
    a.triggers,
    CASE WHEN e.triggers = a.triggers THEN '✓ PASS' ELSE '✗ FAIL' END
FROM expected e, actual a;

\echo '';

-- =====================================================
-- 18. Critical Tables Check
-- =====================================================
\echo '18. Critical Tables Existence:';
\echo '------------------------------';

WITH critical_tables AS (
    SELECT unnest(ARRAY[
        'users', 'manufacturers', 'audit_log', 'sessions',
        'payments', 'xrpl_payments', 'exchange_rates',
        'tax_reports', 'tax_rates'
    ]) as table_name
)
SELECT
    ct.table_name,
    CASE
        WHEN t.table_name IS NOT NULL THEN '✓ EXISTS'
        ELSE '✗ MISSING'
    END as status
FROM critical_tables ct
LEFT JOIN information_schema.tables t
    ON ct.table_name = t.table_name
    AND t.table_schema = 'public';

\echo '';

-- =====================================================
-- 19. Critical Views Check
-- =====================================================
\echo '19. Critical Views Existence:';
\echo '-----------------------------';

WITH critical_views AS (
    SELECT unnest(ARRAY[
        'dashboard_metrics', 'payment_summary',
        'xrpl_payment_stats', 'manufacturer_report_summary'
    ]) as view_name
)
SELECT
    cv.view_name,
    CASE
        WHEN v.table_name IS NOT NULL THEN '✓ EXISTS'
        ELSE '✗ MISSING'
    END as status
FROM critical_views cv
LEFT JOIN information_schema.views v
    ON cv.view_name = v.table_name
    AND v.table_schema = 'public';

\echo '';

-- =====================================================
-- 20. Final Summary
-- =====================================================
\echo '20. Final Summary:';
\echo '-----------------';

SELECT
    'Migration Version' as check_item,
    (SELECT CAST(version AS TEXT) FROM schema_migrations) as value,
    CASE
        WHEN (SELECT version FROM schema_migrations) = 4 THEN '✓ PASS'
        ELSE '✗ FAIL'
    END as status
UNION ALL
SELECT
    'Dirty State',
    CASE WHEN (SELECT dirty FROM schema_migrations) THEN 'YES' ELSE 'NO' END,
    CASE
        WHEN NOT (SELECT dirty FROM schema_migrations) THEN '✓ PASS'
        ELSE '✗ FAIL'
    END
UNION ALL
SELECT
    'All Tables Created',
    CAST((SELECT COUNT(*) FROM information_schema.tables
          WHERE table_schema = 'public' AND table_type = 'BASE TABLE') AS TEXT) || '/11',
    CASE
        WHEN (SELECT COUNT(*) FROM information_schema.tables
              WHERE table_schema = 'public' AND table_type = 'BASE TABLE') = 11 THEN '✓ PASS'
        ELSE '✗ FAIL'
    END
UNION ALL
SELECT
    'All Views Created',
    CAST((SELECT COUNT(*) FROM information_schema.views
          WHERE table_schema = 'public') AS TEXT) || '/7',
    CASE
        WHEN (SELECT COUNT(*) FROM information_schema.views
              WHERE table_schema = 'public') = 7 THEN '✓ PASS'
        ELSE '✗ FAIL'
    END;

\echo '';
\echo '==========================================';
\echo 'Verification Complete';
\echo '==========================================';
\echo '';
\echo 'Check for any ✗ FAIL status above.';
\echo 'All checks should show ✓ PASS for successful migration.';
\echo '';
