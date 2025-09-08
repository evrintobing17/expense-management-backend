-- Create the database user if it doesn't exist
DO $$
BEGIN
  IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'expense_user') THEN
    CREATE ROLE expense_user WITH LOGIN PASSWORD 'expense_password';
  END IF;
END
$$;

-- Create the database if it doesn't exist
SELECT 'CREATE DATABASE expense_db WITH OWNER expense_user'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'expense_db')\gexec

-- Grant privileges to the user
GRANT ALL PRIVILEGES ON DATABASE expense_db TO expense_user;