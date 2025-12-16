package citation

import (
	"github.com/gongahkia/kite/pkg/models"
)

// NetworkAnalyzer analyzes citation networks and builds citation graphs
type NetworkAnalyzer struct {
	graph *models.CitationNetwork
}

// NewNetworkAnalyzer creates a new citation network analyzer
func NewNetworkAnalyzer() *NetworkAnalyzer {
	return &NetworkAnalyzer{
		graph: &models.CitationNetwork{
			Nodes: make(map[string]*models.CitationNode),
			Edges: make([]*models.CitationEdge, 0),
		},
	}
}

// BuildNetwork builds a citation network from a collection of cases
func (na *NetworkAnalyzer) BuildNetwork(cases []*models.Case, citations []*models.Citation) *models.CitationNetwork {
	// Reset the graph
	na.graph = &models.CitationNetwork{
		Nodes: make(map[string]*models.CitationNode),
		Edges: make([]*models.CitationEdge, 0),
	}

	// Add nodes for each case
	for _, c := range cases {
		na.addNode(c)
	}

	// Add edges for each citation
	for _, citation := range citations {
		na.addEdge(citation)
	}

	// Calculate metrics
	na.calculateMetrics()

	return na.graph
}

// AddCase adds a case to the network
func (na *NetworkAnalyzer) AddCase(c *models.Case) {
	na.addNode(c)
	na.calculateMetrics()
}

// AddCitation adds a citation to the network
func (na *NetworkAnalyzer) AddCitation(citation *models.Citation) {
	na.addEdge(citation)
	na.calculateMetrics()
}

// GetMostCitedCases returns the most cited cases in the network
func (na *NetworkAnalyzer) GetMostCitedCases(limit int) []*models.CitationNode {
	nodes := make([]*models.CitationNode, 0, len(na.graph.Nodes))
	for _, node := range na.graph.Nodes {
		nodes = append(nodes, node)
	}

	// Sort by inbound citations (descending)
	for i := 0; i < len(nodes); i++ {
		for j := i + 1; j < len(nodes); j++ {
			if nodes[j].InboundCitations > nodes[i].InboundCitations {
				nodes[i], nodes[j] = nodes[j], nodes[i]
			}
		}
	}

	// Return top N
	if limit > len(nodes) {
		limit = len(nodes)
	}

	return nodes[:limit]
}

// GetCitationChain finds the citation chain between two cases
func (na *NetworkAnalyzer) GetCitationChain(fromCaseID, toCaseID string) []string {
	// BFS to find shortest path
	if _, exists := na.graph.Nodes[fromCaseID]; !exists {
		return nil
	}
	if _, exists := na.graph.Nodes[toCaseID]; !exists {
		return nil
	}

	visited := make(map[string]bool)
	parent := make(map[string]string)
	queue := []string{fromCaseID}
	visited[fromCaseID] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if current == toCaseID {
			// Reconstruct path
			path := []string{toCaseID}
			for path[0] != fromCaseID {
				path = append([]string{parent[path[0]]}, path...)
			}
			return path
		}

		// Explore neighbors
		for _, edge := range na.graph.Edges {
			if edge.FromCaseID == current {
				if !visited[edge.ToCaseID] {
					visited[edge.ToCaseID] = true
					parent[edge.ToCaseID] = current
					queue = append(queue, edge.ToCaseID)
				}
			}
		}
	}

	return nil // No path found
}

// GetCitationDepth calculates the maximum citation depth for a case
func (na *NetworkAnalyzer) GetCitationDepth(caseID string) int {
	if _, exists := na.graph.Nodes[caseID]; !exists {
		return 0
	}

	visited := make(map[string]bool)
	return na.dfs(caseID, visited)
}

// dfs performs depth-first search to calculate citation depth
func (na *NetworkAnalyzer) dfs(caseID string, visited map[string]bool) int {
	if visited[caseID] {
		return 0 // Cycle detected
	}

	visited[caseID] = true
	maxDepth := 0

	for _, edge := range na.graph.Edges {
		if edge.FromCaseID == caseID {
			depth := na.dfs(edge.ToCaseID, visited)
			if depth > maxDepth {
				maxDepth = depth
			}
		}
	}

	delete(visited, caseID) // Backtrack
	return maxDepth + 1
}

