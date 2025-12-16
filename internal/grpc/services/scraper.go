package services

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/gongahkia/kite/internal/observability"
	"github.com/gongahkia/kite/internal/queue"
	"github.com/gongahkia/kite/internal/storage"
	pb "github.com/gongahkia/kite/api/proto"
)

// ScraperService implements the gRPC ScraperService
type ScraperService struct {
	pb.UnimplementedScraperServiceServer
	storage storage.Storage
	queue   queue.Queue
	logger  *observability.Logger
}

// NewScraperService creates a new scraper service
func NewScraperService(storage storage.Storage, queue queue.Queue, logger *observability.Logger) *ScraperService {
	return &ScraperService{
		storage: storage,
		queue:   queue,
		logger:  logger,
	}
}

// StartScrape initiates a new scraping job
func (s *ScraperService) StartScrape(ctx context.Context, req *pb.StartScrapeRequest) (*pb.ScrapeJob, error) {
	s.logger.WithFields(map[string]interface{}{
		"jurisdiction": req.Jurisdiction,
		"court":        req.Court,
	}).Info("Starting scrape job")

	// Create a new job
	job := queue.NewJob(queue.JobTypeScrape, map[string]interface{}{
		"jurisdiction": req.Jurisdiction,
		"court":        req.Court,
		"start_date":   req.StartDate,
		"end_date":     req.EndDate,
		"max_cases":    req.MaxCases,
		"options":      req.Options,
	})

	// Set priority
	switch req.Priority {
	case "high":
		job.SetPriority(queue.PriorityHigh)
	case "low":
		job.SetPriority(queue.PriorityLow)
	default:
		job.SetPriority(queue.PriorityNormal)
	}

	// Enqueue the job
	if err := s.queue.Enqueue(ctx, job); err != nil {
		s.logger.WithField("error", err).Error("Failed to enqueue scrape job")
		return nil, status.Errorf(codes.Internal, "failed to enqueue job: %v", err)
	}

	// Return job info
	return &pb.ScrapeJob{
		JobId:        job.ID,
		Jurisdiction: req.Jurisdiction,
		Court:        req.Court,
		Status:       string(job.Status),
		CreatedAt:    timestamppb.New(job.CreatedAt),
	}, nil
}

// GetScrapeStatus retrieves the status of a scraping job
func (s *ScraperService) GetScrapeStatus(ctx context.Context, req *pb.GetScrapeStatusRequest) (*pb.ScrapeJob, error) {
	// In production, you'd retrieve this from storage or queue
	// For now, return a placeholder
	return &pb.ScrapeJob{
		JobId:  req.JobId,
		Status: "running",
	}, nil
}

// ListScrapeJobs lists all scraping jobs
func (s *ScraperService) ListScrapeJobs(ctx context.Context, req *pb.ListScrapeJobsRequest) (*pb.ListScrapeJobsResponse, error) {
	// Placeholder implementation
	jobs := []*pb.ScrapeJob{}

	pagination := &pb.Pagination{
		Page:       1,
		PageSize:   int32(req.Limit),
		TotalCount: 0,
		TotalPages: 0,
	}

	return &pb.ListScrapeJobsResponse{
		Jobs:       jobs,
		Pagination: pagination,
	}, nil
}

// StreamScrapeProgress streams real-time progress updates
func (s *ScraperService) StreamScrapeProgress(req *pb.StreamScrapeProgressRequest, stream pb.ScraperService_StreamScrapeProgressServer) error {
	s.logger.WithField("job_id", req.JobId).Info("Streaming scrape progress")

	// Simulate progress updates
	for i := 0; i <= 100; i += 10 {
		update := &pb.ScrapeProgressUpdate{
			JobId:           req.JobId,
			Status:          "running",
			ScrapedCases:    int32(i),
			TotalCases:      100,
			ProgressPercent: float64(i),
			CurrentCase:     fmt.Sprintf("Case %d", i),
			Timestamp:       timestamppb.Now(),
		}

		if err := stream.Send(update); err != nil {
			return status.Errorf(codes.Internal, "failed to send update: %v", err)
		}

		// Simulate work
		time.Sleep(500 * time.Millisecond)
	}

	// Final update
	finalUpdate := &pb.ScrapeProgressUpdate{
		JobId:           req.JobId,
		Status:          "completed",
		ScrapedCases:    100,
		TotalCases:      100,
		ProgressPercent: 100,
		Timestamp:       timestamppb.Now(),
	}

	return stream.Send(finalUpdate)
}

// CancelScrape cancels a running scraping job
func (s *ScraperService) CancelScrape(ctx context.Context, req *pb.CancelScrapeRequest) (*emptypb.Empty, error) {
	s.logger.WithField("job_id", req.JobId).Info("Cancelling scrape job")

	// In production, implement job cancellation logic

	return &emptypb.Empty{}, nil
}
