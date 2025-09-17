package v1

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

// GetJobStatusRequest requests job status
type GetJobStatusRequest struct {
	JobID string
}

// GetJobStatusResponse contains job status
type GetJobStatusResponse struct {
	JobStatus
}

// WaitOptions configures job polling behavior
type WaitOptions struct {
	JobID        string
	InitialDelay time.Duration
	MaxDelay     time.Duration
	Jitter       time.Duration
}

// GetJobStatus checks status of async job
func (c *Client) GetJobStatus(ctx context.Context, req GetJobStatusRequest, resp *GetJobStatusResponse) error {
	path := fmt.Sprintf("job_status/%s", req.JobID)
	return c.do(ctx, "GET", path, nil, resp)
}

// WaitForJob polls job status until completion with configurable timing
func (c *Client) WaitForJob(ctx context.Context, opts WaitOptions, result *JobResult) error {
	initialDelay := opts.InitialDelay
	if initialDelay == 0 {
		initialDelay = time.Second
	}
	maxDelay := opts.MaxDelay
	if maxDelay == 0 {
		maxDelay = 30 * time.Second
	}
	jitter := opts.Jitter
	if jitter == 0 {
		jitter = 500 * time.Millisecond
	}

	delay := initialDelay
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
			var statusResp GetJobStatusResponse
			err := c.GetJobStatus(ctx, GetJobStatusRequest{JobID: opts.JobID}, &statusResp)
			if err != nil {
				return err
			}

			switch statusResp.Status {
			case "completed":
				if statusResp.Result != nil {
					*result = *statusResp.Result
				} else {
					*result = JobResult{Success: true}
				}
				return nil
			case "failed", "cancelled":
				if statusResp.Result != nil {
					*result = *statusResp.Result
				} else {
					*result = JobResult{Success: false, Error: statusResp.Error}
				}
				return fmt.Errorf("job %s: %s", statusResp.Status, statusResp.Error)
			case "pending", "working", "processing":
				if delay < maxDelay {
					delay *= 2
					if delay > maxDelay {
						delay = maxDelay
					}
				}
				r := rand.New(rand.NewSource(time.Now().UnixNano()))
				delay += time.Duration(r.Intn(int(jitter/time.Millisecond))) * time.Millisecond
			default:
				return fmt.Errorf("unknown job status: %s", statusResp.Status)
			}
		}
	}
}