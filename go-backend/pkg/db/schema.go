package db

// 与 infra/sql/001_schema.sql 保持一致，供启动时存在性检测与补齐。
const defaultCharset = "utf8mb4"
const defaultCollation = "utf8mb4_unicode_ci"

var schemaTables = []tableSchema{
	{
		name: "users",
		createSQL: `CREATE TABLE IF NOT EXISTS users (
			id BIGINT UNSIGNED NOT NULL,
			username VARCHAR(64) NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			nickname VARCHAR(64) NOT NULL DEFAULT '',
			avatar_url VARCHAR(512) NOT NULL DEFAULT '',
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (id),
			UNIQUE KEY uk_username (username)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		columns: []columnSchema{
			{name: "id", addSQL: "ADD COLUMN id BIGINT UNSIGNED NOT NULL FIRST"},
			{name: "username", addSQL: "ADD COLUMN username VARCHAR(64) NOT NULL"},
			{name: "password_hash", addSQL: "ADD COLUMN password_hash VARCHAR(255) NOT NULL"},
			{name: "nickname", addSQL: "ADD COLUMN nickname VARCHAR(64) NOT NULL DEFAULT ''"},
			{name: "avatar_url", addSQL: "ADD COLUMN avatar_url VARCHAR(512) NOT NULL DEFAULT ''"},
			{name: "created_at", addSQL: "ADD COLUMN created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP"},
		},
		indexes: []indexSchema{
			{name: "uk_username", createSQL: "CREATE UNIQUE INDEX uk_username ON users (username)"},
		},
	},
	{
		name: "friendships",
		createSQL: `CREATE TABLE IF NOT EXISTS friendships (
			id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
			user_id BIGINT UNSIGNED NOT NULL,
			friend_id BIGINT UNSIGNED NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (id),
			UNIQUE KEY uk_pair (user_id, friend_id),
			KEY idx_user (user_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		columns: []columnSchema{
			{name: "id", addSQL: "ADD COLUMN id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT FIRST"},
			{name: "user_id", addSQL: "ADD COLUMN user_id BIGINT UNSIGNED NOT NULL"},
			{name: "friend_id", addSQL: "ADD COLUMN friend_id BIGINT UNSIGNED NOT NULL"},
			{name: "created_at", addSQL: "ADD COLUMN created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP"},
		},
		indexes: []indexSchema{
			{name: "uk_pair", createSQL: "CREATE UNIQUE INDEX uk_pair ON friendships (user_id, friend_id)"},
			{name: "idx_user", createSQL: "CREATE INDEX idx_user ON friendships (user_id)"},
		},
		foreignKeys: []fkSchema{
			{
				name:       "fk_friendships_user",
				createSQL:  "ALTER TABLE friendships ADD CONSTRAINT fk_friendships_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE",
				referenced: "users",
			},
			{
				name:       "fk_friendships_friend",
				createSQL:  "ALTER TABLE friendships ADD CONSTRAINT fk_friendships_friend FOREIGN KEY (friend_id) REFERENCES users(id) ON DELETE CASCADE",
				referenced: "users",
			},
		},
	},
	{
		name: "friend_requests",
		createSQL: `CREATE TABLE IF NOT EXISTS friend_requests (
			id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
			from_user_id BIGINT UNSIGNED NOT NULL,
			to_user_id BIGINT UNSIGNED NOT NULL,
			status VARCHAR(16) NOT NULL DEFAULT 'pending',
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (id),
			UNIQUE KEY uk_request_pair (from_user_id, to_user_id),
			KEY idx_to_status (to_user_id, status)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		columns: []columnSchema{
			{name: "id", addSQL: "ADD COLUMN id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT FIRST"},
			{name: "from_user_id", addSQL: "ADD COLUMN from_user_id BIGINT UNSIGNED NOT NULL"},
			{name: "to_user_id", addSQL: "ADD COLUMN to_user_id BIGINT UNSIGNED NOT NULL"},
			{name: "status", addSQL: "ADD COLUMN status VARCHAR(16) NOT NULL DEFAULT 'pending'"},
			{name: "created_at", addSQL: "ADD COLUMN created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP"},
		},
		indexes: []indexSchema{
			{name: "uk_request_pair", createSQL: "CREATE UNIQUE INDEX uk_request_pair ON friend_requests (from_user_id, to_user_id)"},
			{name: "idx_to_status", createSQL: "CREATE INDEX idx_to_status ON friend_requests (to_user_id, status)"},
		},
		foreignKeys: []fkSchema{
			{
				name:       "fk_friend_requests_from",
				createSQL:  "ALTER TABLE friend_requests ADD CONSTRAINT fk_friend_requests_from FOREIGN KEY (from_user_id) REFERENCES users(id) ON DELETE CASCADE",
				referenced: "users",
			},
			{
				name:       "fk_friend_requests_to",
				createSQL:  "ALTER TABLE friend_requests ADD CONSTRAINT fk_friend_requests_to FOREIGN KEY (to_user_id) REFERENCES users(id) ON DELETE CASCADE",
				referenced: "users",
			},
		},
	},
	{
		name: "teams",
		createSQL: `CREATE TABLE IF NOT EXISTS teams (
			id BIGINT UNSIGNED NOT NULL,
			name VARCHAR(128) NOT NULL,
			ledger_id VARCHAR(32) NOT NULL,
			creator_id BIGINT UNSIGNED NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (id),
			KEY idx_teams_creator (creator_id),
			KEY idx_teams_ledger (ledger_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		columns: []columnSchema{
			{name: "id", addSQL: "ADD COLUMN id BIGINT UNSIGNED NOT NULL FIRST"},
			{name: "name", addSQL: "ADD COLUMN name VARCHAR(128) NOT NULL"},
			{name: "ledger_id", addSQL: "ADD COLUMN ledger_id VARCHAR(32) NOT NULL"},
			{name: "creator_id", addSQL: "ADD COLUMN creator_id BIGINT UNSIGNED NOT NULL"},
			{name: "created_at", addSQL: "ADD COLUMN created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP"},
		},
		foreignKeys: []fkSchema{
			{
				name:       "fk_teams_creator",
				createSQL:  "ALTER TABLE teams ADD CONSTRAINT fk_teams_creator FOREIGN KEY (creator_id) REFERENCES users(id) ON DELETE CASCADE",
				referenced: "users",
			},
		},
	},
	{
		name: "team_members",
		createSQL: `CREATE TABLE IF NOT EXISTS team_members (
			team_id BIGINT UNSIGNED NOT NULL,
			user_id BIGINT UNSIGNED NOT NULL,
			joined_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (team_id, user_id),
			KEY idx_team_members_user (user_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		columns: []columnSchema{
			{name: "team_id", addSQL: "ADD COLUMN team_id BIGINT UNSIGNED NOT NULL FIRST"},
			{name: "user_id", addSQL: "ADD COLUMN user_id BIGINT UNSIGNED NOT NULL"},
			{name: "joined_at", addSQL: "ADD COLUMN joined_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP"},
		},
		foreignKeys: []fkSchema{
			{
				name:       "fk_team_members_team",
				createSQL:  "ALTER TABLE team_members ADD CONSTRAINT fk_team_members_team FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE",
				referenced: "teams",
			},
			{
				name:       "fk_team_members_user",
				createSQL:  "ALTER TABLE team_members ADD CONSTRAINT fk_team_members_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE",
				referenced: "users",
			},
		},
	},
	{
		name: "team_ledgers",
		createSQL: `CREATE TABLE IF NOT EXISTS team_ledgers (
			team_id BIGINT UNSIGNED NOT NULL,
			ledger_id VARCHAR(32) NOT NULL,
			added_by_id BIGINT UNSIGNED NULL,
			added_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (team_id, ledger_id),
			KEY idx_team_ledgers_ledger (ledger_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		columns: []columnSchema{
			{name: "team_id", addSQL: "ADD COLUMN team_id BIGINT UNSIGNED NOT NULL FIRST"},
			{name: "ledger_id", addSQL: "ADD COLUMN ledger_id VARCHAR(32) NOT NULL"},
			{name: "added_by_id", addSQL: "ADD COLUMN added_by_id BIGINT UNSIGNED NULL"},
			{name: "added_at", addSQL: "ADD COLUMN added_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP"},
		},
		indexes: []indexSchema{
			{name: "idx_team_ledgers_ledger", createSQL: "CREATE INDEX idx_team_ledgers_ledger ON team_ledgers (ledger_id)"},
		},
		foreignKeys: []fkSchema{
			{
				name:       "fk_team_ledgers_team",
				createSQL:  "ALTER TABLE team_ledgers ADD CONSTRAINT fk_team_ledgers_team FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE",
				referenced: "teams",
			},
		},
	},
	{
		name: "team_messages",
		createSQL: `CREATE TABLE IF NOT EXISTS team_messages (
			id BIGINT UNSIGNED NOT NULL,
			team_id BIGINT UNSIGNED NOT NULL,
			sender_id BIGINT UNSIGNED NOT NULL,
			msg_type VARCHAR(16) NOT NULL DEFAULT 'text',
			body TEXT NULL,
			file_name VARCHAR(255) NOT NULL DEFAULT '',
			file_path VARCHAR(512) NOT NULL DEFAULT '',
			file_size BIGINT NOT NULL DEFAULT 0,
			content_type VARCHAR(128) NOT NULL DEFAULT '',
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (id),
			KEY idx_team_messages_team (team_id, id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		columns: []columnSchema{
			{name: "id", addSQL: "ADD COLUMN id BIGINT UNSIGNED NOT NULL FIRST"},
			{name: "team_id", addSQL: "ADD COLUMN team_id BIGINT UNSIGNED NOT NULL"},
			{name: "sender_id", addSQL: "ADD COLUMN sender_id BIGINT UNSIGNED NOT NULL"},
			{name: "msg_type", addSQL: "ADD COLUMN msg_type VARCHAR(16) NOT NULL DEFAULT 'text'"},
			{name: "body", addSQL: "ADD COLUMN body TEXT NULL"},
			{name: "file_name", addSQL: "ADD COLUMN file_name VARCHAR(255) NOT NULL DEFAULT ''"},
			{name: "file_path", addSQL: "ADD COLUMN file_path VARCHAR(512) NOT NULL DEFAULT ''"},
			{name: "file_size", addSQL: "ADD COLUMN file_size BIGINT NOT NULL DEFAULT 0"},
			{name: "content_type", addSQL: "ADD COLUMN content_type VARCHAR(128) NOT NULL DEFAULT ''"},
			{name: "created_at", addSQL: "ADD COLUMN created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP"},
		},
		indexes: []indexSchema{
			{name: "idx_team_messages_team", createSQL: "CREATE INDEX idx_team_messages_team ON team_messages (team_id, id)"},
		},
		foreignKeys: []fkSchema{
			{
				name:       "fk_team_messages_team",
				createSQL:  "ALTER TABLE team_messages ADD CONSTRAINT fk_team_messages_team FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE",
				referenced: "teams",
			},
			{
				name:       "fk_team_messages_sender",
				createSQL:  "ALTER TABLE team_messages ADD CONSTRAINT fk_team_messages_sender FOREIGN KEY (sender_id) REFERENCES users(id) ON DELETE CASCADE",
				referenced: "users",
			},
		},
	},
	{
		name: "team_read_state",
		createSQL: `CREATE TABLE IF NOT EXISTS team_read_state (
			team_id BIGINT UNSIGNED NOT NULL,
			user_id BIGINT UNSIGNED NOT NULL,
			last_read_message_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			PRIMARY KEY (team_id, user_id),
			KEY idx_team_read_user (user_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		columns: []columnSchema{
			{name: "team_id", addSQL: "ADD COLUMN team_id BIGINT UNSIGNED NOT NULL FIRST"},
			{name: "user_id", addSQL: "ADD COLUMN user_id BIGINT UNSIGNED NOT NULL"},
			{name: "last_read_message_id", addSQL: "ADD COLUMN last_read_message_id BIGINT UNSIGNED NOT NULL DEFAULT 0"},
			{name: "updated_at", addSQL: "ADD COLUMN updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"},
		},
		indexes: []indexSchema{
			{name: "idx_team_read_user", createSQL: "CREATE INDEX idx_team_read_user ON team_read_state (user_id)"},
		},
		foreignKeys: []fkSchema{
			{
				name:       "fk_team_read_team",
				createSQL:  "ALTER TABLE team_read_state ADD CONSTRAINT fk_team_read_team FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE",
				referenced: "teams",
			},
			{
				name:       "fk_team_read_user",
				createSQL:  "ALTER TABLE team_read_state ADD CONSTRAINT fk_team_read_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE",
				referenced: "users",
			},
		},
	},
	{
		name: "user_public_keys",
		createSQL: `CREATE TABLE IF NOT EXISTS user_public_keys (
			user_id BIGINT UNSIGNED NOT NULL,
			public_key_pem TEXT NOT NULL,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			PRIMARY KEY (user_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		columns: []columnSchema{
			{name: "user_id", addSQL: "ADD COLUMN user_id BIGINT UNSIGNED NOT NULL FIRST"},
			{name: "public_key_pem", addSQL: "ADD COLUMN public_key_pem TEXT NOT NULL"},
			{name: "updated_at", addSQL: "ADD COLUMN updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"},
		},
		foreignKeys: []fkSchema{
			{
				name:       "fk_user_public_keys_user",
				createSQL:  "ALTER TABLE user_public_keys ADD CONSTRAINT fk_user_public_keys_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE",
				referenced: "users",
			},
		},
	},
	{
		name: "entry_templates",
		createSQL: `CREATE TABLE IF NOT EXISTS entry_templates (
			id BIGINT UNSIGNED NOT NULL,
			owner_id BIGINT UNSIGNED NOT NULL,
			name VARCHAR(128) NOT NULL,
			fields_json JSON NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			PRIMARY KEY (id),
			KEY idx_entry_templates_owner (owner_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		columns: []columnSchema{
			{name: "id", addSQL: "ADD COLUMN id BIGINT UNSIGNED NOT NULL FIRST"},
			{name: "owner_id", addSQL: "ADD COLUMN owner_id BIGINT UNSIGNED NOT NULL"},
			{name: "name", addSQL: "ADD COLUMN name VARCHAR(128) NOT NULL"},
			{name: "fields_json", addSQL: "ADD COLUMN fields_json JSON NOT NULL"},
			{name: "created_at", addSQL: "ADD COLUMN created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP"},
			{name: "updated_at", addSQL: "ADD COLUMN updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"},
		},
		foreignKeys: []fkSchema{
			{
				name:       "fk_entry_templates_owner",
				createSQL:  "ALTER TABLE entry_templates ADD CONSTRAINT fk_entry_templates_owner FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE",
				referenced: "users",
			},
		},
	},
}

type tableSchema struct {
	name        string
	createSQL   string
	columns     []columnSchema
	indexes     []indexSchema
	foreignKeys []fkSchema
}

type columnSchema struct {
	name   string
	addSQL string
}

type indexSchema struct {
	name      string
	createSQL string
}

type fkSchema struct {
	name       string
	createSQL  string
	referenced string
}
