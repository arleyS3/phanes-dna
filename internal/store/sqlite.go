package store

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strings"
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

// GetViolations returns all stored violations, recalculating active architecture rules first.
func (s *Store) GetViolations(ctx context.Context) ([]*dna.Violation, error) {
	// Clear all previous violations globally before auditing all projects
	_, _ = s.db.ExecContext(ctx, "DELETE FROM violations")

	projects, err := s.ListProjects(ctx)
	if err == nil {
		for _, p := range projects {
			// Try to sync rules from PHANES_RULES.md in the project root if it exists
			mdPath := filepath.Join(p.RootPath, "PHANES_RULES.md")
			if _, errStat := os.Stat(mdPath); errStat == nil {
				_ = s.SyncRulesFromMarkdown(ctx, mdPath)
			}

			if errAud := s.AuditArchitecture(ctx, p.ID); errAud != nil {
				fmt.Fprintf(os.Stderr, "  ⚠️ Architecture audit failed for project %d: %v\n", p.ID, errAud)
			}

			if errNaming := s.AuditNamingConventions(ctx, p.ID); errNaming != nil {
				fmt.Fprintf(os.Stderr, "  ⚠️ Naming audit failed for project %d: %v\n", p.ID, errNaming)
			}
		}
	}

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

// AuditArchitecture recalculates architectural violations based on imported dependencies and layer rules.
func (s *Store) AuditArchitecture(ctx context.Context, projectID int64) error {
	// Rules are validated for the given projectID context.

	// 2. Load layer rules
	rowsRules, err := s.db.QueryContext(ctx, "SELECT src_layer, tgt_layer, allowed, severity FROM layer_rules")
	if err != nil {
		return fmt.Errorf("list layer rules: %w", err)
	}
	defer rowsRules.Close()

	type ruleKey struct {
		src string
		tgt string
	}
	rules := make(map[ruleKey]struct {
		allowed  bool
		severity string
	})
	for rowsRules.Next() {
		var src, tgt, sev string
		var allowedInt int
		if err := rowsRules.Scan(&src, &tgt, &allowedInt, &sev); err != nil {
			return err
		}
		rules[ruleKey{src, tgt}] = struct {
			allowed  bool
			severity string
		}{allowedInt == 1, sev}
	}

	// 3. Load all project source files
	type sourceFileInfo struct {
		id      int64
		relPath string
		layer   string
		lang    string
	}
	rowsFiles, err := s.db.QueryContext(ctx, "SELECT id, rel_path, layer FROM source_files WHERE project_id = ?", projectID)
	if err != nil {
		return fmt.Errorf("list source files: %w", err)
	}
	defer rowsFiles.Close()

	var files []sourceFileInfo
	for rowsFiles.Next() {
		var f sourceFileInfo
		if err := rowsFiles.Scan(&f.id, &f.relPath, &f.layer); err != nil {
			return err
		}
		
		// Infer language from file extension
		ext := filepath.Ext(f.relPath)
		if ext == ".go" {
			f.lang = "go"
		} else if ext == ".java" {
			f.lang = "java"
		}

		files = append(files, f)
	}

	// 4. Audit imports for each file
	for _, f := range files {
		if f.layer == "" {
			continue
		}

		rowsImports, err := s.db.QueryContext(ctx, "SELECT name, start_line FROM ast_nodes WHERE file_id = ? AND node_type = 'import'", f.id)
		if err != nil {
			return fmt.Errorf("list imports for %s: %w", f.relPath, err)
		}

		type importInfo struct {
			name string
			line int
		}
		var imports []importInfo
		for rowsImports.Next() {
			var imp importInfo
			if err := rowsImports.Scan(&imp.name, &imp.line); err == nil {
				imports = append(imports, imp)
			}
		}
		rowsImports.Close()

		for _, imp := range imports {
			for _, otherF := range files {
				if otherF.id == f.id || otherF.layer == "" || otherF.layer == f.layer {
					continue
				}

				matched := false

				if f.lang == "go" && otherF.lang == "go" {
					dir := filepath.Dir(otherF.relPath)
					if dir != "." && dir != "/" && strings.HasSuffix(imp.name, dir) {
						matched = true
					}
				} else if f.lang == "java" && otherF.lang == "java" {
					javaIdx := strings.Index(otherF.relPath, "/java/")
					if javaIdx != -1 {
						classPath := otherF.relPath[javaIdx+6:]
						classPath = strings.TrimSuffix(classPath, ".java")
						classPath = strings.ReplaceAll(classPath, "/", ".")
						pkgPath := classPath
						lastDot := strings.LastIndex(classPath, ".")
						if lastDot != -1 {
							pkgPath = classPath[:lastDot]
						}

						cleanImp := strings.TrimPrefix(imp.name, "import ")
						cleanImp = strings.TrimSuffix(cleanImp, ";")
						cleanImp = strings.TrimSpace(cleanImp)

						if cleanImp == classPath || cleanImp == pkgPath+".*" {
							matched = true
						}
					}
				}

				if matched {
					r, hasRule := rules[ruleKey{f.layer, otherF.layer}]
					if hasRule && !r.allowed {
						recommendation := fmt.Sprintf("Capa '%s' tiene prohibido importar '%s' (violación en %s hacia %s).", f.layer, otherF.layer, f.relPath, otherF.relPath)
						ruleRef := fmt.Sprintf("%s-imports-%s", f.layer, otherF.layer)
						
						_, _ = s.db.ExecContext(ctx,
							`INSERT INTO violations(file, line, from_class, to_class, severity, rule_ref, recommendation)
							 VALUES(?, ?, ?, ?, ?, ?, ?)`,
							f.relPath, imp.line, f.layer, otherF.layer, r.severity, ruleRef, recommendation)
					}
				}
			}
		}
	}

	return nil
}

// AuditNamingConventions checks if the source file names comply with naming patterns defined in PHANES_RULES.md
func (s *Store) AuditNamingConventions(ctx context.Context, projectID int64) error {
	// 1. Get project root to locate PHANES_RULES.md
	var rootPath string
	err := s.db.QueryRowContext(ctx, "SELECT root_path FROM projects WHERE id = ?", projectID).Scan(&rootPath)
	if err != nil {
		return fmt.Errorf("get project root: %w", err)
	}

	rulesPath := filepath.Join(rootPath, "PHANES_RULES.md")
	if _, err := os.Stat(rulesPath); err != nil {
		// If rules file is missing, nothing to audit
		return nil
	}

	// 2. Parse Case Style and Patterns from PHANES_RULES.md
	caseStyle, patterns, err := parseNamingRules(rulesPath)
	if err != nil {
		return fmt.Errorf("parse naming rules: %w", err)
	}

	if len(patterns) == 0 {
		return nil
	}

	// 3. Get all files in project
	rows, err := s.db.QueryContext(ctx, "SELECT rel_path, layer FROM source_files WHERE project_id = ?", projectID)
	if err != nil {
		return fmt.Errorf("list source files: %w", err)
	}
	defer rows.Close()

	type fileInfo struct {
		relPath string
		layer   string
	}
	var files []fileInfo
	for rows.Next() {
		var f fileInfo
		if err := rows.Scan(&f.relPath, &f.layer); err == nil {
			files = append(files, f)
		}
	}

	// Helper for case regex
	getCaseStyleRegex := func(style string) string {
		switch strings.ToLower(strings.ReplaceAll(style, " ", "")) {
		case "snake_case":
			return `[a-z0-9]+(_[a-z0-9]+)*`
		case "kebab-case":
			return `[a-z0-9]+(-[a-z0-9]+)*`
		case "camelcase":
			return `[a-z0-9]+([A-Z][a-z0-9]+)*`
		case "pascalcase":
			return `[A-Z][a-z0-9]+([A-Z][a-z0-9]+)*`
		default:
			return `.+`
		}
	}

	// 4. Validate each file against layer pattern
	for _, f := range files {
		if f.layer == "" {
			continue
		}

		pattern, hasPattern := patterns[f.layer]
		if !hasPattern {
			continue
		}

		// Compile regex pattern
		token := "__NAME_TOKEN__"
		patternWithToken := strings.ReplaceAll(pattern, "[name]", token)
		escapedPattern := regexp.QuoteMeta(patternWithToken)
		regexStr := "^" + strings.ReplaceAll(escapedPattern, token, getCaseStyleRegex(caseStyle)) + "$"

		re, err := regexp.Compile(regexStr)
		if err != nil {
			continue
		}

		fileName := filepath.Base(f.relPath)
		if !re.MatchString(fileName) {
			recommendation := fmt.Sprintf("Nomenclatura inválida en capa '%s': el archivo '%s' no cumple con el patrón '%s' en formato '%s'.", f.layer, fileName, pattern, caseStyle)
			ruleRef := fmt.Sprintf("%s-naming-convention", f.layer)

			_, _ = s.db.ExecContext(ctx,
				`INSERT INTO violations(file, line, from_class, to_class, severity, rule_ref, recommendation)
				 VALUES(?, ?, ?, ?, ?, ?, ?)`,
				f.relPath, 1, f.layer, "", "warning", ruleRef, recommendation)
		}
	}

	return nil
}

func parseNamingRules(path string) (string, map[string]string, error) {
	caseStyle := "snake_case"
	patterns := make(map[string]string)

	file, err := os.Open(path)
	if err != nil {
		return caseStyle, patterns, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	inSection := false

	caseRegex := regexp.MustCompile(`(?i)-\s*case\s*style:\s*` + "`" + `([^` + "`" + `]+)` + "`")
	patternRegex := regexp.MustCompile(`(?i)-\s*([a-zA-Z0-9_\-]+):\s*` + "`" + `([^` + "`" + `]+)` + "`")

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, "##") {
			if strings.Contains(strings.ToLower(line), "naming") || strings.Contains(strings.ToLower(line), "scaffolding") {
				inSection = true
			} else {
				inSection = false
			}
			continue
		}

		if inSection {
			if match := caseRegex.FindStringSubmatch(line); len(match) > 1 {
				caseStyle = strings.TrimSpace(match[1])
			} else if match := patternRegex.FindStringSubmatch(line); len(match) > 1 {
				patterns[strings.TrimSpace(match[1])] = strings.TrimSpace(match[2])
			}
		}
	}

	return caseStyle, patterns, scanner.Err()
}

