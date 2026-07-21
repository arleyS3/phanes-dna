package store

import (
	"context"
	"database/sql"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/arleyS3/phanes-dna/internal/dna"

	_ "modernc.org/sqlite"
)

// Store wraps a SQLite database with CRUD operations for Phanes DNA models.
type Store struct {
	db *sql.DB
}

// NewStore opens (or creates) a SQLite database at path and runs pending
// migrations. Use ":memory:" for an ephemeral test database.
func NewStore(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	// Sensible pragmas for performance and concurrency.
	pragmas := []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA busy_timeout=5000",
		"PRAGMA foreign_keys=ON",
	}
	for _, p := range pragmas {
		if _, err := db.Exec(p); err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("%s: %w", p, err)
		}
	}

	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return s, nil
}

// Close releases the database connection.
func (s *Store) Close() error {
	return s.db.Close()
}

// ---------------------------------------------------------------------------
// Projects
// ---------------------------------------------------------------------------

// InsertProject stores a project and returns its row ID.
func (s *Store) InsertProject(ctx context.Context, p *dna.Project) (int64, error) {
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO projects(name, root_path, analyzed_at) VALUES(?, ?, ?)`,
		p.Name, p.RootPath, p.DetectedAt.UTC().Format(time.RFC3339))
	if err != nil {
		return 0, fmt.Errorf("insert project: %w", err)
	}
	return res.LastInsertId()
}

// GetProject retrieves a project by ID.
func (s *Store) GetProject(ctx context.Context, id int64) (*dna.Project, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, name, root_path, analyzed_at FROM projects WHERE id = ?`, id)
	var p dna.Project
	var analyzedAt string
	if err := row.Scan(&p.ID, &p.Name, &p.RootPath, &analyzedAt); err != nil {
		return nil, fmt.Errorf("get project %d: %w", id, err)
	}
	t, err := time.Parse(time.RFC3339, analyzedAt)
	if err != nil {
		return nil, fmt.Errorf("parse analyzed_at %q: %w", analyzedAt, err)
	}
	p.DetectedAt = t
	return &p, nil
}

// ListProjects returns all stored projects.
func (s *Store) ListProjects(ctx context.Context) ([]*dna.Project, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, name, root_path, analyzed_at FROM projects ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}
	defer rows.Close()

	var projects []*dna.Project
	for rows.Next() {
		var p dna.Project
		var analyzedAt string
		if err := rows.Scan(&p.ID, &p.Name, &p.RootPath, &analyzedAt); err != nil {
			return nil, fmt.Errorf("scan project: %w", err)
		}
		t, err := time.Parse(time.RFC3339, analyzedAt)
		if err != nil {
			return nil, fmt.Errorf("parse analyzed_at %q: %w", analyzedAt, err)
		}
		p.DetectedAt = t
		projects = append(projects, &p)
	}
	return projects, rows.Err()
}

// ---------------------------------------------------------------------------
// SourceFiles
// ---------------------------------------------------------------------------

// InsertSourceFile stores a source file linked to a project and returns its row ID.
func (s *Store) InsertSourceFile(ctx context.Context, projectID int64, sf *dna.SourceFile) (int64, error) {
	layer := ""
	if len(sf.Layers) > 0 {
		layer = sf.Layers[0]
	}
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO source_files(project_id, rel_path, package_path, layer) VALUES(?, ?, ?, ?)`,
		projectID, sf.Path, "", layer)
	if err != nil {
		return 0, fmt.Errorf("insert source_file: %w", err)
	}
	return res.LastInsertId()
}

// GetSourceFilesByProject returns all source files for a given project.
func (s *Store) GetSourceFilesByProject(ctx context.Context, projectID int64) ([]*dna.SourceFile, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT rel_path, layer FROM source_files WHERE project_id = ? ORDER BY id`, projectID)
	if err != nil {
		return nil, fmt.Errorf("list source_files: %w", err)
	}
	defer rows.Close()

	var files []*dna.SourceFile
	for rows.Next() {
		var relPath, layer string
		if err := rows.Scan(&relPath, &layer); err != nil {
			return nil, fmt.Errorf("scan source_file: %w", err)
		}
		layers := []string{}
		if layer != "" {
			layers = append(layers, layer)
		}
		files = append(files, &dna.SourceFile{
			Path:   relPath,
			Layers: layers,
		})
	}
	return files, rows.Err()
}

