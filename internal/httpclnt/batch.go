package httpclnt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
)

const (
	// DefaultBatchSize is the default number of operations per batch request
	DefaultBatchSize = 90

	// Batch boundary prefixes (must match OData multipart/mixed format)
	batchBoundaryPrefix     = "batch_"
	changesetBoundaryPrefix = "changeset_"
)

// BatchOperation represents a single operation in a batch request
type BatchOperation struct {
	Method    string            // HTTP method (POST, PUT, DELETE, PATCH, GET)
	Path      string            // API path (e.g., "/api/v1/StringParameters")
	Body      []byte            // Request body (raw bytes - caller handles marshaling)
	ContentID string            // Content-ID for tracking this operation
	Headers   map[string]string // Additional headers (e.g., If-Match, Content-Type)
	IsQuery   bool              // True for GET operations (goes in query section, not changeset)
}

// BatchResponse represents the response from a batch request
type BatchResponse struct {
	Operations []BatchOperationResponse
}

// BatchOperationResponse represents a single operation response
type BatchOperationResponse struct {
	ContentID  string
	StatusCode int
	Headers    http.Header
	Body       []byte
	Error      error
}

// BatchRequest handles building and executing OData $batch requests
type BatchRequest struct {
	exe               *HTTPExecuter
	operations        []BatchOperation
	batchBoundary     string
	changesetBoundary string
}

// boundaryCounter is used to generate unique boundary strings
var boundaryCounter = 0

// NewBatchRequest creates a new batch request builder
func (e *HTTPExecuter) NewBatchRequest() *BatchRequest {
	return &BatchRequest{
		exe:               e,
		operations:        make([]BatchOperation, 0),
		batchBoundary:     generateBoundary(batchBoundaryPrefix),
		changesetBoundary: generateBoundary(changesetBoundaryPrefix),
	}
}

// AddOperation adds an operation to the batch
func (br *BatchRequest) AddOperation(op BatchOperation) {
	br.operations = append(br.operations, op)
}

