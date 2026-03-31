package api

import (
	"context"

	"exchange-data-platform/internal/model/raw"
)

type FetchRequest struct {
	Dataset string
	Market  string
	Limit   int
	JobID   string
}

type Connector interface {
	Name() string
	Region() string
	Fetch(ctx context.Context, req FetchRequest) ([]raw.Envelope, error)
}