// ---------------------------------------------------------------------------
// ASTNodes
// ---------------------------------------------------------------------------

// InsertASTNode stores an AST node and returns its row ID. Set parentID to nil for root nodes.
func (s *Store) InsertASTNode(ctx context.Context, fileID int64, parentID *int64, n *dna.ASTNode) (int64, error) {
	annJSON, err := json.Marshal(n.Children)
	if err != nil {
		return 0, fmt.Errorf("marshal annotations: %w", err)
	}

	res, err := s.db.ExecContext(ctx,
		`INSERT INTO ast_nodes(file_id, parent_id, node_type, name, annotations_json, start_line) VALUES(?, ?, ?, ?, ?, ?)`,
		fileID, parentID, n.Type, n.Name, string(annJSON), n.Line)
	if err != nil {
		return 0, fmt.Errorf("insert ast_node: %w", err)
	}
	return res.LastInsertId()
}

// GetASTNodesByFile returns all AST nodes for a given source file, organised as a flat list.
// The caller can reconstruct the tree by matching parent_id.
func (s *Store) GetASTNodesByFile(ctx context.Context, fileID int64) ([]*dna.ASTNode, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, parent_id, node_type, name, start_line FROM ast_nodes WHERE file_id = ? ORDER BY id`, fileID)
	if err != nil {
		return nil, fmt.Errorf("list ast_nodes: %w", err)
	}
	defer rows.Close()

	var nodes []*dna.ASTNode
	for rows.Next() {
		var n dna.ASTNode
		var id int64
		var parentID sql.NullInt64
		if err := rows.Scan(&id, &parentID, &n.Type, &n.Name, &n.Line); err != nil {
			return nil, fmt.Errorf("scan ast_node: %w", err)
		}
		nodes = append(nodes, &n)
	}
	return nodes, rows.Err()
}

// ---------------------------------------------------------------------------
// Dependencies
// ---------------------------------------------------------------------------

// InsertDependency stores a dependency edge between two AST nodes and returns its row ID.
func (s *Store) InsertDependency(ctx context.Context, d *dna.Dependency) (int64, error) {
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO dependencies(src_node_id, tgt_node_id, dep_type) VALUES(?, ?, ?)`,
		0, 0, d.Type)
	if err != nil {
		return 0, fmt.Errorf("insert dependency: %w", err)
	}
	return res.LastInsertId()
}

// GetDependencies returns all stored dependencies.
func (s *Store) GetDependencies(ctx context.Context) ([]*dna.Dependency, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT dep_type FROM dependencies ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("list dependencies: %w", err)
	}
	defer rows.Close()

	var deps []*dna.Dependency
	for rows.Next() {
		var d dna.Dependency
		if err := rows.Scan(&d.Type); err != nil {
			return nil, fmt.Errorf("scan dependency: %w", err)
		}
		deps = append(deps, &d)
	}
	return deps, rows.Err()
}

// ---------------------------------------------------------------------------
// LayerRules
// ---------------------------------------------------------------------------