// Execute sends the batch request and returns the responses
func (br *BatchRequest) Execute() (*BatchResponse, error) {
	if len(br.operations) == 0 {
		return &BatchResponse{Operations: []BatchOperationResponse{}}, nil
	}

	// Build multipart batch request body
	body, err := br.buildBatchBody()
	if err != nil {
		return nil, fmt.Errorf("failed to build batch body: %w", err)
	}

	// Execute the batch request
	contentType := fmt.Sprintf("multipart/mixed; boundary=%s", br.batchBoundary)
	headers := map[string]string{
		"Content-Type": contentType,
		"Accept":       "multipart/mixed",
	}

	resp, err := br.exe.ExecRequestWithCookies("POST", "/api/v1/$batch", bytes.NewReader(body), headers, nil)
	if err != nil {
		return nil, fmt.Errorf("batch request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("batch request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse the multipart response
	return br.parseBatchResponse(resp)
}

// ExecuteInBatches splits operations into batches and executes them
func (br *BatchRequest) ExecuteInBatches(batchSize int) (*BatchResponse, error) {
	if batchSize <= 0 {
		batchSize = DefaultBatchSize
	}

	allOps := br.operations
	var allResponses []BatchOperationResponse

	for i := 0; i < len(allOps); i += batchSize {
		end := i + batchSize
		if end > len(allOps) {
			end = len(allOps)
		}

		// Create a batch for this chunk
		batch := br.exe.NewBatchRequest()
		batch.operations = allOps[i:end]

		// Execute this batch
		resp, err := batch.Execute()
		if err != nil {
			return nil, fmt.Errorf("batch %d-%d failed: %w", i, end, err)
		}

		allResponses = append(allResponses, resp.Operations...)
	}

	return &BatchResponse{Operations: allResponses}, nil
}

// buildBatchBody constructs the multipart batch request body
func (br *BatchRequest) buildBatchBody() ([]byte, error) {
	var buf bytes.Buffer

	// Separate query and changeset operations
	var queryOps []BatchOperation
	var changesetOps []BatchOperation

	for _, op := range br.operations {
		if op.IsQuery {
			queryOps = append(queryOps, op)
		} else {
			changesetOps = append(changesetOps, op)
		}
	}

	// Start batch boundary
	fmt.Fprintf(&buf, "--%s\r\n", br.batchBoundary)

	// Add query operations (if any) - these go directly in batch, not in changeset
	if len(queryOps) > 0 {
		for _, op := range queryOps {
			if err := br.writeQueryOperation(&buf, op); err != nil {
				return nil, err
			}
			fmt.Fprintf(&buf, "--%s\r\n", br.batchBoundary)
		}
	}

	// Add changeset for modifying operations (POST, PUT, DELETE, PATCH)
	if len(changesetOps) > 0 {
		fmt.Fprintf(&buf, "Content-Type: multipart/mixed; boundary=%s\r\n", br.changesetBoundary)
		fmt.Fprintf(&buf, "\r\n")

		// Add each operation as a changeset part
		for _, op := range changesetOps {
			if err := br.writeChangesetOperation(&buf, op); err != nil {
				return nil, err
			}
		}

		// End changeset boundary
		fmt.Fprintf(&buf, "--%s--\r\n", br.changesetBoundary)
		fmt.Fprintf(&buf, "\r\n")
	}

	// End batch boundary
	fmt.Fprintf(&buf, "--%s--\r\n", br.batchBoundary)

	return buf.Bytes(), nil
}

// writeQueryOperation writes a query (GET) operation to the batch body
func (br *BatchRequest) writeQueryOperation(buf *bytes.Buffer, op BatchOperation) error {
	fmt.Fprintf(buf, "Content-Type: application/http\r\n")
	fmt.Fprintf(buf, "Content-Transfer-Encoding: binary\r\n")

	if op.ContentID != "" {
		fmt.Fprintf(buf, "Content-ID: %s\r\n", op.ContentID)
	}

	fmt.Fprintf(buf, "\r\n")

	// HTTP request line
	fmt.Fprintf(buf, "%s %s HTTP/1.1\r\n", op.Method, op.Path)

	// Headers
	for key, value := range op.Headers {
		fmt.Fprintf(buf, "%s: %s\r\n", key, value)
	}

	fmt.Fprintf(buf, "\r\n")

	return nil
}

// writeChangesetOperation writes a changeset operation to the batch body
func (br *BatchRequest) writeChangesetOperation(buf *bytes.Buffer, op BatchOperation) error {
	// Changeset part boundary
	fmt.Fprintf(buf, "--%s\r\n", br.changesetBoundary)
	fmt.Fprintf(buf, "Content-Type: application/http\r\n")
	fmt.Fprintf(buf, "Content-Transfer-Encoding: binary\r\n")

	if op.ContentID != "" {
		fmt.Fprintf(buf, "Content-ID: %s\r\n", op.ContentID)
	}

	fmt.Fprintf(buf, "\r\n")

	// HTTP request line
	fmt.Fprintf(buf, "%s %s HTTP/1.1\r\n", op.Method, op.Path)

	// Headers
	for key, value := range op.Headers {
		fmt.Fprintf(buf, "%s: %s\r\n", key, value)
	}

	// Body
	if len(op.Body) > 0 {
		fmt.Fprintf(buf, "Content-Length: %d\r\n", len(op.Body))
		fmt.Fprintf(buf, "\r\n")
		buf.Write(op.Body)
	} else {
		fmt.Fprintf(buf, "\r\n")
	}

	fmt.Fprintf(buf, "\r\n")

	return nil
}

// parseBatchResponse parses the multipart batch response
func (br *BatchRequest) parseBatchResponse(resp *http.Response) (*BatchResponse, error) {
	mediaType, params, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse response content-type: %w", err)
	}

	if !strings.HasPrefix(mediaType, "multipart/") {
		return nil, fmt.Errorf("expected multipart response, got %s", mediaType)
	}

	boundary := params["boundary"]
	if boundary == "" {
		return nil, fmt.Errorf("no boundary in multipart response")
	}

	mr := multipart.NewReader(resp.Body, boundary)

	var operations []BatchOperationResponse

	// Read batch parts
	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read batch part: %w", err)
		}

		// Check if this is a changeset
		partContentType := part.Header.Get("Content-Type")
		if strings.HasPrefix(partContentType, "multipart/mixed") {
			// Parse the changeset
			changesetOps, err := br.parseChangeset(part)
			if err != nil {
				return nil, fmt.Errorf("failed to parse changeset: %w", err)
			}
			operations = append(operations, changesetOps...)
		} else if strings.HasPrefix(partContentType, "application/http") {
			// Single operation response (query result)
			op, err := br.parseOperationResponseFromPart(part)
			if err != nil {
				op = BatchOperationResponse{Error: err}
			}
			operations = append(operations, op)
		}
	}

	return &BatchResponse{Operations: operations}, nil
}

