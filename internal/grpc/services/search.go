package services

import (
	"context"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/gongahkia/kite/internal/observability"
	"github.com/gongahkia/kite/internal/storage"
	"github.com/gongahkia/kite/pkg/models"
	pb "github.com/gongahkia/kite/api/proto"
)

// SearchService implements the gRPC SearchService
type SearchService struct {
	pb.UnimplementedSearchServiceServer
	storage storage.Storage
	logger  *observability.Logger
}

// NewSearchService creates a new search service
func NewSearchService(storage storage.Storage, logger *observability.Logger) *SearchService {
	return &SearchService{
		storage: storage,
		logger:  logger,
	}
}

// SearchCases performs full-text search on cases
func (s *SearchService) SearchCases(ctx context.Context, req *pb.SearchCasesRequest) (*pb.SearchCasesResponse, error) {
	start := time.Now()

	s.logger.WithField("query", req.Query).Info("Searching cases")

	// Build search query
	searchQuery := storage.SearchQuery{
		Query:  req.Query,
		Fields: req.Fields,
		Fuzzy:  req.Fuzzy,
		Limit:  int(req.Limit),
		Offset: int(req.Offset),
	}

	// Apply filters if provided
	if req.Filters != nil {
		searchQuery.Filters = protoFilterToStorageFilter(req.Filters)
	}

	// Perform search
	cases, err := s.storage.SearchCases(ctx, searchQuery)
	if err != nil {
		s.logger.WithField("error", err).Error("Failed to search cases")
		return nil, status.Errorf(codes.Internal, "search failed: %v", err)
	}

	// Convert to proto results
	results := make([]*pb.CaseSearchResult, len(cases))
	for i, c := range cases {
		results[i] = &pb.CaseSearchResult{
			Case:  modelCaseToProto(c),
			Score: 1.0, // Placeholder score
		}
	}

	searchTime := time.Since(start)

	return &pb.SearchCasesResponse{
		Results:      results,
		TotalHits:    int32(len(results)),
		SearchTimeMs: float64(searchTime.Milliseconds()),
		Pagination: &pb.Pagination{
			Page:       int32(req.Offset/req.Limit) + 1,
			PageSize:   req.Limit,
			TotalCount: int32(len(results)),
		},
	}, nil
}

// GetCase retrieves a single case by ID
func (s *SearchService) GetCase(ctx context.Context, req *pb.GetCaseRequest) (*pb.Case, error) {
	s.logger.WithField("id", req.Id).Info("Getting case")

	c, err := s.storage.GetCase(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "case not found: %v", err)
	}

	return modelCaseToProto(c), nil
}

// ListCases lists cases with filtering
func (s *SearchService) ListCases(ctx context.Context, req *pb.ListCasesRequest) (*pb.ListCasesResponse, error) {
	filter := protoFilterToStorageFilter(req.Filter)

	cases, err := s.storage.ListCases(ctx, filter)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list cases: %v", err)
	}

	protoCases := make([]*pb.Case, len(cases))
	for i, c := range cases {
		protoCases[i] = modelCaseToProto(c)
	}

	count, _ := s.storage.CountCases(ctx, filter)

	return &pb.ListCasesResponse{
		Cases: protoCases,
		Pagination: &pb.Pagination{
			TotalCount: int32(count),
			PageSize:   int32(filter.Limit),
		},
	}, nil
}

// StreamCases streams cases matching criteria
func (s *SearchService) StreamCases(req *pb.StreamCasesRequest, stream pb.SearchService_StreamCasesServer) error {
	filter := protoFilterToStorageFilter(req.Filter)

	cases, err := s.storage.ListCases(stream.Context(), filter)
	if err != nil {
		return status.Errorf(codes.Internal, "failed to list cases: %v", err)
	}

	// Stream cases in batches
	batchSize := int(req.BatchSize)
	if batchSize == 0 {
		batchSize = 10
	}

	for i := 0; i < len(cases); i += batchSize {
		end := i + batchSize
		if end > len(cases) {
			end = len(cases)
		}

		for _, c := range cases[i:end] {
			if err := stream.Send(modelCaseToProto(c)); err != nil {
				return status.Errorf(codes.Internal, "failed to stream case: %v", err)
			}
		}

		// Small delay between batches
		time.Sleep(100 * time.Millisecond)
	}

	return nil
}

// GetCitationNetwork retrieves citation network for a case
func (s *SearchService) GetCitationNetwork(ctx context.Context, req *pb.GetCitationNetworkRequest) (*pb.CitationNetworkResponse, error) {
	// Placeholder implementation
	return &pb.CitationNetworkResponse{
		Nodes: []*pb.CitationNode{},
		Edges: []*pb.CitationEdge{},
		Stats: &pb.NetworkStats{
			TotalNodes: 0,
			TotalEdges: 0,
		},
	}, nil
}

// Helper functions

func protoFilterToStorageFilter(pf *pb.CaseFilter) storage.CaseFilter {
	filter := storage.CaseFilter{}

	if pf == nil {
		return filter
	}

	filter.IDs = pf.Ids
	filter.Jurisdiction = pf.Jurisdiction
	filter.Court = pf.Court
	filter.Status = models.CaseStatus(pf.Status)
	filter.Judges = pf.Judges
	filter.Concepts = pf.Concepts
	filter.MinQuality = pf.MinQuality
	filter.Limit = int(pf.Limit)
	filter.Offset = int(pf.Offset)
	filter.OrderBy = pf.OrderBy
	filter.OrderDesc = pf.OrderDesc

	if pf.CourtLevel != 0 {
		cl := models.CourtLevel(pf.CourtLevel)
		filter.CourtLevel = &cl
	}

	if pf.StartDate != nil {
		t := pf.StartDate.AsTime()
		filter.StartDate = &t
	}

	if pf.EndDate != nil {
		t := pf.EndDate.AsTime()
		filter.EndDate = &t
	}

	return filter
}

func modelCaseToProto(c *models.Case) *pb.Case {
	pc := &pb.Case{
		Id:                 c.ID,
		CaseNumber:         c.CaseNumber,
		CaseName:           c.CaseName,
		Court:              c.Court,
		CourtLevel:         int32(c.CourtLevel),
		CourtType:          c.CourtType,
		Jurisdiction:       c.Jurisdiction,
		Docket:             c.Docket,
		Summary:            c.Summary,
		FullText:           c.FullText,
		Outcome:            c.Outcome,
		ProceduralHistory:  c.ProceduralHistory,
		Url:                c.URL,
		PdfUrl:             c.PDFURL,
		SourceDatabase:     c.SourceDatabase,
		Language:           c.Language,
		Status:             string(c.Status),
	}

	if c.DecisionDate != nil {
		pc.DecisionDate = timestamppb.New(*c.DecisionDate)
	}

	if c.ScrapedAt != nil {
		pc.ScrapedAt = timestamppb.New(*c.ScrapedAt)
	}

	if c.LastUpdated != nil {
		pc.LastUpdated = timestamppb.New(*c.LastUpdated)
	}

	// Convert slices
	if c.Parties != nil {
		pc.Parties = c.Parties
	}
	if c.Judges != nil {
		pc.Judges = c.Judges
	}
	if c.KeyIssues != nil {
		pc.KeyIssues = c.KeyIssues
	}
	if c.LegalConcepts != nil {
		pc.LegalConcepts = c.LegalConcepts
	}
	if c.Citations != nil {
		pc.Citations = c.Citations
	}

	return pc
}
