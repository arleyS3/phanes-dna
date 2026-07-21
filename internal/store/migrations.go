package store

import (
	"database/sql"
	"fmt"
	"time"
)

type migration struct {
	version int
	apply   func(*sql.Tx) error
}

var migrations = []migration{
	{version: 1, apply: migration001},
}

func (s *Store) migrate() error {
	// Create the schema version table if it doesn't exist.
	if _, err := s.db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version     INTEGER PRIMARY KEY,
		applied_at  TEXT NOT NULL
	)`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	// Read the current max version.
	var current int
	err := s.db.QueryRow(`SELECT COALESCE(MAX(version), 0) FROM schema_migrations`).Scan(&current)
	if err != nil {
		return fmt.Errorf("read schema version: %w", err)
	}

	// Apply pending migrations in order.
	for _, m := range migrations {
		if m.version > current {
			tx, err := s.db.Begin()
			if err != nil {
				return fmt.Errorf("begin tx for migration %d: %w", m.version, err)
			}
			if err := m.apply(tx); err != nil {
				_ = tx.Rollback()
				return fmt.Errorf("migration %d: %w", m.version, err)
			}
			if _, err := tx.Exec(`INSERT INTO schema_migrations(version, applied_at) VALUES(?, ?)`,
				m.version, time.Now().UTC().Format(time.RFC3339)); err != nil {
				_ = tx.Rollback()
				return fmt.Errorf("record migration %d: %w", m.version, err)
			}
			if err := tx.Commit(); err != nil {
				return fmt.Errorf("commit migration %d: %w", m.version, err)
			}
		}
	}
	return nil
}

func migration001(tx *sql.Tx) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS projects (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			name        TEXT NOT NULL,
			root_path   TEXT NOT NULL,
			analyzed_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS source_files (
			id           INTEGER PRIMARY KEY AUTOINCREMENT,
			project_id   INTEGER NOT NULL,
			rel_path     TEXT NOT NULL,
			content_hash TEXT NOT NULL DEFAULT '',
			package_path TEXT NOT NULL DEFAULT '',
			layer        TEXT NOT NULL DEFAULT '',
			FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS ast_nodes (
			id               INTEGER PRIMARY KEY AUTOINCREMENT,
			file_id          INTEGER NOT NULL,
			parent_id        INTEGER,
			node_type        TEXT NOT NULL,
			name             TEXT NOT NULL DEFAULT '',
			annotations_json TEXT NOT NULL DEFAULT '[]',
			start_line       INTEGER NOT NULL DEFAULT 0,
			FOREIGN KEY (file_id) REFERENCES source_files(id) ON DELETE CASCADE,
			FOREIGN KEY (parent_id) REFERENCES ast_nodes(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS dependencies (
			id           INTEGER PRIMARY KEY AUTOINCREMENT,
			src_node_id  INTEGER NOT NULL,
			tgt_node_id  INTEGER NOT NULL,
			dep_type     TEXT NOT NULL DEFAULT '',
			FOREIGN KEY (src_node_id) REFERENCES ast_nodes(id) ON DELETE CASCADE,
			FOREIGN KEY (tgt_node_id) REFERENCES ast_nodes(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS layer_rules (
			id        INTEGER PRIMARY KEY AUTOINCREMENT,
			src_layer TEXT NOT NULL,
			tgt_layer TEXT NOT NULL,
			allowed   INTEGER NOT NULL DEFAULT 1,
			severity  TEXT NOT NULL DEFAULT 'info'
		)`,
		`CREATE TABLE IF NOT EXISTS chunks (
			id        INTEGER PRIMARY KEY AUTOINCREMENT,
			file_id   INTEGER NOT NULL,
			chunk_idx INTEGER NOT NULL,
			content   TEXT NOT NULL,
			embedding BLOB,
			FOREIGN KEY (file_id) REFERENCES source_files(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS violations (
			id             INTEGER PRIMARY KEY AUTOINCREMENT,
			file           TEXT NOT NULL,
			line           INTEGER NOT NULL DEFAULT 0,
			from_class     TEXT NOT NULL DEFAULT '',
			to_class       TEXT NOT NULL DEFAULT '',
			severity       TEXT NOT NULL DEFAULT 'info',
			rule_ref       TEXT NOT NULL DEFAULT '',
			recommendation TEXT NOT NULL DEFAULT ''
		)`,
		// Indexes for common query paths.
		`CREATE INDEX IF NOT EXISTS idx_source_files_project ON source_files(project_id)`,
		`CREATE INDEX IF NOT EXISTS idx_ast_nodes_file ON ast_nodes(file_id)`,
		`CREATE INDEX IF NOT EXISTS idx_ast_nodes_parent ON ast_nodes(parent_id)`,
		`CREATE INDEX IF NOT EXISTS idx_dependencies_src ON dependencies(src_node_id)`,
		`CREATE INDEX IF NOT EXISTS idx_dependencies_tgt ON dependencies(tgt_node_id)`,
		`CREATE INDEX IF NOT EXISTS idx_chunks_file ON chunks(file_id)`,
		`CREATE INDEX IF NOT EXISTS idx_violations_file ON violations(file)`,
	}

	for _, stmt := range stmts {
		if _, err := tx.Exec(stmt); err != nil {
			return fmt.Errorf("exec: %s: %w", stmt[:60], err)
		}
	}
	return nil
}
