#!/bin/bash
set -e

# create pgBouncer user
echo "CREATE USER pgbouncer;" | psql --username postgres --dbname postgres

echo "CREATE SCHEMA pgbouncer AUTHORIZATION pgbouncer;" |
    psql --username postgres --dbname postgres

# create pgBouncer auth function
psql --username postgres --dbname postgres <<'EOSQL'
CREATE OR REPLACE FUNCTION pgbouncer.get_auth(p_usename TEXT)
RETURNS TABLE(username TEXT, password TEXT) AS
$$
BEGIN
    RAISE WARNING 'pgBouncer auth request: %', p_usename;
 
    RETURN QUERY
    SELECT usename::TEXT, passwd::TEXT
    FROM pg_catalog.pg_shadow
    WHERE usename = p_usename
    AND NOT usesuper;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;
EOSQL

echo "REVOKE ALL ON FUNCTION pgbouncer.get_auth(p_usename TEXT) FROM PUBLIC;" |
    psql --username postgres --dbname postgres

echo "GRANT EXECUTE ON FUNCTION pgbouncer.get_auth(p_usename TEXT) TO pgbouncer;" |
    psql --username postgres --dbname postgres
