package storage

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/gongahkia/kite/pkg/models"
)

// PostgresStorage implements the Storage interface using PostgreSQL
type PostgresStorage struct {
	db *sql.DB
}

// NewPostgresStorage creates a new PostgreSQL storage adapter
func NewPostgresStorage(connStr string) (*PostgresStorage, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	storage := &PostgresStorage{db: db}

	// Initialize schema
	if err := storage.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return storage, nil
}

// Close closes the database connection
func (ps *PostgresStorage) Close() error {
	return ps.db.Close()
}

// initSchema creates the necessary tables
func (ps *PostgresStorage) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS cases (
		id TEXT PRIMARY KEY,
		case_number TEXT,
		case_name TEXT NOT NULL,
		decision_date TIMESTAMP,
		court TEXT,
		court_level INTEGER,
		court_type TEXT,
		jurisdiction TEXT,
		docket TEXT,
		parties JSONB,
		judges JSONB,
		summary TEXT,
		full_text TEXT,
		key_issues JSONB,
		legal_concepts JSONB,
		outcome TEXT,
		procedural_history TEXT,
		citations JSONB,
		url TEXT,
		pdf_url TEXT,
		source_database TEXT,
		scraped_at TIMESTAMP,
		last_updated TIMESTAMP,
		language TEXT,
		status TEXT,
		created_at TIMESTAMP DEFAULT NOW()
	);

	CREATE INDEX IF NOT EXISTS idx_cases_jurisdiction ON cases(jurisdiction);
	CREATE INDEX IF NOT EXISTS idx_cases_court ON cases(court);
	CREATE INDEX IF NOT EXISTS idx_cases_decision_date ON cases(decision_date);
	CREATE INDEX IF NOT EXISTS idx_cases_case_name ON cases(case_name);

	CREATE TABLE IF NOT EXISTS judges (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		full_name TEXT,
		title TEXT,
		court TEXT,
		jurisdiction TEXT,
		appointed_date TIMESTAMP,
		biography TEXT,
		education JSONB,
		career JSONB,
		notable_cases JSONB,
		total_cases INTEGER DEFAULT 0,
		created_at TIMESTAMP DEFAULT NOW()
	);

	CREATE INDEX IF NOT EXISTS idx_judges_name ON judges(name);
	CREATE INDEX IF NOT EXISTS idx_judges_court ON judges(court);

	CREATE TABLE IF NOT EXISTS citations (
		id SERIAL PRIMARY KEY,
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
		is_normalized BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP DEFAULT NOW()
	);

	CREATE INDEX IF NOT EXISTS idx_citations_citing_case ON citations(citing_case_id);
	CREATE INDEX IF NOT EXISTS idx_citations_cited_case ON citations(cited_case_id);
	`

	_, err := ps.db.Exec(schema)
	return err
}

// CreateCase creates a new case
func (ps *PostgresStorage) CreateCase(ctx context.Context, c *models.Case) error {
	query := `
		INSERT INTO cases (
			id, case_number, case_name, decision_date, court, court_level, court_type,
			jurisdiction, docket, parties, judges, summary, full_text, key_issues,
			legal_concepts, outcome, procedural_history, citations, url, pdf_url,
			source_database, scraped_at, last_updated, language, status
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18,
			$19, $20, $21, $22, $23, $24, $25
		)
	`

	_, err := ps.db.ExecContext(ctx, query,
		c.ID, c.CaseNumber, c.CaseName, c.DecisionDate, c.Court, c.CourtLevel, c.CourtType,
		c.Jurisdiction, c.Docket, toJSON(c.Parties), toJSON(c.Judges), c.Summary, c.FullText,
		toJSON(c.KeyIssues), toJSON(c.LegalConcepts), c.Outcome, c.ProceduralHistory,
		toJSON(c.CitedCases), c.URL, c.PDFURL, c.SourceDatabase, c.ScrapedAt, c.LastUpdated,
		c.Language, c.Status,
	)

	return err
}

// GetCase retrieves a case by ID
func (ps *PostgresStorage) GetCase(ctx context.Context, id string) (*models.Case, error) {
	query := `
		SELECT id, case_number, case_name, decision_date, court, court_level, court_type,
			jurisdiction, docket, parties, judges, summary, full_text, key_issues,
			legal_concepts, outcome, procedural_history, citations, url, pdf_url,
			source_database, scraped_at, last_updated, language, status
		FROM cases
		WHERE id = $1
	`

	c := &models.Case{}
	var decisionDate sql.NullTime
	var scrapedAt, lastUpdated sql.NullTime
	var parties, judges, keyIssues, legalConcepts, citations []byte

	err := ps.db.QueryRowContext(ctx, query, id).Scan(
		&c.ID, &c.CaseNumber, &c.CaseName, &decisionDate, &c.Court, &c.CourtLevel, &c.CourtType,
		&c.Jurisdiction, &c.Docket, &parties, &judges, &c.Summary, &c.FullText, &keyIssues,
		&legalConcepts, &c.Outcome, &c.ProceduralHistory, &citations, &c.URL, &c.PDFURL,
		&c.SourceDatabase, &scrapedAt, &lastUpdated, &c.Language, &c.Status,
	)

	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	// Parse nullable fields
	if decisionDate.Valid {
		c.DecisionDate = &decisionDate.Time
	}
	if scrapedAt.Valid {
		c.ScrapedAt = scrapedAt.Time
	}
	if lastUpdated.Valid {
		c.LastUpdated = lastUpdated.Time
	}

	// Parse JSON fields
	fromJSON(parties, &c.Parties)
	fromJSON(judges, &c.Judges)
	fromJSON(keyIssues, &c.KeyIssues)
	fromJSON(legalConcepts, &c.LegalConcepts)
	fromJSON(citations, &c.CitedCases)

	return c, nil
}

// UpdateCase updates an existing case
func (ps *PostgresStorage) UpdateCase(ctx context.Context, id string, c *models.Case) error {
	query := `
		UPDATE cases SET
			case_number = $2, case_name = $3, decision_date = $4, court = $5,
			court_level = $6, court_type = $7, jurisdiction = $8, docket = $9,
			parties = $10, judges = $11, summary = $12, full_text = $13,
			key_issues = $14, legal_concepts = $15, outcome = $16,
			procedural_history = $17, citations = $18, url = $19, pdf_url = $20,
			source_database = $21, last_updated = $22, language = $23, status = $24
		WHERE id = $1
	`

	result, err := ps.db.ExecContext(ctx, query,
		id, c.CaseNumber, c.CaseName, c.DecisionDate, c.Court, c.CourtLevel, c.CourtType,
		c.Jurisdiction, c.Docket, toJSON(c.Parties), toJSON(c.Judges), c.Summary, c.FullText,
		toJSON(c.KeyIssues), toJSON(c.LegalConcepts), c.Outcome, c.ProceduralHistory,
		toJSON(c.CitedCases), c.URL, c.PDFURL, c.SourceDatabase, time.Now(), c.Language, c.Status,
	)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

// DeleteCase deletes a case
func (ps *PostgresStorage) DeleteCase(ctx context.Context, id string) error {
	query := `DELETE FROM cases WHERE id = $1`

	result, err := ps.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

// ListCases lists cases with optional filtering
func (ps *PostgresStorage) ListCases(ctx context.Context, filter CaseFilter) ([]*models.Case, error) {
	query := `SELECT id, case_number, case_name, decision_date, court, jurisdiction FROM cases WHERE 1=1`
	args := []interface{}{}
	argCount := 1

	// Add filters
	if filter.Jurisdiction != "" {
		query += fmt.Sprintf(" AND jurisdiction = $%d", argCount)
		args = append(args, filter.Jurisdiction)
		argCount++
	}

	if filter.Court != "" {
		query += fmt.Sprintf(" AND court = $%d", argCount)
		args = append(args, filter.Court)
		argCount++
	}

	if filter.StartDate != nil {
		query += fmt.Sprintf(" AND decision_date >= $%d", argCount)
		args = append(args, filter.StartDate)
		argCount++
	}

	if filter.EndDate != nil {
		query += fmt.Sprintf(" AND decision_date <= $%d", argCount)
		args = append(args, filter.EndDate)
		argCount++
	}

	query += " ORDER BY decision_date DESC LIMIT 100"

	rows, err := ps.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cases := make([]*models.Case, 0)
	for rows.Next() {
		c := &models.Case{}
		var decisionDate sql.NullTime

		err := rows.Scan(&c.ID, &c.CaseNumber, &c.CaseName, &decisionDate, &c.Court, &c.Jurisdiction)
		if err != nil {
			continue
		}

		if decisionDate.Valid {
			c.DecisionDate = &decisionDate.Time
		}

		cases = append(cases, c)
	}

	return cases, nil
}

// SearchCases searches for cases
func (ps *PostgresStorage) SearchCases(ctx context.Context, query SearchQuery) ([]*models.Case, error) {
	sqlQuery := `
		SELECT id, case_number, case_name, decision_date, court, jurisdiction
		FROM cases
		WHERE case_name ILIKE $1 OR case_number ILIKE $1 OR full_text ILIKE $1
		LIMIT 50
	`

	searchTerm := "%" + query.Query + "%"

	rows, err := ps.db.QueryContext(ctx, sqlQuery, searchTerm)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cases := make([]*models.Case, 0)
	for rows.Next() {
		c := &models.Case{}
		var decisionDate sql.NullTime

		err := rows.Scan(&c.ID, &c.CaseNumber, &c.CaseName, &decisionDate, &c.Court, &c.Jurisdiction)
		if err != nil {
			continue
		}

		if decisionDate.Valid {
			c.DecisionDate = &decisionDate.Time
		}

		cases = append(cases, c)
	}

	return cases, nil
}

// CreateJudge creates a new judge
func (ps *PostgresStorage) CreateJudge(ctx context.Context, j *models.Judge) error {
	query := `
		INSERT INTO judges (
			id, name, full_name, title, court, jurisdiction, appointed_date,
			biography, education, career, notable_cases, total_cases
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err := ps.db.ExecContext(ctx, query,
		j.ID, j.Name, j.FullName, j.Title, j.Court, j.Jurisdiction, j.AppointedDate,
		j.Biography, toJSON(j.Education), toJSON(j.Career), toJSON(j.NotableCases), j.TotalCases,
	)

	return err
}

