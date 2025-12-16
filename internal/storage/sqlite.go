package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
	"github.com/gongahkia/kite/pkg/models"
)

// SQLiteStorage implements the Storage interface using SQLite
type SQLiteStorage struct {
	db *sql.DB
}

// NewSQLiteStorage creates a new SQLite storage adapter
func NewSQLiteStorage(dbPath string) (*SQLiteStorage, error) {
	// Add pragmas for better performance and safety
	connStr := fmt.Sprintf("%s?_journal_mode=WAL&_synchronous=NORMAL&_foreign_keys=ON", dbPath)

	db, err := sql.Open("sqlite3", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Set connection pool settings (SQLite benefits from limited connections)
	db.SetMaxOpenConns(1) // SQLite performs best with single writer
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)

	storage := &SQLiteStorage{db: db}

	// Initialize schema
	if err := storage.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return storage, nil
}

// Close closes the database connection
func (ss *SQLiteStorage) Close() error {
	return ss.db.Close()
}

// initSchema creates the necessary tables
func (ss *SQLiteStorage) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS cases (
		id TEXT PRIMARY KEY,
		case_number TEXT,
		case_name TEXT NOT NULL,
		decision_date DATETIME,
		court TEXT,
		court_level INTEGER,
		court_type TEXT,
		jurisdiction TEXT,
		docket TEXT,
		parties TEXT, -- JSON
		judges TEXT, -- JSON
		summary TEXT,
		full_text TEXT,
		key_issues TEXT, -- JSON
		legal_concepts TEXT, -- JSON
		outcome TEXT,
		procedural_history TEXT,
		citations TEXT, -- JSON
		url TEXT,
		pdf_url TEXT,
		source_database TEXT,
		scraped_at DATETIME,
		last_updated DATETIME,
		language TEXT,
		status TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_cases_jurisdiction ON cases(jurisdiction);
	CREATE INDEX IF NOT EXISTS idx_cases_court ON cases(court);
	CREATE INDEX IF NOT EXISTS idx_cases_decision_date ON cases(decision_date);
	CREATE INDEX IF NOT EXISTS idx_cases_case_name ON cases(case_name);
	CREATE INDEX IF NOT EXISTS idx_cases_status ON cases(status);

	CREATE TABLE IF NOT EXISTS judges (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		full_name TEXT,
		title TEXT,
		court TEXT,
		jurisdiction TEXT,
		appointed_date DATETIME,
		biography TEXT,
		education TEXT, -- JSON
		career TEXT, -- JSON
		notable_cases TEXT, -- JSON
		total_cases INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_judges_name ON judges(name);
	CREATE INDEX IF NOT EXISTS idx_judges_court ON judges(court);
	CREATE INDEX IF NOT EXISTS idx_judges_jurisdiction ON judges(jurisdiction);

	CREATE TABLE IF NOT EXISTS citations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		format TEXT,
		raw_citation TEXT NOT NULL,
		normalized_citation TEXT,
		volume TEXT,
		reporter TEXT,
		page TEXT,
		year TEXT,
		court TEXT,
		case_number TEXT,
		country TEXT,
		citing_case_id TEXT,
		cited_case_id TEXT,
		is_normalized INTEGER DEFAULT 0, -- SQLite uses INTEGER for boolean
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (citing_case_id) REFERENCES cases(id),
		FOREIGN KEY (cited_case_id) REFERENCES cases(id)
	);

	CREATE INDEX IF NOT EXISTS idx_citations_citing_case ON citations(citing_case_id);
	CREATE INDEX IF NOT EXISTS idx_citations_cited_case ON citations(cited_case_id);
	CREATE INDEX IF NOT EXISTS idx_citations_format ON citations(format);

	-- Full-text search support
	CREATE VIRTUAL TABLE IF NOT EXISTS cases_fts USING fts5(
		id UNINDEXED,
		case_name,
		summary,
		full_text,
		content=cases,
		content_rowid=rowid
	);

	-- Triggers to keep FTS index updated
	CREATE TRIGGER IF NOT EXISTS cases_fts_insert AFTER INSERT ON cases BEGIN
		INSERT INTO cases_fts(id, case_name, summary, full_text)
		VALUES (new.id, new.case_name, new.summary, new.full_text);
	END;

	CREATE TRIGGER IF NOT EXISTS cases_fts_delete AFTER DELETE ON cases BEGIN
		DELETE FROM cases_fts WHERE id = old.id;
	END;

	CREATE TRIGGER IF NOT EXISTS cases_fts_update AFTER UPDATE ON cases BEGIN
		DELETE FROM cases_fts WHERE id = old.id;
		INSERT INTO cases_fts(id, case_name, summary, full_text)
		VALUES (new.id, new.case_name, new.summary, new.full_text);
	END;
	`

	_, err := ss.db.Exec(schema)
	return err
}

// SaveCase saves or updates a case
func (ss *SQLiteStorage) SaveCase(ctx context.Context, c *models.Case) error {
	query := `
		INSERT OR REPLACE INTO cases (
			id, case_number, case_name, decision_date, court, court_level, court_type,
			jurisdiction, docket, parties, judges, summary, full_text, key_issues,
			legal_concepts, outcome, procedural_history, citations, url, pdf_url,
			source_database, scraped_at, last_updated, language, status
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
		)
	`

	_, err := ss.db.ExecContext(ctx, query,
		c.ID, c.CaseNumber, c.CaseName, c.DecisionDate, c.Court, c.CourtLevel, c.CourtType,
		c.Jurisdiction, c.Docket, toJSONString(c.Parties), toJSONString(c.Judges), c.Summary, c.FullText,
		toJSONString(c.KeyIssues), toJSONString(c.LegalConcepts), c.Outcome, c.ProceduralHistory,
		toJSONString(c.Citations), c.URL, c.PDFURL, c.SourceDatabase, c.ScrapedAt, c.LastUpdated,
		c.Language, c.Status,
	)

	return err
}

// GetCase retrieves a case by ID
func (ss *SQLiteStorage) GetCase(ctx context.Context, id string) (*models.Case, error) {
	query := `
		SELECT id, case_number, case_name, decision_date, court, court_level, court_type,
			jurisdiction, docket, parties, judges, summary, full_text, key_issues,
			legal_concepts, outcome, procedural_history, citations, url, pdf_url,
			source_database, scraped_at, last_updated, language, status, created_at
		FROM cases WHERE id = ?
	`

	var c models.Case
	var partiesJSON, judgesJSON, keyIssuesJSON, legalConceptsJSON, citationsJSON sql.NullString
	var decisionDate, scrapedAt, lastUpdated, createdAt sql.NullTime

	err := ss.db.QueryRowContext(ctx, query, id).Scan(
		&c.ID, &c.CaseNumber, &c.CaseName, &decisionDate, &c.Court, &c.CourtLevel, &c.CourtType,
		&c.Jurisdiction, &c.Docket, &partiesJSON, &judgesJSON, &c.Summary, &c.FullText, &keyIssuesJSON,
		&legalConceptsJSON, &c.Outcome, &c.ProceduralHistory, &citationsJSON, &c.URL, &c.PDFURL,
		&c.SourceDatabase, &scrapedAt, &lastUpdated, &c.Language, &c.Status, &createdAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("case not found: %s", id)
		}
		return nil, err
	}

	// Parse JSON fields
	if decisionDate.Valid {
		c.DecisionDate = &decisionDate.Time
	}
	if scrapedAt.Valid {
		c.ScrapedAt = &scrapedAt.Time
	}
	if lastUpdated.Valid {
		c.LastUpdated = &lastUpdated.Time
	}
	if createdAt.Valid {
		c.CreatedAt = &createdAt.Time
	}

	if partiesJSON.Valid {
		json.Unmarshal([]byte(partiesJSON.String), &c.Parties)
	}
	if judgesJSON.Valid {
		json.Unmarshal([]byte(judgesJSON.String), &c.Judges)
	}
	if keyIssuesJSON.Valid {
		json.Unmarshal([]byte(keyIssuesJSON.String), &c.KeyIssues)
	}
	if legalConceptsJSON.Valid {
		json.Unmarshal([]byte(legalConceptsJSON.String), &c.LegalConcepts)
	}
	if citationsJSON.Valid {
		json.Unmarshal([]byte(citationsJSON.String), &c.Citations)
	}

	return &c, nil
}

// UpdateCase updates an existing case
func (ss *SQLiteStorage) UpdateCase(ctx context.Context, c *models.Case) error {
	c.LastUpdated = timePtr(time.Now())
	return ss.SaveCase(ctx, c)
}

// DeleteCase deletes a case by ID
func (ss *SQLiteStorage) DeleteCase(ctx context.Context, id string) error {
	query := `DELETE FROM cases WHERE id = ?`
	result, err := ss.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return fmt.Errorf("case not found: %s", id)
	}

	return nil
}

// ListCases lists cases with filtering
func (ss *SQLiteStorage) ListCases(ctx context.Context, filter CaseFilter) ([]*models.Case, error) {
	query := `SELECT id, case_number, case_name, decision_date, court, court_level, court_type,
		jurisdiction, docket, parties, judges, summary, full_text, key_issues,
		legal_concepts, outcome, procedural_history, citations, url, pdf_url,
		source_database, scraped_at, last_updated, language, status, created_at
		FROM cases WHERE 1=1`

	var args []interface{}
	argIndex := 1

	if filter.Jurisdiction != "" {
		query += fmt.Sprintf(" AND jurisdiction = ?%d", argIndex)
		args = append(args, filter.Jurisdiction)
		argIndex++
	}
	if filter.Court != "" {
		query += fmt.Sprintf(" AND court = ?%d", argIndex)
		args = append(args, filter.Court)
		argIndex++
	}
	if filter.CourtLevel != nil {
		query += fmt.Sprintf(" AND court_level = ?%d", argIndex)
		args = append(args, *filter.CourtLevel)
		argIndex++
	}
	if filter.Status != "" {
		query += fmt.Sprintf(" AND status = ?%d", argIndex)
		args = append(args, filter.Status)
		argIndex++
	}
	if filter.StartDate != nil {
		query += fmt.Sprintf(" AND decision_date >= ?%d", argIndex)
		args = append(args, filter.StartDate)
		argIndex++
	}
	if filter.EndDate != nil {
		query += fmt.Sprintf(" AND decision_date <= ?%d", argIndex)
		args = append(args, filter.EndDate)
		argIndex++
	}

	// Order and limit
	if filter.OrderBy != "" {
		direction := "ASC"
		if filter.OrderDesc {
			direction = "DESC"
		}
		query += fmt.Sprintf(" ORDER BY %s %s", filter.OrderBy, direction)
	} else {
		query += " ORDER BY created_at DESC"
	}

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT ?%d", argIndex)
		args = append(args, filter.Limit)
		argIndex++
	}
	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET ?%d", argIndex)
		args = append(args, filter.Offset)
		argIndex++
	}

	// Replace placeholders
	for i := 1; i <= len(args); i++ {
		query = strings.Replace(query, fmt.Sprintf("?%d", i), "?", 1)
	}

	rows, err := ss.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cases []*models.Case
	for rows.Next() {
		var c models.Case
		var partiesJSON, judgesJSON, keyIssuesJSON, legalConceptsJSON, citationsJSON sql.NullString
		var decisionDate, scrapedAt, lastUpdated, createdAt sql.NullTime

		err := rows.Scan(
			&c.ID, &c.CaseNumber, &c.CaseName, &decisionDate, &c.Court, &c.CourtLevel, &c.CourtType,
			&c.Jurisdiction, &c.Docket, &partiesJSON, &judgesJSON, &c.Summary, &c.FullText, &keyIssuesJSON,
			&legalConceptsJSON, &c.Outcome, &c.ProceduralHistory, &citationsJSON, &c.URL, &c.PDFURL,
			&c.SourceDatabase, &scrapedAt, &lastUpdated, &c.Language, &c.Status, &createdAt,
		)
		if err != nil {
			return nil, err
		}

		// Parse dates
		if decisionDate.Valid {
			c.DecisionDate = &decisionDate.Time
		}
		if scrapedAt.Valid {
			c.ScrapedAt = &scrapedAt.Time
		}
		if lastUpdated.Valid {
			c.LastUpdated = &lastUpdated.Time
		}
		if createdAt.Valid {
			c.CreatedAt = &createdAt.Time
		}

		// Parse JSON
		if partiesJSON.Valid {
			json.Unmarshal([]byte(partiesJSON.String), &c.Parties)
		}
		if judgesJSON.Valid {
			json.Unmarshal([]byte(judgesJSON.String), &c.Judges)
		}
		if keyIssuesJSON.Valid {
			json.Unmarshal([]byte(keyIssuesJSON.String), &c.KeyIssues)
		}
		if legalConceptsJSON.Valid {
			json.Unmarshal([]byte(legalConceptsJSON.String), &c.LegalConcepts)
		}
		if citationsJSON.Valid {
			json.Unmarshal([]byte(citationsJSON.String), &c.Citations)
		}

		cases = append(cases, &c)
	}

	return cases, rows.Err()
}

// CountCases counts cases matching filter
func (ss *SQLiteStorage) CountCases(ctx context.Context, filter CaseFilter) (int64, error) {
	query := `SELECT COUNT(*) FROM cases WHERE 1=1`
	var args []interface{}

	if filter.Jurisdiction != "" {
		query += " AND jurisdiction = ?"
		args = append(args, filter.Jurisdiction)
	}
	if filter.Court != "" {
		query += " AND court = ?"
		args = append(args, filter.Court)
	}
	if filter.Status != "" {
		query += " AND status = ?"
		args = append(args, filter.Status)
	}

	var count int64
	err := ss.db.QueryRowContext(ctx, query, args...).Scan(&count)
	return count, err
}

// SaveJudge saves a judge
func (ss *SQLiteStorage) SaveJudge(ctx context.Context, j *models.Judge) error {
	query := `
		INSERT OR REPLACE INTO judges (
			id, name, full_name, title, court, jurisdiction, appointed_date,
			biography, education, career, notable_cases, total_cases
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := ss.db.ExecContext(ctx, query,
		j.ID, j.Name, j.FullName, j.Title, j.Court, j.Jurisdiction, j.AppointedDate,
		j.Biography, toJSONString(j.Education), toJSONString(j.Career), toJSONString(j.NotableCases), j.TotalCases,
	)

	return err
}

// GetJudge retrieves a judge by ID
func (ss *SQLiteStorage) GetJudge(ctx context.Context, id string) (*models.Judge, error) {
	query := `
		SELECT id, name, full_name, title, court, jurisdiction, appointed_date,
			biography, education, career, notable_cases, total_cases, created_at
		FROM judges WHERE id = ?
	`

	var j models.Judge
	var appointedDate, createdAt sql.NullTime
	var educationJSON, careerJSON, notableCasesJSON sql.NullString

	err := ss.db.QueryRowContext(ctx, query, id).Scan(
		&j.ID, &j.Name, &j.FullName, &j.Title, &j.Court, &j.Jurisdiction, &appointedDate,
		&j.Biography, &educationJSON, &careerJSON, &notableCasesJSON, &j.TotalCases, &createdAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("judge not found: %s", id)
		}
		return nil, err
	}

	if appointedDate.Valid {
		j.AppointedDate = &appointedDate.Time
	}
	if createdAt.Valid {
		j.CreatedAt = &createdAt.Time
	}
	if educationJSON.Valid {
		json.Unmarshal([]byte(educationJSON.String), &j.Education)
	}
	if careerJSON.Valid {
		json.Unmarshal([]byte(careerJSON.String), &j.Career)
	}
	if notableCasesJSON.Valid {
		json.Unmarshal([]byte(notableCasesJSON.String), &j.NotableCases)
	}

	return &j, nil
}

