-- PostgreSQL-compatible schema for Sociomile
-- Creates necessary tables for messages, conversations, users, customers, tickets, events

-- Create extension for UUID generation (optional)
DO $$ BEGIN
    CREATE EXTENSION IF NOT EXISTS "pgcrypto";
EXCEPTION WHEN duplicate_object THEN
    -- already exists
END $$;

-- Users
CREATE TABLE IF NOT EXISTS users (
  id VARCHAR(36) PRIMARY KEY,
  tenant_id VARCHAR(36) NOT NULL,
  email VARCHAR(255) NOT NULL UNIQUE,
  password VARCHAR(255) NOT NULL,
  name VARCHAR(255) NOT NULL,
  role VARCHAR(20) NOT NULL DEFAULT 'agent',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_users_tenant_id ON users(tenant_id);

-- Customers
CREATE TABLE IF NOT EXISTS customers (
  id VARCHAR(36) PRIMARY KEY,
  tenant_id VARCHAR(36) NOT NULL,
  external_id VARCHAR(255) NOT NULL,
  name VARCHAR(255) NOT NULL,
  channel VARCHAR(50) NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (tenant_id, external_id)
);
CREATE INDEX IF NOT EXISTS idx_customers_tenant_id ON customers(tenant_id);

-- Channels
CREATE TABLE IF NOT EXISTS channels (
  id VARCHAR(36) PRIMARY KEY,
  tenant_id VARCHAR(36) NOT NULL,
  name VARCHAR(255) NOT NULL,
  slug VARCHAR(255) NOT NULL UNIQUE,
  description TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_channels_tenant_id ON channels(tenant_id);

-- Conversations
CREATE TABLE IF NOT EXISTS conversations (
  id VARCHAR(36) PRIMARY KEY,
  tenant_id VARCHAR(36) NOT NULL,
  customer_id VARCHAR(36) NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
  status VARCHAR(20) NOT NULL DEFAULT 'open',
  assigned_agent_id VARCHAR(36) NULL REFERENCES users(id) ON DELETE SET NULL,
  channel VARCHAR(50) NOT NULL,
  last_message_at TIMESTAMPTZ NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_conversations_tenant_id ON conversations(tenant_id);

-- Messages
CREATE TABLE IF NOT EXISTS messages (
  id VARCHAR(36) PRIMARY KEY,
  conversation_id VARCHAR(36) NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
  sender_type VARCHAR(20) NOT NULL,
  sender_id VARCHAR(36) NOT NULL,
  sender_name VARCHAR(255) NOT NULL,
  message TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_messages_conversation_id ON messages(conversation_id);

-- Tickets
CREATE TABLE IF NOT EXISTS tickets (
  id VARCHAR(36) PRIMARY KEY,
  tenant_id VARCHAR(36) NOT NULL,
  conversation_id VARCHAR(36) NULL REFERENCES conversations(id) ON DELETE CASCADE,
  title VARCHAR(255) NOT NULL,
  description TEXT NOT NULL,
  status VARCHAR(20) NOT NULL DEFAULT 'open',
  priority VARCHAR(20) NOT NULL DEFAULT 'medium',
  assigned_agent_id VARCHAR(36) NULL REFERENCES users(id) ON DELETE SET NULL,
  created_by_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Events
CREATE TABLE IF NOT EXISTS events (
  id VARCHAR(36) PRIMARY KEY,
  tenant_id VARCHAR(36) NOT NULL,
  event_type VARCHAR(100) NOT NULL,
  entity_type VARCHAR(50) NOT NULL,
  entity_id VARCHAR(36) NOT NULL,
  data JSONB,
  user_id VARCHAR(36),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- After both conversations and tickets exist, add selected_ticket_id to conversations
ALTER TABLE conversations ADD COLUMN IF NOT EXISTS selected_ticket_id VARCHAR(36) NULL REFERENCES tickets(id) ON DELETE SET NULL;

-- Join table to allow many-to-many relation between conversations and tickets
CREATE TABLE IF NOT EXISTS conversation_tickets (
  conversation_id VARCHAR(36) NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
  ticket_id VARCHAR(36) NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (conversation_id, ticket_id)
);

-- Add human-friendly ticket code column and unique index
ALTER TABLE tickets ADD COLUMN IF NOT EXISTS code VARCHAR(50) NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_tickets_code ON tickets(code);


-- Insert a local admin user safely using pgcrypto's crypt()
INSERT INTO users (id, tenant_id, email, password, name, role)
SELECT 'user_local_admin', 'tenant_001', 'localadmin@sociomile.com', crypt('password', gen_salt('bf', 10)), 'Local Admin', 'admin'
WHERE NOT EXISTS (SELECT 1 FROM users WHERE email='localadmin@sociomile.com');
-- Insert a local agent user safely using pgcrypto's crypt()
INSERT INTO users (id, tenant_id, email, password, name, role)
SELECT 'user_local_agent', 'tenant_001', 'localagent@sociomile.com', crypt('agent123', gen_salt('bf', 10)), 'Local Agent', 'agent'
WHERE NOT EXISTS (SELECT 1 FROM users WHERE email='localagent@sociomile.com');
