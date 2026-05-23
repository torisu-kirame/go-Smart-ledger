package db

// 与 infra/sql/001_schema.sql 保持一致，供启动时存在性检测与补齐。
const defaultCharset = "utf8mb4"
const defaultCollation = "utf8mb4_unicode_ci"

var schemaTables = []tableSchema{
	{
		name: "users",
		createSQL: `CREATE TABLE IF NOT EXISTS users (
			id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
			username VARCHAR(64) NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (id),
			UNIQUE KEY uk_username (username)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		columns: []columnSchema{
			{name: "id", addSQL: "ADD COLUMN id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT FIRST"},
			{name: "username", addSQL: "ADD COLUMN username VARCHAR(64) NOT NULL"},
			{name: "password_hash", addSQL: "ADD COLUMN password_hash VARCHAR(255) NOT NULL"},
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