// parseChangeset parses a changeset multipart section
func (br *BatchRequest) parseChangeset(changesetReader io.Reader) ([]BatchOperationResponse, error) {
	// Read the changeset to get its boundary
	changesetBytes, err := io.ReadAll(changesetReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read changeset: %w", err)
	}

	// Extract boundary from the first line
	lines := strings.Split(string(changesetBytes), "\r\n")
	if len(lines) == 0 {
		return nil, fmt.Errorf("empty changeset")
	}

	// Find the boundary (first line starting with --)
	var changesetBoundary string
	for _, line := range lines {
		if strings.HasPrefix(line, "--") {
			changesetBoundary = strings.TrimPrefix(line, "--")
			break
		}
	}

	if changesetBoundary == "" {
		return nil, fmt.Errorf("no changeset boundary found")
	}

	mr := multipart.NewReader(bytes.NewReader(changesetBytes), changesetBoundary)

	var operations []BatchOperationResponse

	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read changeset part: %w", err)
		}

		op, err := br.parseOperationResponseFromPart(part)
		if err != nil {
			// Log error but continue with other operations
			log.Warn().Msgf("Failed to parse changeset part: %v", err)
			op = BatchOperationResponse{Error: err}
		}

		operations = append(operations, op)
	}

	return operations, nil
}

// parseOperationResponseFromPart parses a single operation response from a multipart part
func (br *BatchRequest) parseOperationResponseFromPart(part *multipart.Part) (BatchOperationResponse, error) {
	contentID := part.Header.Get("Content-Id")
	if contentID == "" {
		contentID = part.Header.Get("Content-ID")
	}

	// Read the HTTP response
	bodyBytes, err := io.ReadAll(part)
	if err != nil {
		return BatchOperationResponse{}, fmt.Errorf("failed to read operation response: %w", err)
	}

	// Parse HTTP response
	lines := strings.Split(string(bodyBytes), "\r\n")
	if len(lines) < 1 {
		return BatchOperationResponse{}, fmt.Errorf("invalid HTTP response")
	}

	// Parse status line (e.g., "HTTP/1.1 201 Created")
	statusLine := lines[0]
	parts := strings.SplitN(statusLine, " ", 3)
	if len(parts) < 2 {
		return BatchOperationResponse{}, fmt.Errorf("invalid status line: %s", statusLine)
	}

	statusCode, err := strconv.Atoi(parts[1])
	if err != nil {
		return BatchOperationResponse{}, fmt.Errorf("invalid status code: %s", parts[1])
	}

	// Parse headers
	headers := make(http.Header)
	i := 1
	for ; i < len(lines); i++ {
		line := lines[i]
		if line == "" {
			i++
			break
		}

		// Parse header
		colonIdx := strings.Index(line, ":")
		if colonIdx > 0 {
			key := strings.TrimSpace(line[:colonIdx])
			value := strings.TrimSpace(line[colonIdx+1:])
			headers.Add(key, value)
		}
	}

	// Remaining lines are the body
	var body []byte
	if i < len(lines) {
		bodyStr := strings.Join(lines[i:], "\r\n")
		body = []byte(strings.TrimSpace(bodyStr))
	}

	return BatchOperationResponse{
		ContentID:  contentID,
		StatusCode: statusCode,
		Headers:    headers,
		Body:       body,
	}, nil
}

