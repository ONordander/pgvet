CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE IF NOT EXISTS users (
  user_id uuid PRIMARY KEY,
  email text NOT NULL,
  first_name text NOT NULL,
  last_name text NOT NULL,

  created_at timestamp with time zone DEFAULT clock_timestamp() NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS users_email_key ON users (lower(email));

GRANT INSERT, UPDATE, SELECT, DELETE ON users TO lorem;

CREATE TABLE IF NOT EXISTS stones (
  stone_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name text NOT NULL,
  author_id uuid references users(user_id) NOT NULL,
  content text NOT NULL,
  is_published boolean NOT NULL,

  created_at timestamp with time zone DEFAULT clock_timestamp() NOT NULL
);

CREATE INDEX IF NOT EXISTS stones_author_id_idx ON stones (author_id);
CREATE INDEX IF NOT EXISTS stones_content_trgm_idx ON stones USING GIST (content gist_trgm_ops);

CREATE TABLE IF NOT EXISTS stone_users (
  stone_id uuid references stones(stone_id) NOT NULL,
  user_id uuid references users(user_id) NOT NULL
);

CREATE INDEX IF NOT EXISTS stone_users_stone_id_idx ON stone_users (stone_id);
CREATE INDEX IF NOT EXISTS stone_users_user_id_idx ON stone_users (user_id);
CREATE UNIQUE INDEX IF NOT EXISTS stone_users_stone_id_user_id_key ON stone_users (stone_id, user_id);

GRANT INSERT, UPDATE, SELECT, DELETE ON stones TO lorem;
GRANT INSERT, UPDATE, SELECT, DELETE ON stone_users TO lorem;

CREATE TABLE IF NOT EXISTS groups (
  group_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name text NOT NULL,
	description text NOT NULL,
  owner_id uuid REFERENCES users(user_id) NOT NULL,

  created_at timestamp with time zone DEFAULT clock_timestamp() NOT NULL
);

CREATE INDEX IF NOT EXISTS groups_owner_id_idx ON groups (owner_id);
CREATE UNIQUE INDEX IF NOT EXISTS groups_name_owner_id_key ON groups (name, owner_id);

GRANT INSERT, UPDATE, SELECT, DELETE ON groups TO lorem;

CREATE TABLE IF NOT EXISTS group_stones (
	group_id uuid REFERENCES groups(group_id) ON DELETE CASCADE NOT NULL,
	stone_id uuid REFERENCES stones(stone_id) ON DELETE CASCADE NOT NULL
);

CREATE INDEX IF NOT EXISTS group_stones_group_id_idx ON group_stones (group_id);
CREATE UNIQUE INDEX IF NOT EXISTS group_stones_stone_id_key ON group_stones (stone_id);

GRANT INSERT, UPDATE, SELECT, DELETE ON group_stones TO lorem;

CREATE TABLE IF NOT EXISTS user_group_order (
	group_id uuid REFERENCES groups(group_id) ON DELETE CASCADE NOT NULL,
	user_id uuid REFERENCES users(user_id) ON DELETE CASCADE NOT NULL,
	rank integer NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS user_group_order_key ON user_group_order(group_id, user_id);
GRANT INSERT, UPDATE, SELECT, DELETE ON user_group_order TO lorem;

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

ALTER TABLE stones ADD COLUMN updated_at timestamp with time zone DEFAULT clock_timestamp() NOT NULL;
CREATE TRIGGER stones_updated_at
	BEFORE UPDATE
	ON stones
	FOR EACH ROW
	EXECUTE PROCEDURE updated_at_timestamp();

ALTER TABLE groups ADD COLUMN updated_at timestamp with time zone DEFAULT clock_timestamp() NOT NULL;
CREATE TRIGGER groups_updated_at
	BEFORE UPDATE
	ON groups
	FOR EACH ROW
	EXECUTE PROCEDURE updated_at_timestamp();

CREATE TABLE IF NOT EXISTS instant_stone_links (
	instant_stone_link_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	stone_id uuid references stones(stone_id) ON DELETE CASCADE NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS instant_stone_links_stone_id_key ON instant_stone_links (stone_id);

GRANT INSERT, UPDATE, SELECT, DELETE ON instant_stone_links TO lorem;

CREATE TABLE IF NOT EXISTS user_settings (
  user_id uuid references users(user_id) ON DELETE CASCADE NOT NULL,
  tab_size integer NOT NULL DEFAULT 4
);

CREATE UNIQUE INDEX IF NOT EXISTS user_settings_user_id ON user_settings (user_id);

INSERT INTO user_settings (user_id) SELECT user_id FROM users;

GRANT INSERT, UPDATE, SELECT on user_settings TO lorem;

CREATE TABLE IF NOT EXISTS group_roles (
  role_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name text NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS group_roles_name_key ON group_roles(lower(name));

INSERT INTO group_roles (role_id, name) VALUES ('c949fbe9-fec8-4d5e-95fc-a9db5f6de5bf', 'collaborator');
GRANT SELECT ON group_roles TO lorem;

CREATE TABLE IF NOT EXISTS group_collaborators (
  group_id uuid references groups(group_id) ON DELETE CASCADE NOT NULL,
  user_id uuid references users(user_id) ON DELETE CASCADE NOT NULL,
  role_id uuid references group_roles(role_id) NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS group_collaborators_group_id_user_id_idx ON group_collaborators(group_id, user_id);
CREATE INDEX IF NOT EXISTS group_collaborators_role_id ON group_collaborators(role_id);

GRANT INSERT, SELECT, UPDATE, DELETE ON group_collaborators TO lorem;

CREATE TABLE IF NOT EXISTS stone_revisions (
  stone_id uuid references stones(stone_id) ON DELETE CASCADE NOT NULL,
  version int NOT NULL,
  content text NOT NULL,
  created_at timestamp with time zone NOT NULL,

  PRIMARY KEY(stone_id, version)
);

GRANT INSERT, UPDATE, SELECT ON stone_revisions TO lorem;