// InsertLayerRule stores a layer rule and returns its row ID.
func (s *Store) InsertLayerRule(ctx context.Context, r *dna.LayerRule) (int64, error) {
	allowedInt := 0
	if r.Allowed {
		allowedInt = 1
	}
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO layer_rules(src_layer, tgt_layer, allowed, severity) VALUES(?, ?, ?, ?)`,
		r.Layer, r.TargetLayer, allowedInt, r.Severity)
	if err != nil {
		return 0, fmt.Errorf("insert layer_rule: %w", err)
	}
	return res.LastInsertId()
}

// GetLayerRules returns all stored layer rules.
func (s *Store) GetLayerRules(ctx context.Context) ([]*dna.LayerRule, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT src_layer, tgt_layer, allowed, severity FROM layer_rules ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("list layer_rules: %w", err)
	}
	defer rows.Close()

	var rules []*dna.LayerRule
	for rows.Next() {
		var r dna.LayerRule
		var allowed int
		if err := rows.Scan(&r.Layer, &r.TargetLayer, &allowed, &r.Severity); err != nil {
			return nil, fmt.Errorf("scan layer_rule: %w", err)
		}
		r.Allowed = (allowed != 0)
		rules = append(rules, &r)
	}
	return rules, rows.Err()
}

// ---------------------------------------------------------------------------
// Chunks
// ---------------------------------------------------------------------------

// InsertChunk stores a chunk (embedding is optional) and returns its row ID.
func (s *Store) InsertChunk(ctx context.Context, fileID int64, c *dna.Chunk) (int64, error) {
	var emb []byte
	if len(c.Embedding) > 0 {
		emb = floatsToBytes(c.Embedding)
	}
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO chunks(file_id, chunk_idx, content, embedding) VALUES(?, ?, ?, ?)`,
		fileID, c.StartLine, c.Content, emb)
	if err != nil {
		return 0, fmt.Errorf("insert chunk: %w", err)
	}
	return res.LastInsertId()
}

// GetChunksByFile returns all chunks for a given file.
func (s *Store) GetChunksByFile(ctx context.Context, fileID int64) ([]*dna.Chunk, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, chunk_idx, content, embedding FROM chunks WHERE file_id = ? ORDER BY chunk_idx`, fileID)
	if err != nil {
		return nil, fmt.Errorf("list chunks: %w", err)
	}
	defer rows.Close()

	var chunks []*dna.Chunk
	for rows.Next() {
		var c dna.Chunk
		var id int64
		var emb []byte
		if err := rows.Scan(&id, &c.StartLine, &c.Content, &emb); err != nil {
			return nil, fmt.Errorf("scan chunk: %w", err)
		}
		if emb != nil {
			c.Embedding = bytesToFloats(emb)
		}
		chunks = append(chunks, &c)
	}
	return chunks, rows.Err()
}

// ---------------------------------------------------------------------------
// Violations
// ---------------------------------------------------------------------------

// InsertViolation stores a violation and returns its row ID.
func (s *Store) InsertViolation(ctx context.Context, v *dna.Violation) (int64, error) {
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO violations(file, line, from_class, to_class, severity, rule_ref, recommendation) VALUES(?, ?, ?, ?, ?, ?, ?)`,
		v.Location, 0, "", "", v.Severity, v.Rule, v.Message)
	if err != nil {
		return 0, fmt.Errorf("insert violation: %w", err)
	}
	return res.LastInsertId()
}

// GetViolations returns all stored violations.
func (s *Store) GetViolations(ctx context.Context) ([]*dna.Violation, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT file, severity, rule_ref, recommendation FROM violations ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("list violations: %w", err)
	}
	defer rows.Close()

	var violations []*dna.Violation
	for rows.Next() {
		var v dna.Violation
		if err := rows.Scan(&v.Location, &v.Severity, &v.Rule, &v.Message); err != nil {
			return nil, fmt.Errorf("scan violation: %w", err)
		}
		violations = append(violations, &v)
	}
	return violations, rows.Err()
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

// floatsToBytes serialises a float32 slice to a little-endian byte blob for
// storage as a SQLite BLOB. Returns nil for a nil/empty slice.
func floatsToBytes(f []float32) []byte {
	if len(f) == 0 {
		return nil
	}
	b := make([]byte, len(f)*4)
	for i, v := range f {
		binary.LittleEndian.PutUint32(b[i*4:], math.Float32bits(v))
	}
	return b
}

// bytesToFloats deserialises a little-endian byte blob back to a float32 slice.
func bytesToFloats(b []byte) []float32 {
	if len(b) == 0 || len(b)%4 != 0 {
		return nil
	}
	f := make([]float32, len(b)/4)
	for i := range f {
		f[i] = math.Float32frombits(binary.LittleEndian.Uint32(b[i*4:]))
	}
	return f
}
