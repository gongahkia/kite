# Kite GraphQL API Documentation

Powerful GraphQL API for querying and manipulating legal case data.

## Table of Contents

- [Quick Start](#quick-start)
- [Endpoint](#endpoint)
- [Schema](#schema)
- [Queries](#queries)
- [Mutations](#mutations)
- [Types](#types)
- [Examples](#examples)
- [Best Practices](#best-practices)

## Quick Start

### GraphQL Playground

Access the interactive GraphQL Playground at:

```
http://localhost:8080/graphql/playground
```

### Basic Query

```graphql
query {
  searchCases(query: "contract", limit: 5) {
    cases {
      id
      caseName
      court
      jurisdiction
    }
    total
  }
}
```

### With curl

```bash
curl -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  -d '{
    "query": "{ searchCases(query: \"contract\", limit: 5) { cases { id caseName } total } }"
  }'
```

## Endpoint

```
POST /graphql
```

Request body:

```json
{
  "query": "...",
  "variables": {},
  "operationName": "..."
}
```

## Schema

### Query Type

```graphql
type Query {
  # Get a single case by ID
  case(id: String!): Case

  # Search for cases
  searchCases(
    query: String
    filter: FilterInput
    limit: Int = 20
    offset: Int = 0
  ): SearchResult!

  # Get cases by jurisdiction
  casesByJurisdiction(
    jurisdiction: String!
    limit: Int = 20
    offset: Int = 0
  ): [Case!]!

  # Get system statistics
  stats: Stats!
}
```

### Mutation Type

```graphql
type Mutation {
  # Create a new case
  createCase(input: CreateCaseInput!): Case!

  # Update an existing case
  updateCase(id: String!, input: UpdateCaseInput!): Case!

  # Delete a case
  deleteCase(id: String!): Boolean!
}
```

## Queries

### Get Case by ID

Retrieve a single case by its unique identifier.

```graphql
query GetCase {
  case(id: "cth/HCA/2023/15") {
    id
    caseName
    caseNumber
    court
    jurisdiction
    decisionDate
    summary
    judges
    parties
    citations {
      targetCaseNumber
      context
    }
    legalTopics
    qualityScore
  }
}
```

### Search Cases

Search for cases with optional filters and pagination.

```graphql
query SearchCases {
  searchCases(
    query: "negligence"
    filter: {
      jurisdiction: "Australia"
      court: "High Court of Australia"
      minQualityScore: 0.8
    }
    limit: 20
    offset: 0
  ) {
    cases {
      id
      caseName
      caseNumber
      court
      jurisdiction
      decisionDate
      summary
    }
    total
    offset
    limit
    hasMore
  }
}
```

### Get Cases by Jurisdiction

Retrieve cases from a specific jurisdiction.

```graphql
query CasesByJurisdiction {
  casesByJurisdiction(jurisdiction: "Canada", limit: 10) {
    id
    caseName
    caseNumber
    court
    decisionDate
  }
}
```

### Get Statistics

Retrieve system statistics.

```graphql
query GetStats {
  stats {
    totalCases
    casesByJurisdiction {
      jurisdiction
      count
      percentage
    }
    averageQualityScore
  }
}
```

## Mutations

### Create Case

Create a new legal case.

```graphql
mutation CreateCase {
  createCase(
    input: {
      caseName: "Smith v Jones"
      caseNumber: "[2023] HCA 42"
      court: "High Court of Australia"
      jurisdiction: "Australia"
      decisionDate: "2023-12-15T00:00:00Z"
      summary: "Contract law case involving breach of agreement"
      judges: ["Chief Justice Kiefel", "Justice Gageler"]
      parties: ["John Smith", "Mary Jones"]
      legalTopics: ["Contract Law", "Breach of Agreement"]
      url: "https://example.com/cases/2023/42"
    }
  ) {
    id
    caseName
    caseNumber
    createdAt
  }
}
```

### Update Case

Update an existing case.

```graphql
mutation UpdateCase {
  updateCase(
    id: "cth/HCA/2023/42"
    input: {
      summary: "Updated summary with more details"
      legalTopics: ["Contract Law", "Breach of Agreement", "Damages"]
    }
  ) {
    id
    caseName
    summary
    legalTopics
    updatedAt
  }
}
```

### Delete Case

Delete a case by ID.

```graphql
mutation DeleteCase {
  deleteCase(id: "cth/HCA/2023/42")
}
```

## Types

### Case

```graphql
type Case {
  id: String!
  caseName: String!
  caseNumber: String!
  court: String!
  jurisdiction: String!
  decisionDate: DateTime
  summary: String
  fullText: String
  judges: [String!]!
  parties: [String!]!
  citations: [Citation!]!
  legalTopics: [String!]!
  url: String
  qualityScore: Float
  createdAt: DateTime!
  updatedAt: DateTime!
}
```

### Citation

```graphql
type Citation {
  targetCaseId: String!
  targetCaseNumber: String!
  context: String
}
```

### SearchResult

```graphql
type SearchResult {
  cases: [Case!]!
  total: Int!
  offset: Int!
  limit: Int!
  hasMore: Boolean!
}
```

### Stats

```graphql
type Stats {
  totalCases: Int!
  casesByJurisdiction: [JurisdictionStats!]!
  averageQualityScore: Float!
}
```

### JurisdictionStats

```graphql
type JurisdictionStats {
  jurisdiction: String!
  count: Int!
  percentage: Float!
}
```

### Input Types

#### FilterInput

```graphql
input FilterInput {
  jurisdiction: String
  court: String
  startDate: DateTime
  endDate: DateTime
  legalTopics: [String!]
  minQualityScore: Float
}
```

#### CreateCaseInput

```graphql
input CreateCaseInput {
  caseName: String!
  caseNumber: String!
  court: String!
  jurisdiction: String!
  decisionDate: DateTime
  summary: String
  fullText: String
  judges: [String!]
  parties: [String!]
  legalTopics: [String!]
  url: String
}
```

#### UpdateCaseInput

```graphql
input UpdateCaseInput {
  caseName: String
  summary: String
  fullText: String
  judges: [String!]
  parties: [String!]
  legalTopics: [String!]
}
```

## Examples

### Advanced Search with Filters

```graphql
query AdvancedSearch {
  searchCases(
    query: "employment discrimination"
    filter: {
      jurisdiction: "Canada"
      startDate: "2020-01-01T00:00:00Z"
      endDate: "2023-12-31T23:59:59Z"
      legalTopics: ["Employment Law", "Discrimination"]
      minQualityScore: 0.7
    }
    limit: 50
  ) {
    cases {
      id
      caseName
      caseNumber
      court
      decisionDate
      summary
      legalTopics
      qualityScore
    }
    total
    hasMore
  }
}
```

### Pagination

```graphql
query PaginatedSearch {
  searchCases(query: "contract", limit: 20, offset: 40) {
    cases {
      id
      caseName
    }
    total
    offset
    limit
    hasMore
  }
}
```

### Field Selection

Request only the fields you need:

```graphql
query MinimalFields {
  searchCases(query: "tort", limit: 10) {
    cases {
      id
      caseName
      court
    }
    total
  }
}
```

### Complex Query

```graphql
query ComplexQuery {
  case(id: "cth/HCA/2023/15") {
    id
    caseName
    caseNumber
    court
    jurisdiction
    decisionDate
    summary
    judges
    parties
    citations {
      targetCaseNumber
      context
    }
    legalTopics
  }

  stats {
    totalCases
    casesByJurisdiction {
      jurisdiction
      count
    }
  }
}
```

### Using Variables

Query:

```graphql
query SearchWithVariables($query: String!, $jurisdiction: String, $limit: Int) {
  searchCases(
    query: $query
    filter: { jurisdiction: $jurisdiction }
    limit: $limit
  ) {
    cases {
      id
      caseName
      court
    }
    total
  }
}
```

Variables:

```json
{
  "query": "negligence",
  "jurisdiction": "Australia",
  "limit": 10
}
```

### Fragments

```graphql
fragment CaseDetails on Case {
  id
  caseName
  caseNumber
  court
  jurisdiction
  decisionDate
}

query GetMultipleCases {
  case1: case(id: "cth/HCA/2023/15") {
    ...CaseDetails
  }
  case2: case(id: "cth/HCA/2023/16") {
    ...CaseDetails
  }
}
```

## Client Examples

### JavaScript (fetch)

```javascript
async function searchCases(query, limit = 20) {
  const response = await fetch('http://localhost:8080/graphql', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      query: `
        query SearchCases($query: String!, $limit: Int) {
          searchCases(query: $query, limit: $limit) {
            cases {
              id
              caseName
              court
              jurisdiction
            }
            total
          }
        }
      `,
      variables: {
        query,
        limit,
      },
    }),
  });

  const { data, errors } = await response.json();

  if (errors) {
    throw new Error(errors[0].message);
  }

  return data.searchCases;
}

// Usage
searchCases('contract', 10)
  .then(result => {
    console.log(`Found ${result.total} cases`);
    console.log(result.cases);
  });
```

### Python (requests)

```python
import requests

def search_cases(query, limit=20):
    url = "http://localhost:8080/graphql"

    query_string = """
        query SearchCases($query: String!, $limit: Int) {
            searchCases(query: $query, limit: $limit) {
                cases {
                    id
                    caseName
                    court
                    jurisdiction
                }
                total
            }
        }
    """

    response = requests.post(
        url,
        json={
            "query": query_string,
            "variables": {
                "query": query,
                "limit": limit
            }
        }
    )

    result = response.json()

    if "errors" in result:
        raise Exception(result["errors"][0]["message"])

    return result["data"]["searchCases"]

# Usage
result = search_cases("contract", 10)
print(f"Found {result['total']} cases")
print(result["cases"])
```

### Apollo Client (React)

```javascript
import { ApolloClient, InMemoryCache, gql, useQuery } from '@apollo/client';

const client = new ApolloClient({
  uri: 'http://localhost:8080/graphql',
  cache: new InMemoryCache(),
});

const SEARCH_CASES = gql`
  query SearchCases($query: String!, $limit: Int) {
    searchCases(query: $query, limit: $limit) {
      cases {
        id
        caseName
        court
        jurisdiction
      }
      total
    }
  }
`;

function SearchResults({ query }) {
  const { loading, error, data } = useQuery(SEARCH_CASES, {
    variables: { query, limit: 20 },
  });

  if (loading) return <p>Loading...</p>;
  if (error) return <p>Error: {error.message}</p>;

  return (
    <div>
      <h2>Found {data.searchCases.total} cases</h2>
      {data.searchCases.cases.map(case => (
        <div key={case.id}>
          <h3>{case.caseName}</h3>
          <p>{case.court} - {case.jurisdiction}</p>
        </div>
      ))}
    </div>
  );
}
```

## Best Practices

### 1. Request Only Needed Fields

```graphql
# Good - specific fields
query {
  searchCases(query: "contract", limit: 10) {
    cases {
      id
      caseName
      court
    }
  }
}

# Avoid - requesting fullText unnecessarily
query {
  searchCases(query: "contract", limit: 10) {
    cases {
      id
      caseName
      fullText  # Large field
    }
  }
}
```

### 2. Use Variables

```graphql
# Good - using variables
query SearchCases($query: String!) {
  searchCases(query: $query) {
    cases { id caseName }
  }
}

# Avoid - hardcoded values
query {
  searchCases(query: "contract") {
    cases { id caseName }
  }
}
```

### 3. Paginate Results

```graphql
query PaginatedResults($offset: Int!, $limit: Int!) {
  searchCases(query: "contract", limit: $limit, offset: $offset) {
    cases { id caseName }
    total
    hasMore
  }
}
```

### 4. Use Fragments for Reusability

```graphql
fragment CaseBasicInfo on Case {
  id
  caseName
  caseNumber
  court
  jurisdiction
}

query {
  case1: case(id: "abc") { ...CaseBasicInfo }
  case2: case(id: "def") { ...CaseBasicInfo }
}
```

### 5. Handle Errors

```javascript
const result = await fetch('/graphql', { ... });
const { data, errors } = await result.json();

if (errors) {
  errors.forEach(error => {
    console.error('GraphQL Error:', error.message);
  });
}
```

## Error Handling

GraphQL errors are returned in the `errors` array:

```json
{
  "data": null,
  "errors": [
    {
      "message": "Case not found: invalid-id"
    }
  ]
}
```

## Performance Tips

1. **Limit Results**: Always use `limit` to avoid large responses
2. **Pagination**: Use `offset` and `limit` for large result sets
3. **Field Selection**: Request only necessary fields
4. **Caching**: Enable HTTP caching headers
5. **Batching**: Combine multiple queries into one request

## Support

- GraphQL Playground: http://localhost:8080/graphql/playground
- Schema Introspection: http://localhost:8080/graphql/schema
- Documentation: https://docs.kite.example.com/graphql
- GitHub: https://github.com/gongahkia/kite
