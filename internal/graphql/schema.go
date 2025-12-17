package graphql

import (
	"github.com/graphql-go/graphql"
)

// Schema defines the GraphQL schema
var Schema graphql.Schema

// CaseType represents a legal case in GraphQL
var CaseType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Case",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type: graphql.String,
		},
		"caseName": &graphql.Field{
			Type: graphql.String,
		},
		"caseNumber": &graphql.Field{
			Type: graphql.String,
		},
		"court": &graphql.Field{
			Type: graphql.String,
		},
		"jurisdiction": &graphql.Field{
			Type: graphql.String,
		},
		"decisionDate": &graphql.Field{
			Type: graphql.DateTime,
		},
		"summary": &graphql.Field{
			Type: graphql.String,
		},
		"fullText": &graphql.Field{
			Type: graphql.String,
		},
		"judges": &graphql.Field{
			Type: graphql.NewList(graphql.String),
		},
		"parties": &graphql.Field{
			Type: graphql.NewList(graphql.String),
		},
		"citations": &graphql.Field{
			Type: graphql.NewList(CitationType),
		},
		"legalTopics": &graphql.Field{
			Type: graphql.NewList(graphql.String),
		},
		"url": &graphql.Field{
			Type: graphql.String,
		},
		"qualityScore": &graphql.Field{
			Type: graphql.Float,
		},
		"createdAt": &graphql.Field{
			Type: graphql.DateTime,
		},
		"updatedAt": &graphql.Field{
			Type: graphql.DateTime,
		},
	},
})

// CitationType represents a citation
var CitationType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Citation",
	Fields: graphql.Fields{
		"targetCaseId": &graphql.Field{
			Type: graphql.String,
		},
		"targetCaseNumber": &graphql.Field{
			Type: graphql.String,
		},
		"context": &graphql.Field{
			Type: graphql.String,
		},
	},
})

// SearchResultType represents search results with pagination
var SearchResultType = graphql.NewObject(graphql.ObjectConfig{
	Name: "SearchResult",
	Fields: graphql.Fields{
		"cases": &graphql.Field{
			Type: graphql.NewList(CaseType),
		},
		"total": &graphql.Field{
			Type: graphql.Int,
		},
		"offset": &graphql.Field{
			Type: graphql.Int,
		},
		"limit": &graphql.Field{
			Type: graphql.Int,
		},
		"hasMore": &graphql.Field{
			Type: graphql.Boolean,
		},
	},
})

// StatsType represents system statistics
var StatsType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Stats",
	Fields: graphql.Fields{
		"totalCases": &graphql.Field{
			Type: graphql.Int,
		},
		"casesByJurisdiction": &graphql.Field{
			Type: graphql.NewList(JurisdictionStatsType),
		},
		"recentCases": &graphql.Field{
			Type: graphql.Int,
		},
		"averageQualityScore": &graphql.Field{
			Type: graphql.Float,
		},
	},
})

// JurisdictionStatsType represents statistics for a jurisdiction
var JurisdictionStatsType = graphql.NewObject(graphql.ObjectConfig{
	Name: "JurisdictionStats",
	Fields: graphql.Fields{
		"jurisdiction": &graphql.Field{
			Type: graphql.String,
		},
		"count": &graphql.Field{
			Type: graphql.Int,
		},
		"percentage": &graphql.Field{
			Type: graphql.Float,
		},
	},
})

// ScrapeJobType represents a scraping job
var ScrapeJobType = graphql.NewObject(graphql.ObjectConfig{
	Name: "ScrapeJob",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type: graphql.String,
		},
		"jurisdiction": &graphql.Field{
			Type: graphql.String,
		},
		"status": &graphql.Field{
			Type: graphql.String,
		},
		"casesScraped": &graphql.Field{
			Type: graphql.Int,
		},
		"startedAt": &graphql.Field{
			Type: graphql.DateTime,
		},
		"completedAt": &graphql.Field{
			Type: graphql.DateTime,
		},
		"errorMessage": &graphql.Field{
			Type: graphql.String,
		},
	},
})

// FilterInputType for filtering cases
var FilterInputType = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "FilterInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"jurisdiction": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
		"court": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
		"startDate": &graphql.InputObjectFieldConfig{
			Type: graphql.DateTime,
		},
		"endDate": &graphql.InputObjectFieldConfig{
			Type: graphql.DateTime,
		},
		"legalTopics": &graphql.InputObjectFieldConfig{
			Type: graphql.NewList(graphql.String),
		},
		"minQualityScore": &graphql.InputObjectFieldConfig{
			Type: graphql.Float,
		},
	},
})

// CreateCaseInputType for creating cases
var CreateCaseInputType = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "CreateCaseInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"caseName": &graphql.InputObjectFieldConfig{
			Type: graphql.NewNonNull(graphql.String),
		},
		"caseNumber": &graphql.InputObjectFieldConfig{
			Type: graphql.NewNonNull(graphql.String),
		},
		"court": &graphql.InputObjectFieldConfig{
			Type: graphql.NewNonNull(graphql.String),
		},
		"jurisdiction": &graphql.InputObjectFieldConfig{
			Type: graphql.NewNonNull(graphql.String),
		},
		"decisionDate": &graphql.InputObjectFieldConfig{
			Type: graphql.DateTime,
		},
		"summary": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
		"fullText": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
		"judges": &graphql.InputObjectFieldConfig{
			Type: graphql.NewList(graphql.String),
		},
		"parties": &graphql.InputObjectFieldConfig{
			Type: graphql.NewList(graphql.String),
		},
		"legalTopics": &graphql.InputObjectFieldConfig{
			Type: graphql.NewList(graphql.String),
		},
		"url": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
	},
})

// UpdateCaseInputType for updating cases
var UpdateCaseInputType = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "UpdateCaseInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"caseName": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
		"summary": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
		"fullText": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
		"judges": &graphql.InputObjectFieldConfig{
			Type: graphql.NewList(graphql.String),
		},
		"parties": &graphql.InputObjectFieldConfig{
			Type: graphql.NewList(graphql.String),
		},
		"legalTopics": &graphql.InputObjectFieldConfig{
			Type: graphql.NewList(graphql.String),
		},
	},
})