// generateBoundary generates a unique boundary string
func generateBoundary(prefix string) string {
	boundaryCounter++
	return fmt.Sprintf("%s%d", prefix, boundaryCounter)
}

// Helper functions for building batch operations

// AddCreateStringParameterOp adds a CREATE operation for a string parameter to the batch
func AddCreateStringParameterOp(batch *BatchRequest, pid, id, value, contentID string) {
	body := map[string]string{
		"Pid":   pid,
		"Id":    id,
		"Value": value,
	}
	bodyJSON, _ := json.Marshal(body)

	batch.AddOperation(BatchOperation{
		Method:    "POST",
		Path:      "/api/v1/StringParameters",
		Body:      bodyJSON,
		ContentID: contentID,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	})
}

// AddUpdateStringParameterOp adds an UPDATE operation for a string parameter to the batch
func AddUpdateStringParameterOp(batch *BatchRequest, pid, id, value, contentID string) {
	body := map[string]string{
		"Value": value,
	}
	bodyJSON, _ := json.Marshal(body)

	path := fmt.Sprintf("/api/v1/StringParameters(Pid='%s',Id='%s')", pid, id)

	batch.AddOperation(BatchOperation{
		Method:    "PUT",
		Path:      path,
		Body:      bodyJSON,
		ContentID: contentID,
		Headers: map[string]string{
			"Content-Type": "application/json",
			"If-Match":     "*",
		},
	})
}

// AddDeleteStringParameterOp adds a DELETE operation for a string parameter to the batch
func AddDeleteStringParameterOp(batch *BatchRequest, pid, id, contentID string) {
	path := fmt.Sprintf("/api/v1/StringParameters(Pid='%s',Id='%s')", pid, id)

	batch.AddOperation(BatchOperation{
		Method:    "DELETE",
		Path:      path,
		ContentID: contentID,
		Headers: map[string]string{
			"If-Match": "*",
		},
	})
}

// AddCreateBinaryParameterOp adds a CREATE operation for a binary parameter to the batch
func AddCreateBinaryParameterOp(batch *BatchRequest, pid, id, value, contentType, contentID string) {
	body := map[string]string{
		"Pid":         pid,
		"Id":          id,
		"Value":       value,
		"ContentType": contentType,
	}
	bodyJSON, _ := json.Marshal(body)

	batch.AddOperation(BatchOperation{
		Method:    "POST",
		Path:      "/api/v1/BinaryParameters",
		Body:      bodyJSON,
		ContentID: contentID,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	})
}

// AddUpdateBinaryParameterOp adds an UPDATE operation for a binary parameter to the batch
func AddUpdateBinaryParameterOp(batch *BatchRequest, pid, id, value, contentType, contentID string) {
	body := map[string]string{
		"Value":       value,
		"ContentType": contentType,
	}
	bodyJSON, _ := json.Marshal(body)

	path := fmt.Sprintf("/api/v1/BinaryParameters(Pid='%s',Id='%s')", pid, id)

	batch.AddOperation(BatchOperation{
		Method:    "PUT",
		Path:      path,
		Body:      bodyJSON,
		ContentID: contentID,
		Headers: map[string]string{
			"Content-Type": "application/json",
			"If-Match":     "*",
		},
	})
}

// AddDeleteBinaryParameterOp adds a DELETE operation for a binary parameter to the batch
func AddDeleteBinaryParameterOp(batch *BatchRequest, pid, id, contentID string) {
	path := fmt.Sprintf("/api/v1/BinaryParameters(Pid='%s',Id='%s')", pid, id)

	batch.AddOperation(BatchOperation{
		Method:    "DELETE",
		Path:      path,
		ContentID: contentID,
		Headers: map[string]string{
			"If-Match": "*",
		},
	})
}