// SyncRulesFromMarkdown reads PHANES_RULES.md from project root and populates layer_rules in SQLite.
func (s *Store) SyncRulesFromMarkdown(ctx context.Context, mdPath string) error {
	file, err := os.Open(mdPath)
	if err != nil {
		return fmt.Errorf("open rules markdown: %w", err)
	}
	defer file.Close()

	// Clear existing rules in DB
	_, err = s.db.ExecContext(ctx, "DELETE FROM layer_rules")
	if err != nil {
		return fmt.Errorf("clear layer rules: %w", err)
	}

	scanner := bufio.NewScanner(file)
	inLayerDeps := false
	// Regular expression to match: - `src` cannot/can import `tgt`
	ruleRe := regexp.MustCompile(`^\s*-\s*` + "`" + `([A-Za-z0-9_-]+)` + "`" + `\s+(cannot|can)\s+import\s+` + "`" + `([A-Za-z0-9_-]+)` + "`" + ``)

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "## ") {
			if strings.Contains(strings.ToLower(trimmed), "layer dependencies") {
				inLayerDeps = true
			} else {
				inLayerDeps = false
			}
			continue
		}

		if inLayerDeps && strings.HasPrefix(trimmed, "-") {
			m := ruleRe.FindStringSubmatch(line)
			if len(m) > 3 {
				src := m[1]
				action := m[2]
				tgt := m[3]
				allowed := 1
				if action == "cannot" {
					allowed = 0
				}
				severity := "error"
				if allowed == 1 {
					severity = "info"
				}

				_, err = s.db.ExecContext(ctx,
					"INSERT INTO layer_rules (src_layer, tgt_layer, allowed, severity) VALUES (?, ?, ?, ?)",
					src, tgt, allowed, severity)
				if err != nil {
					return fmt.Errorf("insert layer rule: %w", err)
				}
			}
		}
	}

	return scanner.Err()
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
