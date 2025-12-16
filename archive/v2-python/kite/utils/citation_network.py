"""
Citation network analysis and visualization utilities.
"""

from typing import Dict, List, Set, Tuple, Optional
from collections import defaultdict, Counter
from dataclasses import dataclass
import json

from .logging_config import get_logger

logger = get_logger(__name__)


@dataclass
class NetworkMetrics:
    """Network analysis metrics for a case."""
    case_id: str
    in_degree: int  # Number of cases citing this case
    out_degree: int  # Number of cases this case cites
    authority_score: float  # How authoritative (based on citations received)
    hub_score: float  # How well-connected as a hub (based on citations made)
    
    def to_dict(self) -> Dict:
        return {
            "case_id": self.case_id,
            "in_degree": self.in_degree,
            "out_degree": self.out_degree,
            "authority_score": self.authority_score,
            "hub_score": self.hub_score,
        }


class CitationNetworkAnalyzer:
    """Analyze citation networks to discover patterns and relationships."""
    
    def __init__(self):
        self.graph: Dict[str, Set[str]] = defaultdict(set)  # citing -> cited
        self.reverse_graph: Dict[str, Set[str]] = defaultdict(set)  # cited -> citing
        self.case_metadata: Dict[str, Dict] = {}
        self.logger = logger
    
    def add_case(self, case_id: str, cited_cases: List[str], metadata: Optional[Dict] = None):
        """
        Add a case and its citations to the network.
        
        Args:
            case_id: Identifier for the citing case
            cited_cases: List of case IDs that this case cites
            metadata: Optional metadata about the case
        """
        self.graph[case_id].update(cited_cases)
        
        for cited in cited_cases:
            self.reverse_graph[cited].add(case_id)
        
        if metadata:
            self.case_metadata[case_id] = metadata
    
    def get_metrics(self, case_id: str) -> NetworkMetrics:
        """
        Calculate network metrics for a case.
        
        Args:
            case_id: Case identifier
            
        Returns:
            NetworkMetrics object
        """
        in_degree = len(self.reverse_graph.get(case_id, set()))
        out_degree = len(self.graph.get(case_id, set()))
        
        # Simple authority score based on citation count
        # More sophisticated algorithms like PageRank could be implemented
        authority_score = self._calculate_authority_score(case_id)
        
        # Hub score based on citing important cases
        hub_score = self._calculate_hub_score(case_id)
        
        return NetworkMetrics(
            case_id=case_id,
            in_degree=in_degree,
            out_degree=out_degree,
            authority_score=authority_score,
            hub_score=hub_score,
        )
    
    def _calculate_authority_score(self, case_id: str) -> float:
        """
        Calculate authority score (simplified HITS algorithm).
        
        Args:
            case_id: Case identifier
            
        Returns:
            Authority score (0-1)
        """
        # Count citations
        citation_count = len(self.reverse_graph.get(case_id, set()))
        
        if citation_count == 0:
            return 0.0
        
        # Normalize by max citations in network
        max_citations = max(
            len(citing) for citing in self.reverse_graph.values()
        ) if self.reverse_graph else 1
        
        return min(citation_count / max_citations, 1.0)
    
    def _calculate_hub_score(self, case_id: str) -> float:
        """
        Calculate hub score (simplified HITS algorithm).
        
        Args:
            case_id: Case identifier
            
        Returns:
            Hub score (0-1)
        """
        # Sum authority scores of cited cases
        cited_cases = self.graph.get(case_id, set())
        
        if not cited_cases:
            return 0.0
        
        total_authority = sum(
            len(self.reverse_graph.get(cited, set()))
            for cited in cited_cases
        )
        
        # Normalize
        max_possible = len(cited_cases) * max(
            len(citing) for citing in self.reverse_graph.values()
        ) if self.reverse_graph else 1
        
        return min(total_authority / max_possible if max_possible > 0 else 0.0, 1.0)
    
    def find_influential_cases(self, limit: int = 10) -> List[Tuple[str, NetworkMetrics]]:
        """
        Find most influential cases in the network.
        
        Args:
            limit: Number of cases to return
            
        Returns:
            List of (case_id, metrics) tuples sorted by influence
        """
        all_cases = set(self.graph.keys()) | set(self.reverse_graph.keys())
        
        case_metrics = []
        for case_id in all_cases:
            metrics = self.get_metrics(case_id)
            # Influence = combination of authority and hub scores
            influence = (metrics.authority_score + metrics.hub_score) / 2
            case_metrics.append((case_id, metrics, influence))
        
        # Sort by influence
        case_metrics.sort(key=lambda x: x[2], reverse=True)
        
        return [(case_id, metrics) for case_id, metrics, _ in case_metrics[:limit]]
    
    def find_citation_chains(self, start_case: str, end_case: str, max_depth: int = 5) -> List[List[str]]:
        """
        Find citation chains between two cases.
        
        Args:
            start_case: Starting case ID
            end_case: Target case ID
            max_depth: Maximum chain length
            
        Returns:
            List of citation chains (lists of case IDs)
        """
        chains = []
        
        def dfs(current: str, target: str, path: List[str], depth: int):
            if depth > max_depth:
                return
            
            if current == target:
                chains.append(path.copy())
                return
            
            for cited in self.graph.get(current, set()):
                if cited not in path:  # Avoid cycles
                    path.append(cited)
                    dfs(cited, target, path, depth + 1)
                    path.pop()
        
        dfs(start_case, end_case, [start_case], 1)
        return chains
    
    def get_co_citation_network(self, case_id: str) -> Dict[str, int]:
        """
        Find cases frequently co-cited with the given case.
        
        Args:
            case_id: Case identifier
            
        Returns:
            Dictionary mapping case_id to co-citation count
        """
        # Find all cases that cite the target case
        citing_cases = self.reverse_graph.get(case_id, set())
        
        # Count how often other cases are cited alongside the target
        co_citations = Counter()
        
        for citing_case in citing_cases:
            cited_by_this = self.graph.get(citing_case, set())
            for co_cited in cited_by_this:
                if co_cited != case_id:
                    co_citations[co_cited] += 1
        
        return dict(co_citations)
    
    def export_to_json(self, filename: str):
        """
        Export network to JSON format.
        
        Args:
            filename: Output filename
        """
        data = {
            "nodes": [],
            "edges": [],
        }
        
        # Collect all unique cases
        all_cases = set(self.graph.keys()) | set(self.reverse_graph.keys())
        
        # Add nodes
        for case_id in all_cases:
            metrics = self.get_metrics(case_id)
            node = {
                "id": case_id,
                "metrics": metrics.to_dict(),
            }
            
            if case_id in self.case_metadata:
                node["metadata"] = self.case_metadata[case_id]
            
            data["nodes"].append(node)
        
        # Add edges
        for citing_case, cited_cases in self.graph.items():
            for cited_case in cited_cases:
                data["edges"].append({
                    "source": citing_case,
                    "target": cited_case,
                })
        
        with open(filename, 'w') as f:
            json.dump(data, f, indent=2)
        
        self.logger.info(
            "network_exported",
            filename=filename,
            nodes=len(data["nodes"]),
            edges=len(data["edges"]),
        )
    
    def get_network_stats(self) -> Dict:
        """
        Get overall network statistics.
        
        Returns:
            Dictionary of network statistics
        """
        all_cases = set(self.graph.keys()) | set(self.reverse_graph.keys())
        
        total_citations = sum(len(cited) for cited in self.graph.values())
        
        # Calculate average citations per case
        avg_citations_made = total_citations / len(all_cases) if all_cases else 0
        
        # Find most cited case
        most_cited = max(
            ((case_id, len(citing))
             for case_id, citing in self.reverse_graph.items()),
            key=lambda x: x[1],
            default=("none", 0)
        )
        
        return {
            "total_cases": len(all_cases),
            "total_citations": total_citations,
            "avg_citations_per_case": avg_citations_made,
            "most_cited_case": most_cited[0],
            "most_cited_count": most_cited[1],
        }