// UpdateJudge updates a judge
func (ss *SQLiteStorage) UpdateJudge(ctx context.Context, j *models.Judge) error {
	return ss.SaveJudge(ctx, j)
}

// ListJudges lists judges with filtering
func (ss *SQLiteStorage) ListJudges(ctx context.Context, filter JudgeFilter) ([]*models.Judge, error) {
	query := `
		SELECT id, name, full_name, title, court, jurisdiction, appointed_date,
			biography, education, career, notable_cases, total_cases, created_at
		FROM judges WHERE 1=1
	`

	var args []interface{}
	if filter.Name != "" {
		query += " AND name LIKE ?"
		args = append(args, "%"+filter.Name+"%")
	}
	if filter.Court != "" {
		query += " AND court = ?"
		args = append(args, filter.Court)
	}
	if filter.Jurisdiction != "" {
		query += " AND jurisdiction = ?"
		args = append(args, filter.Jurisdiction)
	}

	query += " ORDER BY name"

	if filter.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filter.Limit)
	}
	if filter.Offset > 0 {
		query += " OFFSET ?"
		args = append(args, filter.Offset)
	}

	rows, err := ss.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var judges []*models.Judge
	for rows.Next() {
		var j models.Judge
		var appointedDate, createdAt sql.NullTime
		var educationJSON, careerJSON, notableCasesJSON sql.NullString

		err := rows.Scan(
			&j.ID, &j.Name, &j.FullName, &j.Title, &j.Court, &j.Jurisdiction, &appointedDate,
			&j.Biography, &educationJSON, &careerJSON, &notableCasesJSON, &j.TotalCases, &createdAt,
		)
		if err != nil {
			return nil, err
		}

		if appointedDate.Valid {
			j.AppointedDate = &appointedDate.Time
		}
		if createdAt.Valid {
			j.CreatedAt = &createdAt.Time
		}
		if educationJSON.Valid {
			json.Unmarshal([]byte(educationJSON.String), &j.Education)
		}
		if careerJSON.Valid {
			json.Unmarshal([]byte(careerJSON.String), &j.Career)
		}
		if notableCasesJSON.Valid {
			json.Unmarshal([]byte(notableCasesJSON.String), &j.NotableCases)
		}

		judges = append(judges, &j)
	}

	return judges, rows.Err()
}

