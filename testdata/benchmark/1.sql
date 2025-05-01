CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE IF NOT EXISTS users (
  user_id uuid PRIMARY KEY,
  email text NOT NULL,
  first_name text NOT NULL,
  last_name text NOT NULL,

  created_at timestamp with time zone DEFAULT clock_timestamp() NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS users_email_key ON users (lower(email));

GRANT INSERT, UPDATE, SELECT, DELETE ON users TO copydog;

CREATE TABLE IF NOT EXISTS documents (
  document_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name text NOT NULL,
  author_id uuid references users(user_id) NOT NULL,
  content text NOT NULL,
  is_published boolean NOT NULL,

  created_at timestamp with time zone DEFAULT clock_timestamp() NOT NULL
);

CREATE INDEX IF NOT EXISTS documents_author_id_idx ON documents (author_id);
CREATE INDEX IF NOT EXISTS documents_content_trgm_idx ON documents USING GIST (content gist_trgm_ops);

CREATE TABLE IF NOT EXISTS document_users (
  document_id uuid references documents(document_id) NOT NULL,
  user_id uuid references users(user_id) NOT NULL
);

CREATE INDEX IF NOT EXISTS document_users_document_id_idx ON document_users (document_id);
CREATE INDEX IF NOT EXISTS document_users_user_id_idx ON document_users (user_id);
CREATE UNIQUE INDEX IF NOT EXISTS document_users_document_id_user_id_key ON document_users (document_id, user_id);

GRANT INSERT, UPDATE, SELECT, DELETE ON documents TO copydog;
GRANT INSERT, UPDATE, SELECT, DELETE ON document_users TO copydog;

CREATE TABLE IF NOT EXISTS collections (
  collection_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name text NOT NULL,
	description text NOT NULL,
  owner_id uuid REFERENCES users(user_id) NOT NULL,

  created_at timestamp with time zone DEFAULT clock_timestamp() NOT NULL
);

CREATE INDEX IF NOT EXISTS collections_owner_id_idx ON collections (owner_id);
CREATE UNIQUE INDEX IF NOT EXISTS collections_name_owner_id_key ON collections (name, owner_id);

GRANT INSERT, UPDATE, SELECT, DELETE ON collections TO copydog;

CREATE TABLE IF NOT EXISTS collection_documents (
	collection_id uuid REFERENCES collections(collection_id) ON DELETE CASCADE NOT NULL,
	document_id uuid REFERENCES documents(document_id) ON DELETE CASCADE NOT NULL
);

CREATE INDEX IF NOT EXISTS collection_documents_collection_id_idx ON collection_documents (collection_id);
CREATE UNIQUE INDEX IF NOT EXISTS collection_documents_document_id_key ON collection_documents (document_id);

GRANT INSERT, UPDATE, SELECT, DELETE ON collection_documents TO copydog;

CREATE TABLE IF NOT EXISTS user_collection_order (
	collection_id uuid REFERENCES collections(collection_id) ON DELETE CASCADE NOT NULL,
	user_id uuid REFERENCES users(user_id) ON DELETE CASCADE NOT NULL,
	rank integer NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS user_collection_order_key ON user_collection_order(collection_id, user_id);
GRANT INSERT, UPDATE, SELECT, DELETE ON user_collection_order TO copydog;

CREATE OR REPLACE FUNCTION updated_at_timestamp() RETURNS TRIGGER 
LANGUAGE plpgsql
AS $$
BEGIN
    IF (NEW != OLD) THEN
        NEW.updated_at = clock_timestamp();
        RETURN NEW;
    END IF;
    RETURN OLD;
END;
$$;
-- +goose StatementEnd

ALTER TABLE users ADD COLUMN updated_at timestamp with time zone DEFAULT clock_timestamp() NOT NULL;
CREATE TRIGGER users_updated_at
	BEFORE UPDATE
	ON users
	FOR EACH ROW
	EXECUTE PROCEDURE updated_at_timestamp();

ALTER TABLE documents ADD COLUMN updated_at timestamp with time zone DEFAULT clock_timestamp() NOT NULL;
CREATE TRIGGER documents_updated_at
	BEFORE UPDATE
	ON documents
	FOR EACH ROW
	EXECUTE PROCEDURE updated_at_timestamp();

ALTER TABLE collections ADD COLUMN updated_at timestamp with time zone DEFAULT clock_timestamp() NOT NULL;
CREATE TRIGGER collections_updated_at
	BEFORE UPDATE
	ON collections
	FOR EACH ROW
	EXECUTE PROCEDURE updated_at_timestamp();

CREATE TABLE IF NOT EXISTS instant_document_links (
	instant_document_link_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	document_id uuid references documents(document_id) ON DELETE CASCADE NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS instant_document_links_document_id_key ON instant_document_links (document_id);

GRANT INSERT, UPDATE, SELECT, DELETE ON instant_document_links TO copydog;

CREATE TABLE IF NOT EXISTS user_settings (
  user_id uuid references users(user_id) ON DELETE CASCADE NOT NULL,
  tab_size integer NOT NULL DEFAULT 4
);

CREATE UNIQUE INDEX IF NOT EXISTS user_settings_user_id ON user_settings (user_id);

INSERT INTO user_settings (user_id) SELECT user_id FROM users;

GRANT INSERT, UPDATE, SELECT on user_settings TO copydog;

CREATE TABLE IF NOT EXISTS collection_roles (
  role_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name text NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS collection_roles_name_key ON collection_roles(lower(name));

INSERT INTO collection_roles (role_id, name) VALUES ('c949fbe9-fec8-4d5e-95fc-a9db5f6de5bf', 'collaborator');
GRANT SELECT ON collection_roles TO copydog;

CREATE TABLE IF NOT EXISTS collection_collaborators (
  collection_id uuid references collections(collection_id) ON DELETE CASCADE NOT NULL,
  user_id uuid references users(user_id) ON DELETE CASCADE NOT NULL,
  role_id uuid references collection_roles(role_id) NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS collection_collaborators_collection_id_user_id_idx ON collection_collaborators(collection_id, user_id);
CREATE INDEX IF NOT EXISTS collection_collaborators_role_id ON collection_collaborators(role_id);

GRANT INSERT, SELECT, UPDATE, DELETE ON collection_collaborators TO copydog;

CREATE TABLE IF NOT EXISTS document_revisions (
  document_id uuid references documents(document_id) ON DELETE CASCADE NOT NULL,
  version int NOT NULL,
  content text NOT NULL,
  created_at timestamp with time zone NOT NULL,

  PRIMARY KEY(document_id, version)
);

GRANT INSERT, UPDATE, SELECT ON document_revisions TO copydog;
