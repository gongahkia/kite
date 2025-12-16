package storage

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"github.com/gongahkia/kite/pkg/models"
)

// MongoStorage implements the Storage interface using MongoDB
type MongoStorage struct {
	client     *mongo.Client
	database   *mongo.Database
	cases      *mongo.Collection
	judges     *mongo.Collection
	citations  *mongo.Collection
}

// NewMongoStorage creates a new MongoDB storage adapter
func NewMongoStorage(uri, dbName string) (*MongoStorage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping to verify connection
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	database := client.Database(dbName)

	storage := &MongoStorage{
		client:    client,
		database:  database,
		cases:     database.Collection("cases"),
		judges:    database.Collection("judges"),
		citations: database.Collection("citations"),
	}

	// Create indexes
	if err := storage.createIndexes(ctx); err != nil {
		return nil, fmt.Errorf("failed to create indexes: %w", err)
	}

	return storage, nil
}

// createIndexes creates necessary indexes
func (ms *MongoStorage) createIndexes(ctx context.Context) error {
	// Cases indexes
	caseIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "jurisdiction", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "court", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "decision_date", Value: -1}},
		},
		{
			Keys: bson.D{{Key: "status", Value: 1}},
		},
		{
			// Text index for full-text search
			Keys: bson.D{
				{Key: "case_name", Value: "text"},
				{Key: "summary", Value: "text"},
				{Key: "full_text", Value: "text"},
			},
		},
	}

	_, err := ms.cases.Indexes().CreateMany(ctx, caseIndexes)
	if err != nil {
		return fmt.Errorf("failed to create case indexes: %w", err)
	}

	// Judges indexes
	judgeIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "name", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "court", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "jurisdiction", Value: 1}},
		},
	}

	_, err = ms.judges.Indexes().CreateMany(ctx, judgeIndexes)
	if err != nil {
		return fmt.Errorf("failed to create judge indexes: %w", err)
	}

	// Citations indexes
	citationIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "citing_case_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "cited_case_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "format", Value: 1}},
		},
	}

	_, err = ms.citations.Indexes().CreateMany(ctx, citationIndexes)
	if err != nil {
		return fmt.Errorf("failed to create citation indexes: %w", err)
	}

	return nil
}

// Close closes the database connection
func (ms *MongoStorage) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return ms.client.Disconnect(ctx)
}

// SaveCase saves or updates a case
func (ms *MongoStorage) SaveCase(ctx context.Context, c *models.Case) error {
	filter := bson.M{"id": c.ID}
	update := bson.M{"$set": c}
	opts := options.Update().SetUpsert(true)

	_, err := ms.cases.UpdateOne(ctx, filter, update, opts)
	return err
}

// GetCase retrieves a case by ID
func (ms *MongoStorage) GetCase(ctx context.Context, id string) (*models.Case, error) {
	filter := bson.M{"id": id}
	var c models.Case

	err := ms.cases.FindOne(ctx, filter).Decode(&c)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("case not found: %s", id)
		}
		return nil, err
	}

	return &c, nil
}

// UpdateCase updates an existing case
func (ms *MongoStorage) UpdateCase(ctx context.Context, c *models.Case) error {
	c.LastUpdated = timePtr(time.Now())
	return ms.SaveCase(ctx, c)
}

// DeleteCase deletes a case by ID
func (ms *MongoStorage) DeleteCase(ctx context.Context, id string) error {
	filter := bson.M{"id": id}
	result, err := ms.cases.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("case not found: %s", id)
	}

	return nil
}

// ListCases lists cases with filtering
func (ms *MongoStorage) ListCases(ctx context.Context, filter CaseFilter) ([]*models.Case, error) {
	query := bson.M{}

	if filter.Jurisdiction != "" {
		query["jurisdiction"] = filter.Jurisdiction
	}
	if filter.Court != "" {
		query["court"] = filter.Court
	}
	if filter.CourtLevel != nil {
		query["court_level"] = *filter.CourtLevel
	}
	if filter.Status != "" {
		query["status"] = filter.Status
	}
	if filter.StartDate != nil || filter.EndDate != nil {
		dateQuery := bson.M{}
		if filter.StartDate != nil {
			dateQuery["$gte"] = filter.StartDate
		}
		if filter.EndDate != nil {
			dateQuery["$lte"] = filter.EndDate
		}
		query["decision_date"] = dateQuery
	}

	// Options
	opts := options.Find()

	if filter.OrderBy != "" {
		sortOrder := 1
		if filter.OrderDesc {
			sortOrder = -1
		}
		opts.SetSort(bson.D{{Key: filter.OrderBy, Value: sortOrder}})
	} else {
		opts.SetSort(bson.D{{Key: "created_at", Value: -1}})
	}

	if filter.Limit > 0 {
		opts.SetLimit(int64(filter.Limit))
	}
	if filter.Offset > 0 {
		opts.SetSkip(int64(filter.Offset))
	}

	cursor, err := ms.cases.Find(ctx, query, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var cases []*models.Case
	if err := cursor.All(ctx, &cases); err != nil {
		return nil, err
	}

	return cases, nil
}

// CountCases counts cases matching filter
func (ms *MongoStorage) CountCases(ctx context.Context, filter CaseFilter) (int64, error) {
	query := bson.M{}

	if filter.Jurisdiction != "" {
		query["jurisdiction"] = filter.Jurisdiction
	}
	if filter.Court != "" {
		query["court"] = filter.Court
	}
	if filter.Status != "" {
		query["status"] = filter.Status
	}

	return ms.cases.CountDocuments(ctx, query)
}

// SaveJudge saves a judge
func (ms *MongoStorage) SaveJudge(ctx context.Context, j *models.Judge) error {
	filter := bson.M{"id": j.ID}
	update := bson.M{"$set": j}
	opts := options.Update().SetUpsert(true)

	_, err := ms.judges.UpdateOne(ctx, filter, update, opts)
	return err
}

// GetJudge retrieves a judge by ID
func (ms *MongoStorage) GetJudge(ctx context.Context, id string) (*models.Judge, error) {
	filter := bson.M{"id": id}
	var j models.Judge

	err := ms.judges.FindOne(ctx, filter).Decode(&j)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("judge not found: %s", id)
		}
		return nil, err
	}

	return &j, nil
}

