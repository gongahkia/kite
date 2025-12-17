package graphql

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/graphql-go/graphql"
	"github.com/rs/zerolog/log"
)

// GraphQLRequest represents a GraphQL request
type GraphQLRequest struct {
	Query         string                 `json:"query"`
	Variables     map[string]interface{} `json:"variables"`
	OperationName string                 `json:"operationName"`
}

// GraphQLResponse represents a GraphQL response
type GraphQLResponse struct {
	Data   interface{}            `json:"data,omitempty"`
	Errors []GraphQLError         `json:"errors,omitempty"`
}

// GraphQLError represents a GraphQL error
type GraphQLError struct {
	Message string `json:"message"`
	Path    []string `json:"path,omitempty"`
}

// Handler creates a Fiber handler for GraphQL
func Handler(schema graphql.Schema) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Parse request
		var req GraphQLRequest
		if err := c.BodyParser(&req); err != nil {
			log.Error().Err(err).Msg("Failed to parse GraphQL request")
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"errors": []GraphQLError{{Message: "Invalid request body"}},
			})
		}

		// Log query
		log.Debug().
			Str("query", req.Query).
			Interface("variables", req.Variables).
			Str("operation", req.OperationName).
			Msg("GraphQL query received")

		// Execute query
		result := ExecuteQuery(schema, req.Query, req.Variables, c.Context())

		// Convert errors
		var gqlErrors []GraphQLError
		if len(result.Errors) > 0 {
			for _, err := range result.Errors {
				gqlErrors = append(gqlErrors, GraphQLError{
					Message: err.Message,
				})
			}
		}

		// Build response
		response := GraphQLResponse{
			Data:   result.Data,
			Errors: gqlErrors,
		}

		// Return response
		return c.JSON(response)
	}
}

// PlaygroundHandler serves GraphQL Playground
func PlaygroundHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		html := `
<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width,initial-scale=1">
  <title>Kite GraphQL Playground</title>
  <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/graphql-playground-react/build/static/css/index.css" />
  <link rel="shortcut icon" href="https://cdn.jsdelivr.net/npm/graphql-playground-react/build/favicon.png" />
  <script src="https://cdn.jsdelivr.net/npm/graphql-playground-react/build/static/js/middleware.js"></script>
</head>
<body>
  <div id="root"></div>
  <script>
    window.addEventListener('load', function (event) {
      GraphQLPlayground.init(document.getElementById('root'), {
        endpoint: '/graphql',
        settings: {
          'editor.theme': 'light',
          'editor.fontSize': 14,
          'editor.fontFamily': '"Source Code Pro", "Consolas", "Monaco", monospace'
        },
        tabs: [
          {
            endpoint: '/graphql',
            query: '# Kite GraphQL API\n# Documentation: https://docs.kite.example.com/graphql\n\n# Example: Search for cases\nquery SearchCases {\n  searchCases(query: "contract", limit: 5) {\n    cases {\n      id\n      caseName\n      caseNumber\n      court\n      jurisdiction\n    }\n    total\n    hasMore\n  }\n}\n\n# Example: Get a specific case\n# query GetCase {\n#   case(id: "cth/HCA/2023/15") {\n#     id\n#     caseName\n#     caseNumber\n#     court\n#     jurisdiction\n#     decisionDate\n#     summary\n#     judges\n#     parties\n#     citations {\n#       targetCaseNumber\n#       context\n#     }\n#   }\n# }\n'
          }
        ]
      })
    })
  </script>
</body>
</html>
`
		c.Set("Content-Type", "text/html")
		return c.SendString(html)
	}
}

// IntrospectionHandler handles schema introspection
func IntrospectionHandler(schema graphql.Schema) fiber.Handler {
	return func(c *fiber.Ctx) error {
		introspectionQuery := `
			query IntrospectionQuery {
				__schema {
					queryType { name }
					mutationType { name }
					subscriptionType { name }
					types {
						...FullType
					}
					directives {
						name
						description
						locations
						args {
							...InputValue
						}
					}
				}
			}

			fragment FullType on __Type {
				kind
				name
				description
				fields(includeDeprecated: true) {
					name
					description
					args {
						...InputValue
					}
					type {
						...TypeRef
					}
					isDeprecated
					deprecationReason
				}
				inputFields {
					...InputValue
				}
				interfaces {
					...TypeRef
				}
				enumValues(includeDeprecated: true) {
					name
					description
					isDeprecated
					deprecationReason
				}
				possibleTypes {
					...TypeRef
				}
			}

			fragment InputValue on __InputValue {
				name
				description
				type { ...TypeRef }
				defaultValue
			}

			fragment TypeRef on __Type {
				kind
				name
				ofType {
					kind
					name
					ofType {
						kind
						name
						ofType {
							kind
							name
							ofType {
								kind
								name
								ofType {
									kind
									name
									ofType {
										kind
										name
										ofType {
											kind
											name
										}
									}
								}
							}
						}
					}
				}
			}
		`

		result := ExecuteQuery(schema, introspectionQuery, nil, c.Context())

		// Convert to JSON
		jsonData, err := json.Marshal(result.Data)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to serialize introspection result",
			})
		}

		return c.Send(jsonData)
	}
}
