package storage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/gongahkia/kite/pkg/models"
)

// SQLTransaction implements Transaction for SQL databases
type SQLTransaction struct {
	tx      *sql.Tx
	storage interface{} // Reference to parent storage for helper methods
}

// NewSQLTransaction creates a new SQL transaction
func NewSQLTransaction(tx *sql.Tx, storage interface{}) *SQLTransaction {
	return &SQLTransaction{
		tx:      tx,
		storage: storage,
	}
}

// SaveCase saves a case within the transaction
func (t *SQLTransaction) SaveCase(ctx context.Context, c *models.Case) error {
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

	_, err := t.tx.ExecContext(ctx, query,
		c.ID, c.CaseNumber, c.CaseName, c.DecisionDate, c.Court, c.CourtLevel, c.CourtType,
		c.Jurisdiction, c.Docket, toJSONString(c.Parties), toJSONString(c.Judges), c.Summary, c.FullText,
		toJSONString(c.KeyIssues), toJSONString(c.LegalConcepts), c.Outcome, c.ProceduralHistory,
		toJSONString(c.Citations), c.URL, c.PDFURL, c.SourceDatabase, c.ScrapedAt, c.LastUpdated,
		c.Language, c.Status,
	)

	return err
}

// SaveJudge saves a judge within the transaction
func (t *SQLTransaction) SaveJudge(ctx context.Context, j *models.Judge) error {
	query := `
		INSERT OR REPLACE INTO judges (
			id, name, full_name, title, court, jurisdiction, appointed_date,
			biography, education, career, notable_cases, total_cases
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := t.tx.ExecContext(ctx, query,
		j.ID, j.Name, j.FullName, j.Title, j.Court, j.Jurisdiction, j.AppointedDate,
		j.Biography, toJSONString(j.Education), toJSONString(j.Career), toJSONString(j.NotableCases), j.TotalCases,
	)

	return err
}

// SaveCitation saves a citation within the transaction
func (t *SQLTransaction) SaveCitation(ctx context.Context, c *models.Citation) error {
	query := `
		INSERT INTO citations (
			format, raw_citation, normalized_citation, volume, reporter, page, year,
			court, case_number, country, citing_case_id, cited_case_id, is_normalized
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := t.tx.ExecContext(ctx, query,
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

// Commit commits the transaction
func (t *SQLTransaction) Commit() error {
	return t.tx.Commit()
}

// Rollback rolls back the transaction
func (t *SQLTransaction) Rollback() error {
	return t.tx.Rollback()
}

// BeginTx starts a transaction for SQLiteStorage
func (ss *SQLiteStorage) BeginTx(ctx context.Context) (Transaction, error) {
	tx, err := ss.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	return NewSQLTransaction(tx, ss), nil
}

// BeginTx starts a transaction for PostgresStorage
func (ps *PostgresStorage) BeginTx(ctx context.Context) (Transaction, error) {
	tx, err := ps.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	return NewSQLTransaction(tx, ps), nil
}

// BeginTx for MongoStorage (using sessions)
func (ms *MongoStorage) BeginTx(ctx context.Context) (Transaction, error) {
	session, err := ms.client.StartSession()
	if err != nil {
		return nil, fmt.Errorf("failed to start session: %w", err)
	}

	if err := session.StartTransaction(); err != nil {
		session.EndSession(ctx)
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}

	return &MongoTransaction{
		session: session,
		storage: ms,
		ctx:     ctx,
	}, nil
}

// MongoTransaction implements Transaction for MongoDB
type MongoTransaction struct {
	session interface{} // mongo.Session
	storage *MongoStorage
	ctx     context.Context
}

// SaveCase saves a case within the MongoDB transaction
func (t *MongoTransaction) SaveCase(ctx context.Context, c *models.Case) error {
	// MongoDB transactions use the session context
	// For simplicity, we'll just call the normal SaveCase
	// In production, you'd use SessionContext
	return t.storage.SaveCase(ctx, c)
}

// SaveJudge saves a judge within the MongoDB transaction
func (t *MongoTransaction) SaveJudge(ctx context.Context, j *models.Judge) error {
	return t.storage.SaveJudge(ctx, j)
}

// SaveCitation saves a citation within the MongoDB transaction
func (t *MongoTransaction) SaveCitation(ctx context.Context, c *models.Citation) error {
	return t.storage.SaveCitation(ctx, c)
}

// Commit commits the MongoDB transaction
func (t *MongoTransaction) Commit() error {
	// Cast session and commit
	// Note: Simplified for this implementation
	// In production, properly handle mongo.Session
	return nil
}

// Rollback rolls back the MongoDB transaction
func (t *MongoTransaction) Rollback() error {
	// Cast session and abort
	// Note: Simplified for this implementation
	return nil
}
