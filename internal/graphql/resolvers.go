package graphql

import (
	"context"

	"github.com/gongahkia/kite/internal/storage"
	"github.com/gongahkia/kite/pkg/models"
	"github.com/graphql-go/graphql"
)

// Resolver holds dependencies for GraphQL resolvers
type Resolver struct {
	storage storage.Storage
}

// NewResolver creates a new resolver
func NewResolver(storage storage.Storage) *Resolver {
	return &Resolver{
		storage: storage,
	}
}

// GetCaseResolver resolves a single case by ID
func (r *Resolver) GetCaseResolver(params graphql.ResolveParams) (interface{}, error) {
	id, ok := params.Args["id"].(string)
	if !ok {
		return nil, nil
	}

	ctx := params.Context
	return r.storage.GetCase(ctx, id)
}

// SearchCasesResolver resolves case search queries
func (r *Resolver) SearchCasesResolver(params graphql.ResolveParams) (interface{}, error) {
	ctx := params.Context

	// Parse arguments
	query, _ := params.Args["query"].(string)
	limit, _ := params.Args["limit"].(int)
	offset, _ := params.Args["offset"].(int)

	// Parse filters
	var filters storage.SearchFilters
	if filterMap, ok := params.Args["filter"].(map[string]interface{}); ok {
		if jurisdiction, ok := filterMap["jurisdiction"].(string); ok {
			filters.Jurisdiction = jurisdiction
		}
		if court, ok := filterMap["court"].(string); ok {
			filters.Court = court
		}
		if topics, ok := filterMap["legalTopics"].([]interface{}); ok {
			for _, t := range topics {
				if topic, ok := t.(string); ok {
					filters.LegalTopics = append(filters.LegalTopics, topic)
				}
			}
		}
		if minScore, ok := filterMap["minQualityScore"].(float64); ok {
			filters.MinQualityScore = minScore
		}
	}

	// Set defaults
	if limit == 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	// Search cases
	cases, total, err := r.storage.SearchCases(ctx, query, filters, limit, offset)
	if err != nil {
		return nil, err
	}

	// Return search result
	return map[string]interface{}{
		"cases":   cases,
		"total":   total,
		"offset":  offset,
		"limit":   limit,
		"hasMore": (offset + len(cases)) < total,
	}, nil
}

// GetCasesByJurisdictionResolver resolves cases by jurisdiction
func (r *Resolver) GetCasesByJurisdictionResolver(params graphql.ResolveParams) (interface{}, error) {
	ctx := params.Context

	jurisdiction, ok := params.Args["jurisdiction"].(string)
	if !ok {
		return nil, nil
	}

	limit, _ := params.Args["limit"].(int)
	offset, _ := params.Args["offset"].(int)

	if limit == 0 {
		limit = 20
	}

	filters := storage.SearchFilters{
		Jurisdiction: jurisdiction,
	}

	cases, _, err := r.storage.SearchCases(ctx, "", filters, limit, offset)
	return cases, err
}

// GetStatsResolver resolves system statistics
func (r *Resolver) GetStatsResolver(params graphql.ResolveParams) (interface{}, error) {
	ctx := params.Context

	// Get total cases
	total, err := r.storage.GetTotalCases(ctx)
	if err != nil {
		return nil, err
	}

	// Get jurisdiction stats
	jurisdictionStats, err := r.storage.GetJurisdictionStats(ctx)
	if err != nil {
		return nil, err
	}

	// Calculate percentages
	var statsWithPercentages []map[string]interface{}
	for _, stat := range jurisdictionStats {
		percentage := float64(0)
		if total > 0 {
			percentage = float64(stat["count"].(int)) / float64(total) * 100
		}
		statsWithPercentages = append(statsWithPercentages, map[string]interface{}{
			"jurisdiction": stat["jurisdiction"],
			"count":        stat["count"],
			"percentage":   percentage,
		})
	}

	return map[string]interface{}{
		"totalCases":           total,
		"casesByJurisdiction":  statsWithPercentages,
		"averageQualityScore":  r.storage.GetAverageQualityScore(ctx),
	}, nil
}