// addNode adds a case as a node to the network
func (na *NetworkAnalyzer) addNode(c *models.Case) {
	if _, exists := na.graph.Nodes[c.ID]; exists {
		return // Already exists
	}

	node := &models.CitationNode{
		CaseID:            c.ID,
		CaseName:          c.CaseName,
		Year:              c.DecisionDate,
		Court:             c.Court,
		InboundCitations:  0,
		OutboundCitations: 0,
	}

	na.graph.Nodes[c.ID] = node
}

// addEdge adds a citation as an edge to the network
func (na *NetworkAnalyzer) addEdge(citation *models.Citation) {
	if citation.CitingCaseID == "" || citation.CitedCaseID == "" {
		return // Invalid edge
	}

	edge := &models.CitationEdge{
		FromCaseID: citation.CitingCaseID,
		ToCaseID:   citation.CitedCaseID,
		Citation:   citation.NormalizedCitation,
		Weight:     1,
	}

	na.graph.Edges = append(na.graph.Edges, edge)

	// Update node citation counts
	if fromNode, exists := na.graph.Nodes[citation.CitingCaseID]; exists {
		fromNode.OutboundCitations++
	}
	if toNode, exists := na.graph.Nodes[citation.CitedCaseID]; exists {
		toNode.InboundCitations++
	}
}

// calculateMetrics calculates various network metrics
func (na *NetworkAnalyzer) calculateMetrics() {
	// Update PageRank-style influence scores
	na.calculateInfluenceScores()
}

// calculateInfluenceScores calculates influence scores for each node
// This is a simplified PageRank-style algorithm
func (na *NetworkAnalyzer) calculateInfluenceScores() {
	dampingFactor := 0.85
	iterations := 10

	// Initialize scores
	scores := make(map[string]float64)
	for caseID := range na.graph.Nodes {
		scores[caseID] = 1.0
	}

	// Iterate to converge
	for i := 0; i < iterations; i++ {
		newScores := make(map[string]float64)

		for caseID := range na.graph.Nodes {
			// Base score
			newScores[caseID] = (1.0 - dampingFactor)

			// Add contributions from citing cases
			for _, edge := range na.graph.Edges {
				if edge.ToCaseID == caseID {
					fromNode := na.graph.Nodes[edge.FromCaseID]
					if fromNode.OutboundCitations > 0 {
						contribution := scores[edge.FromCaseID] / float64(fromNode.OutboundCitations)
						newScores[caseID] += dampingFactor * contribution
					}
				}
			}
		}

		scores = newScores
	}

	// Update nodes with influence scores
	for caseID, score := range scores {
		if node, exists := na.graph.Nodes[caseID]; exists {
			node.InfluenceScore = score
		}
	}
}

// GetNetworkStatistics returns statistics about the citation network
func (na *NetworkAnalyzer) GetNetworkStatistics() map[string]interface{} {
	stats := make(map[string]interface{})

	stats["total_nodes"] = len(na.graph.Nodes)
	stats["total_edges"] = len(na.graph.Edges)

	// Calculate average citations
	totalInbound := 0
	totalOutbound := 0
	for _, node := range na.graph.Nodes {
		totalInbound += node.InboundCitations
		totalOutbound += node.OutboundCitations
	}

	if len(na.graph.Nodes) > 0 {
		stats["avg_inbound_citations"] = float64(totalInbound) / float64(len(na.graph.Nodes))
		stats["avg_outbound_citations"] = float64(totalOutbound) / float64(len(na.graph.Nodes))
	}

	// Find most influential case
	maxInfluence := 0.0
	var mostInfluentialCase string
	for caseID, node := range na.graph.Nodes {
		if node.InfluenceScore > maxInfluence {
			maxInfluence = node.InfluenceScore
			mostInfluentialCase = caseID
		}
	}
	stats["most_influential_case"] = mostInfluentialCase
	stats["max_influence_score"] = maxInfluence

	return stats
}
