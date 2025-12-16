package export

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/gongahkia/kite/pkg/models"
)

// StreamExporter handles streaming export of large datasets
type StreamExporter struct {
	format     ExportFormat
	writer     io.Writer
	compress   bool
	gzipWriter *gzip.Writer
	options    *ExportOptions
}

// NewStreamExporter creates a new streaming exporter
func NewStreamExporter(format ExportFormat, writer io.Writer, options *ExportOptions) *StreamExporter {
	if options == nil {
		options = DefaultExportOptions()
	}

	se := &StreamExporter{
		format:  format,
		writer:  writer,
		compress: options.Compress,
		options: options,
	}

	// Setup gzip compression if requested
	if options.Compress {
		se.gzipWriter = gzip.NewWriter(writer)
		se.writer = se.gzipWriter
	}

	return se
}

// StreamCases streams cases one by one to the output
func (se *StreamExporter) StreamCases(ctx context.Context, cases <-chan *models.Case) error {
	defer se.Close()

	switch se.format {
	case FormatJSON:
		return se.streamJSON(ctx, cases)
	case FormatJSONLines:
		return se.streamJSONLines(ctx, cases)
	case FormatCSV:
		return se.streamCSV(ctx, cases)
	default:
		return fmt.Errorf("streaming not supported for format: %s", se.format)
	}
}

// streamJSON streams cases as a JSON array
func (se *StreamExporter) streamJSON(ctx context.Context, cases <-chan *models.Case) error {
	// Write array opening bracket
	if _, err := se.writer.Write([]byte("[\n")); err != nil {
		return err
	}

	encoder := json.NewEncoder(se.writer)
	if se.options.Pretty {
		encoder.SetIndent("  ", "  ")
	}

	first := true
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case c, ok := <-cases:
			if !ok {
				// Channel closed, write closing bracket
				if _, err := se.writer.Write([]byte("\n]")); err != nil {
					return err
				}
				return nil
			}

			// Write comma separator (except for first item)
			if !first {
				if _, err := se.writer.Write([]byte(",\n")); err != nil {
					return err
				}
			}
			first = false

			// Encode case
			if err := encoder.Encode(c); err != nil {
				return err
			}
		}
	}
}

// streamJSONLines streams cases as newline-delimited JSON
func (se *StreamExporter) streamJSONLines(ctx context.Context, cases <-chan *models.Case) error {
	encoder := json.NewEncoder(se.writer)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case c, ok := <-cases:
			if !ok {
				return nil // Channel closed, done
			}

			if err := encoder.Encode(c); err != nil {
				return err
			}
		}
	}
}

// streamCSV streams cases as CSV
func (se *StreamExporter) streamCSV(ctx context.Context, cases <-chan *models.Case) error {
	// Create a temporary exporter for CSV writing
	exporter := NewExporter(FormatCSV, se.writer)

	// Collect cases into a slice (CSV requires header first)
	// For true streaming CSV, we'd need to write header first, then stream rows
	var caseSlice []*models.Case

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case c, ok := <-cases:
			if !ok {
				// Export all collected cases
				return exporter.Export(caseSlice)
			}
			caseSlice = append(caseSlice, c)
		}
	}
}

// Close closes the stream exporter and flushes any buffered data
func (se *StreamExporter) Close() error {
	if se.gzipWriter != nil {
		return se.gzipWriter.Close()
	}
	return nil
}

// StreamFromStorage streams cases from storage adapter
func StreamFromStorage(ctx context.Context, storage interface{}, query interface{}, format ExportFormat, writer io.Writer, options *ExportOptions) error {
	// Create case channel
	cases := make(chan *models.Case, 100)

	// Create stream exporter
	exporter := NewStreamExporter(format, writer, options)

	// Start streaming in background
	errCh := make(chan error, 1)
	go func() {
		errCh <- exporter.StreamCases(ctx, cases)
	}()

	// Fetch cases and send to channel
	// This is a placeholder - actual implementation would depend on storage adapter
	// For now, we'll just close the channel to signal completion
	close(cases)

	// Wait for streaming to complete
	return <-errCh
}

// ChunkedExporter exports cases in chunks for better memory efficiency
type ChunkedExporter struct {
	format    ExportFormat
	writer    io.Writer
	chunkSize int
	options   *ExportOptions
}

// NewChunkedExporter creates a new chunked exporter
func NewChunkedExporter(format ExportFormat, writer io.Writer, chunkSize int, options *ExportOptions) *ChunkedExporter {
	if options == nil {
		options = DefaultExportOptions()
	}

	if chunkSize <= 0 {
		chunkSize = 1000 // Default chunk size
	}

	return &ChunkedExporter{
		format:    format,
		writer:    writer,
		chunkSize: chunkSize,
		options:   options,
	}
}

// ExportInChunks exports cases in chunks
func (ce *ChunkedExporter) ExportInChunks(ctx context.Context, cases []*models.Case) error {
	exporter := NewExporter(ce.format, ce.writer)

	// Split into chunks
	for i := 0; i < len(cases); i += ce.chunkSize {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		end := i + ce.chunkSize
		if end > len(cases) {
			end = len(cases)
		}

		chunk := cases[i:end]
		if err := exporter.Export(chunk); err != nil {
			return err
		}
	}

	return nil
}

// ExportProgress tracks export progress
type ExportProgress struct {
	Total     int   `json:"total"`
	Exported  int   `json:"exported"`
	Failed    int   `json:"failed"`
	StartTime int64 `json:"start_time"`
	Completed bool  `json:"completed"`
}

// ProgressCallback is called periodically during export
type ProgressCallback func(progress *ExportProgress)

// ExportWithProgress exports cases with progress tracking
func ExportWithProgress(ctx context.Context, cases []*models.Case, format ExportFormat, writer io.Writer, callback ProgressCallback) error {
	progress := &ExportProgress{
		Total:     len(cases),
		Exported:  0,
		Failed:    0,
		StartTime: context.Background().Value("start_time").(int64),
		Completed: false,
	}

	exporter := NewExporter(format, writer)

	// Export each case individually to track progress
	for i, c := range cases {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := exporter.Export([]*models.Case{c}); err != nil {
			progress.Failed++
		} else {
			progress.Exported++
		}

		// Call progress callback every 100 cases
		if callback != nil && i%100 == 0 {
			callback(progress)
		}
	}

	progress.Completed = true
	if callback != nil {
		callback(progress)
	}

	return nil
}