// CreateCaseResolver creates a new case
func (r *Resolver) CreateCaseResolver(params graphql.ResolveParams) (interface{}, error) {
	ctx := params.Context

	input, ok := params.Args["input"].(map[string]interface{})
	if !ok {
		return nil, nil
	}

	// Build case from input
	c := &models.Case{
		CaseName:     input["caseName"].(string),
		CaseNumber:   input["caseNumber"].(string),
		Court:        input["court"].(string),
		Jurisdiction: input["jurisdiction"].(string),
	}

	// Optional fields
	if summary, ok := input["summary"].(string); ok {
		c.Summary = summary
	}
	if fullText, ok := input["fullText"].(string); ok {
		c.FullText = fullText
	}
	if url, ok := input["url"].(string); ok {
		c.URL = url
	}

	// Lists
	if judges, ok := input["judges"].([]interface{}); ok {
		for _, j := range judges {
			if judge, ok := j.(string); ok {
				c.Judges = append(c.Judges, judge)
			}
		}
	}
	if parties, ok := input["parties"].([]interface{}); ok {
		for _, p := range parties {
			if party, ok := p.(string); ok {
				c.Parties = append(c.Parties, party)
			}
		}
	}
	if topics, ok := input["legalTopics"].([]interface{}); ok {
		for _, t := range topics {
			if topic, ok := t.(string); ok {
				c.LegalTopics = append(c.LegalTopics, topic)
			}
		}
	}

	// Create case
	if err := r.storage.CreateCase(ctx, c); err != nil {
		return nil, err
	}

	return c, nil
}

// UpdateCaseResolver updates an existing case
func (r *Resolver) UpdateCaseResolver(params graphql.ResolveParams) (interface{}, error) {
	ctx := params.Context

	id, ok := params.Args["id"].(string)
	if !ok {
		return nil, nil
	}

	input, ok := params.Args["input"].(map[string]interface{})
	if !ok {
		return nil, nil
	}

	// Get existing case
	c, err := r.storage.GetCase(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update fields
	if caseName, ok := input["caseName"].(string); ok {
		c.CaseName = caseName
	}
	if summary, ok := input["summary"].(string); ok {
		c.Summary = summary
	}
	if fullText, ok := input["fullText"].(string); ok {
		c.FullText = fullText
	}

	// Update case
	if err := r.storage.UpdateCase(ctx, c); err != nil {
		return nil, err
	}

	return c, nil
}

// DeleteCaseResolver deletes a case
func (r *Resolver) DeleteCaseResolver(params graphql.ResolveParams) (interface{}, error) {
	ctx := params.Context

	id, ok := params.Args["id"].(string)
	if !ok {
		return false, nil
	}

	if err := r.storage.DeleteCase(ctx, id); err != nil {
		return false, err
	}

	return true, nil
}

// BuildSchema builds the complete GraphQL schema
func BuildSchema(resolver *Resolver) (graphql.Schema, error) {
	// Query type
	queryType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"case": &graphql.Field{
				Type: CaseType,
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
				},
				Resolve: resolver.GetCaseResolver,
			},
			"searchCases": &graphql.Field{
				Type: SearchResultType,
				Args: graphql.FieldConfigArgument{
					"query": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
					"filter": &graphql.ArgumentConfig{
						Type: FilterInputType,
					},
					"limit": &graphql.ArgumentConfig{
						Type:         graphql.Int,
						DefaultValue: 20,
					},
					"offset": &graphql.ArgumentConfig{
						Type:         graphql.Int,
						DefaultValue: 0,
					},
				},
				Resolve: resolver.SearchCasesResolver,
			},
			"casesByJurisdiction": &graphql.Field{
				Type: graphql.NewList(CaseType),
				Args: graphql.FieldConfigArgument{
					"jurisdiction": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
					"limit": &graphql.ArgumentConfig{
						Type:         graphql.Int,
						DefaultValue: 20,
					},
					"offset": &graphql.ArgumentConfig{
						Type:         graphql.Int,
						DefaultValue: 0,
					},
				},
				Resolve: resolver.GetCasesByJurisdictionResolver,
			},
			"stats": &graphql.Field{
				Type:    StatsType,
				Resolve: resolver.GetStatsResolver,
			},
		},
	})

	// Mutation type
	mutationType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Mutation",
		Fields: graphql.Fields{
			"createCase": &graphql.Field{
				Type: CaseType,
				Args: graphql.FieldConfigArgument{
					"input": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(CreateCaseInputType),
					},
				},
				Resolve: resolver.CreateCaseResolver,
			},
			"updateCase": &graphql.Field{
				Type: CaseType,
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
					"input": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(UpdateCaseInputType),
					},
				},
				Resolve: resolver.UpdateCaseResolver,
			},
			"deleteCase": &graphql.Field{
				Type: graphql.Boolean,
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
				},
				Resolve: resolver.DeleteCaseResolver,
			},
		},
	})

	// Build schema
	return graphql.NewSchema(graphql.SchemaConfig{
		Query:    queryType,
		Mutation: mutationType,
	})
}

// ExecuteQuery executes a GraphQL query
func ExecuteQuery(schema graphql.Schema, query string, variables map[string]interface{}, ctx context.Context) *graphql.Result {
	result := graphql.Do(graphql.Params{
		Schema:         schema,
		RequestString:  query,
		VariableValues: variables,
		Context:        ctx,
	})

	return result
}