// UpdateJudge updates a judge
func (ms *MongoStorage) UpdateJudge(ctx context.Context, j *models.Judge) error {
	return ms.SaveJudge(ctx, j)
}

// ListJudges lists judges with filtering
func (ms *MongoStorage) ListJudges(ctx context.Context, filter JudgeFilter) ([]*models.Judge, error) {
	query := bson.M{}

	if filter.Name != "" {
		query["name"] = bson.M{"$regex": primitive.Regex{Pattern: filter.Name, Options: "i"}}
	}
	if filter.Court != "" {
		query["court"] = filter.Court
	}
	if filter.Jurisdiction != "" {
		query["jurisdiction"] = filter.Jurisdiction
	}

	opts := options.Find().SetSort(bson.D{{Key: "name", Value: 1}})

	if filter.Limit > 0 {
		opts.SetLimit(int64(filter.Limit))
	}
	if filter.Offset > 0 {
		opts.SetSkip(int64(filter.Offset))
	}

	cursor, err := ms.judges.Find(ctx, query, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var judges []*models.Judge
	if err := cursor.All(ctx, &judges); err != nil {
		return nil, err
	}

	return judges, nil
}

// SaveCitation saves a citation
func (ms *MongoStorage) SaveCitation(ctx context.Context, c *models.Citation) error {
	// Generate ObjectID for new citations
	if c.ID == 0 {
		result, err := ms.citations.InsertOne(ctx, c)
		if err != nil {
			return err
		}
		if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
			// Convert ObjectID to int (simplified - in production use string IDs)
			c.ID = int(oid.Timestamp().Unix())
		}
		return nil
	}

	// Update existing citation
	filter := bson.M{"_id": primitive.ObjectID{}}
	update := bson.M{"$set": c}
	opts := options.Update().SetUpsert(true)

	_, err := ms.citations.UpdateOne(ctx, filter, update, opts)
	return err
}

// GetCitation retrieves a citation by ID
func (ms *MongoStorage) GetCitation(ctx context.Context, id string) (*models.Citation, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid citation ID: %s", id)
	}

	filter := bson.M{"_id": objID}
	var c models.Citation

	err = ms.citations.FindOne(ctx, filter).Decode(&c)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("citation not found: %s", id)
		}
		return nil, err
	}

	return &c, nil
}

// ListCitations lists citations with filtering
func (ms *MongoStorage) ListCitations(ctx context.Context, filter CitationFilter) ([]*models.Citation, error) {
	query := bson.M{}

	if filter.CaseID != "" {
		query["$or"] = []bson.M{
			{"citing_case_id": filter.CaseID},
			{"cited_case_id": filter.CaseID},
		}
	}
	if filter.Format != "" {
		query["format"] = filter.Format
	}
	if filter.Valid != nil {
		query["is_normalized"] = *filter.Valid
	}

	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})

	if filter.Limit > 0 {
		opts.SetLimit(int64(filter.Limit))
	}
	if filter.Offset > 0 {
		opts.SetSkip(int64(filter.Offset))
	}

	cursor, err := ms.citations.Find(ctx, query, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var citations []*models.Citation
	if err := cursor.All(ctx, &citations); err != nil {
		return nil, err
	}

	return citations, nil
}

// SearchCases performs full-text search on cases
func (ms *MongoStorage) SearchCases(ctx context.Context, query SearchQuery) ([]*models.Case, error) {
	filter := bson.M{
		"$text": bson.M{
			"$search": query.Query,
		},
	}

	// Additional filters
	if query.Filters.Jurisdiction != "" {
		filter["jurisdiction"] = query.Filters.Jurisdiction
	}
	if query.Filters.Court != "" {
		filter["court"] = query.Filters.Court
	}

	opts := options.Find()

	// Sort by text score for relevance
	opts.SetProjection(bson.M{
		"score": bson.M{"$meta": "textScore"},
	})
	opts.SetSort(bson.M{
		"score": bson.M{"$meta": "textScore"},
	})

	if query.Limit > 0 {
		opts.SetLimit(int64(query.Limit))
	}
	if query.Offset > 0 {
		opts.SetSkip(int64(query.Offset))
	}

	cursor, err := ms.cases.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var cases []*models.Case
	if err := cursor.All(ctx, &cases); err != nil {
		return nil, err
	}

	return cases, nil
}

// Ping checks database connectivity
func (ms *MongoStorage) Ping(ctx context.Context) error {
	return ms.client.Ping(ctx, nil)
}