// SaveCitation saves a citation
func (ss *SQLiteStorage) SaveCitation(ctx context.Context, c *models.Citation) error {
	query := `
		INSERT INTO citations (
			format, raw_citation, normalized_citation, volume, reporter, page, year,
			court, case_number, country, citing_case_id, cited_case_id, is_normalized
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := ss.db.ExecContext(ctx, query,
		c.Format, c.RawCitation, c.NormalizedCitation, c.Volume, c.Reporter, c.Page, c.Year,
		c.Court, c.CaseNumber, c.Country, c.CitingCaseID, c.CitedCaseID, boolToInt(c.IsNormalized),
	)

	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err == nil {
		c.ID = int(id)
	}

	return nil
}

// GetCitation retrieves a citation by ID
func (ss *SQLiteStorage) GetCitation(ctx context.Context, id string) (*models.Citation, error) {
	query := `
		SELECT id, format, raw_citation, normalized_citation, volume, reporter, page, year,
			court, case_number, country, citing_case_id, cited_case_id, is_normalized, created_at
		FROM citations WHERE id = ?
	`

	var c models.Citation
	var isNormalized int
	var createdAt sql.NullTime

	err := ss.db.QueryRowContext(ctx, query, id).Scan(
		&c.ID, &c.Format, &c.RawCitation, &c.NormalizedCitation, &c.Volume, &c.Reporter, &c.Page, &c.Year,
		&c.Court, &c.CaseNumber, &c.Country, &c.CitingCaseID, &c.CitedCaseID, &isNormalized, &createdAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("citation not found: %s", id)
		}
		return nil, err
	}

	c.IsNormalized = isNormalized == 1
	if createdAt.Valid {
		c.CreatedAt = &createdAt.Time
	}

	return &c, nil
}

// ListCitations lists citations with filtering
func (ss *SQLiteStorage) ListCitations(ctx context.Context, filter CitationFilter) ([]*models.Citation, error) {
	query := `
		SELECT id, format, raw_citation, normalized_citation, volume, reporter, page, year,
			court, case_number, country, citing_case_id, cited_case_id, is_normalized, created_at
		FROM citations WHERE 1=1
	`

	var args []interface{}
	if filter.CaseID != "" {
		query += " AND (citing_case_id = ? OR cited_case_id = ?)"
		args = append(args, filter.CaseID, filter.CaseID)
	}
	if filter.Format != "" {
		query += " AND format = ?"
		args = append(args, filter.Format)
	}
	if filter.Valid != nil {
		query += " AND is_normalized = ?"
		args = append(args, boolToInt(*filter.Valid))
	}

	query += " ORDER BY created_at DESC"

	if filter.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filter.Limit)
	}
	if filter.Offset > 0 {
		query += " OFFSET ?"
		args = append(args, filter.Offset)
	}

	rows, err := ss.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var citations []*models.Citation
	for rows.Next() {
		var c models.Citation
		var isNormalized int
		var createdAt sql.NullTime

		err := rows.Scan(
			&c.ID, &c.Format, &c.RawCitation, &c.NormalizedCitation, &c.Volume, &c.Reporter, &c.Page, &c.Year,
			&c.Court, &c.CaseNumber, &c.Country, &c.CitingCaseID, &c.CitedCaseID, &isNormalized, &createdAt,
		)
		if err != nil {
			return nil, err
		}

		c.IsNormalized = isNormalized == 1
		if createdAt.Valid {
			c.CreatedAt = &createdAt.Time
		}

		citations = append(citations, &c)
	}

	return citations, rows.Err()
}

// SearchCases performs full-text search on cases
func (ss *SQLiteStorage) SearchCases(ctx context.Context, query SearchQuery) ([]*models.Case, error) {
	// Use FTS5 for full-text search
	ftsQuery := `
		SELECT c.id, c.case_number, c.case_name, c.decision_date, c.court, c.court_level, c.court_type,
			c.jurisdiction, c.docket, c.parties, c.judges, c.summary, c.full_text, c.key_issues,
			c.legal_concepts, c.outcome, c.procedural_history, c.citations, c.url, c.pdf_url,
			c.source_database, c.scraped_at, c.last_updated, c.language, c.status, c.created_at
		FROM cases c
		JOIN cases_fts fts ON c.id = fts.id
		WHERE cases_fts MATCH ?
	`

	// Additional filters
	var args []interface{}
	args = append(args, query.Query)

	if query.Filters.Jurisdiction != "" {
		ftsQuery += " AND c.jurisdiction = ?"
		args = append(args, query.Filters.Jurisdiction)
	}
	if query.Filters.Court != "" {
		ftsQuery += " AND c.court = ?"
		args = append(args, query.Filters.Court)
	}

	ftsQuery += " ORDER BY rank"

	if query.Limit > 0 {
		ftsQuery += " LIMIT ?"
		args = append(args, query.Limit)
	}
	if query.Offset > 0 {
		ftsQuery += " OFFSET ?"
		args = append(args, query.Offset)
	}

	rows, err := ss.db.QueryContext(ctx, ftsQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cases []*models.Case
	for rows.Next() {
		var c models.Case
		var partiesJSON, judgesJSON, keyIssuesJSON, legalConceptsJSON, citationsJSON sql.NullString
		var decisionDate, scrapedAt, lastUpdated, createdAt sql.NullTime

		err := rows.Scan(
			&c.ID, &c.CaseNumber, &c.CaseName, &decisionDate, &c.Court, &c.CourtLevel, &c.CourtType,
			&c.Jurisdiction, &c.Docket, &partiesJSON, &judgesJSON, &c.Summary, &c.FullText, &keyIssuesJSON,
			&legalConceptsJSON, &c.Outcome, &c.ProceduralHistory, &citationsJSON, &c.URL, &c.PDFURL,
			&c.SourceDatabase, &scrapedAt, &lastUpdated, &c.Language, &c.Status, &createdAt,
		)
		if err != nil {
			return nil, err
		}

		// Parse dates
		if decisionDate.Valid {
			c.DecisionDate = &decisionDate.Time
		}
		if scrapedAt.Valid {
			c.ScrapedAt = &scrapedAt.Time
		}
		if lastUpdated.Valid {
			c.LastUpdated = &lastUpdated.Time
		}
		if createdAt.Valid {
			c.CreatedAt = &createdAt.Time
		}

		// Parse JSON
		if partiesJSON.Valid {
			json.Unmarshal([]byte(partiesJSON.String), &c.Parties)
		}
		if judgesJSON.Valid {
			json.Unmarshal([]byte(judgesJSON.String), &c.Judges)
		}
		if keyIssuesJSON.Valid {
			json.Unmarshal([]byte(keyIssuesJSON.String), &c.KeyIssues)
		}
		if legalConceptsJSON.Valid {
			json.Unmarshal([]byte(legalConceptsJSON.String), &c.LegalConcepts)
		}
		if citationsJSON.Valid {
			json.Unmarshal([]byte(citationsJSON.String), &c.Citations)
		}

		cases = append(cases, &c)
	}

	return cases, rows.Err()
}

// Ping checks database connectivity
func (ss *SQLiteStorage) Ping(ctx context.Context) error {
	return ss.db.PingContext(ctx)
}

// Helper functions

func toJSONString(v interface{}) string {
	if v == nil {
		return ""
	}
	b, _ := json.Marshal(v)
	return string(b)
}

func timePtr(t time.Time) *time.Time {
	return &t
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