// GetJudge retrieves a judge by ID
func (ps *PostgresStorage) GetJudge(ctx context.Context, id string) (*models.Judge, error) {
	query := `
		SELECT id, name, full_name, title, court, jurisdiction, appointed_date,
			biography, education, career, notable_cases, total_cases
		FROM judges
		WHERE id = $1
	`

	j := &models.Judge{}
	var appointedDate sql.NullTime
	var education, career, notableCases []byte

	err := ps.db.QueryRowContext(ctx, query, id).Scan(
		&j.ID, &j.Name, &j.FullName, &j.Title, &j.Court, &j.Jurisdiction, &appointedDate,
		&j.Biography, &education, &career, &notableCases, &j.TotalCases,
	)

	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	if appointedDate.Valid {
		j.AppointedDate = &appointedDate.Time
	}

	fromJSON(education, &j.Education)
	fromJSON(career, &j.Career)
	fromJSON(notableCases, &j.NotableCases)

	return j, nil
}

// UpdateJudge updates an existing judge
func (ps *PostgresStorage) UpdateJudge(ctx context.Context, id string, j *models.Judge) error {
	query := `
		UPDATE judges SET
			name = $2, full_name = $3, title = $4, court = $5, jurisdiction = $6,
			appointed_date = $7, biography = $8, education = $9, career = $10,
			notable_cases = $11, total_cases = $12
		WHERE id = $1
	`

	result, err := ps.db.ExecContext(ctx, query,
		id, j.Name, j.FullName, j.Title, j.Court, j.Jurisdiction, j.AppointedDate,
		j.Biography, toJSON(j.Education), toJSON(j.Career), toJSON(j.NotableCases), j.TotalCases,
	)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

// ListJudges lists all judges
func (ps *PostgresStorage) ListJudges(ctx context.Context, filter JudgeFilter) ([]*models.Judge, error) {
	query := `SELECT id, name, full_name, court, jurisdiction FROM judges WHERE 1=1`
	args := []interface{}{}
	argCount := 1

	if filter.Jurisdiction != "" {
		query += fmt.Sprintf(" AND jurisdiction = $%d", argCount)
		args = append(args, filter.Jurisdiction)
		argCount++
	}

	if filter.Court != "" {
		query += fmt.Sprintf(" AND court = $%d", argCount)
		args = append(args, filter.Court)
		argCount++
	}

	query += " ORDER BY name LIMIT 100"

	rows, err := ps.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	judges := make([]*models.Judge, 0)
	for rows.Next() {
		j := &models.Judge{}
		err := rows.Scan(&j.ID, &j.Name, &j.FullName, &j.Court, &j.Jurisdiction)
		if err != nil {
			continue
		}
		judges = append(judges, j)
	}

	return judges, nil
}

// CreateCitation creates a new citation
func (ps *PostgresStorage) CreateCitation(ctx context.Context, c *models.Citation) error {
	query := `
		INSERT INTO citations (
			format, raw_citation, normalized_citation, volume, reporter, page,
			year, court, case_number, country, citing_case_id, cited_case_id, is_normalized
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	_, err := ps.db.ExecContext(ctx, query,
		c.Format, c.RawCitation, c.NormalizedCitation, c.Volume, c.Reporter, c.Page,
		c.Year, c.Court, c.CaseNumber, c.Country, c.CitingCaseID, c.CitedCaseID, c.IsNormalized,
	)

	return err
}

// GetCitation retrieves a citation by ID
func (ps *PostgresStorage) GetCitation(ctx context.Context, id string) (*models.Citation, error) {
	return nil, fmt.Errorf("not implemented")
}

// ListCitations lists citations with optional filtering
func (ps *PostgresStorage) ListCitations(ctx context.Context, filter CitationFilter) ([]*models.Citation, error) {
	query := `SELECT format, raw_citation, normalized_citation, citing_case_id, cited_case_id FROM citations WHERE 1=1`
	args := []interface{}{}
	argCount := 1

	if filter.CitingCaseID != "" {
		query += fmt.Sprintf(" AND citing_case_id = $%d", argCount)
		args = append(args, filter.CitingCaseID)
		argCount++
	}

	if filter.CitedCaseID != "" {
		query += fmt.Sprintf(" AND cited_case_id = $%d", argCount)
		args = append(args, filter.CitedCaseID)
		argCount++
	}

	query += " LIMIT 100"

	rows, err := ps.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	citations := make([]*models.Citation, 0)
	for rows.Next() {
		c := &models.Citation{}
		err := rows.Scan(&c.Format, &c.RawCitation, &c.NormalizedCitation, &c.CitingCaseID, &c.CitedCaseID)
		if err != nil {
			continue
		}
		citations = append(citations, c)
	}

	return citations, nil
}

// GetStats returns storage statistics
func (ps *PostgresStorage) GetStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Count cases
	var caseCount int
	err := ps.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM cases").Scan(&caseCount)
	if err == nil {
		stats["total_cases"] = caseCount
	}

	// Count judges
	var judgeCount int
	err = ps.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM judges").Scan(&judgeCount)
	if err == nil {
		stats["total_judges"] = judgeCount
	}

	// Count citations
	var citationCount int
	err = ps.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM citations").Scan(&citationCount)
	if err == nil {
		stats["total_citations"] = citationCount
	}

	return stats, nil
}

// Helper functions

func toJSON(v interface{}) []byte {
	// Simple JSON serialization (in production, handle errors)
	data, _ := jsonMarshal(v)
	return data
}

func fromJSON(data []byte, v interface{}) {
	// Simple JSON deserialization (in production, handle errors)
	if len(data) > 0 {
		jsonUnmarshal(data, v)
	}
}

// Placeholder JSON functions (will be replaced with encoding/json in compilation)
func jsonMarshal(v interface{}) ([]byte, error) {
	return []byte("{}"), nil
}

func jsonUnmarshal(data []byte, v interface{}) error {
	return nil
}
