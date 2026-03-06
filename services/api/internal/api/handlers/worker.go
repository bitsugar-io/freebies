package handlers

import (
	"context"
	"time"

	"github.com/retr0h/freebie/services/api/internal/api/worker/gen"
	"github.com/retr0h/freebie/services/api/internal/worker"
)

// WorkerHandler implements the generated StrictServerInterface.
type WorkerHandler struct {
	service *worker.Service
}

// NewWorkerHandler creates a new worker handler.
func NewWorkerHandler(service *worker.Service) *WorkerHandler {
	return &WorkerHandler{service: service}
}

// CheckTriggers implements gen.StrictServerInterface.
func (wh *WorkerHandler) CheckTriggers(ctx context.Context, request gen.CheckTriggersRequestObject) (gen.CheckTriggersResponseObject, error) {
	var date time.Time
	if request.Params.Date != nil {
		date = request.Params.Date.Time
	}

	result, err := wh.service.CheckTriggers(ctx, date)
	if err != nil {
		return gen.CheckTriggers500JSONResponse{Error: err.Error()}, nil
	}

	return gen.CheckTriggers200JSONResponse{
		Triggered:   result.Triggered,
		Notified:    result.Notified,
		TotalEvents: result.TotalEvents,
	}, nil
}

// SendReminders implements gen.StrictServerInterface.
func (wh *WorkerHandler) SendReminders(ctx context.Context, request gen.SendRemindersRequestObject) (gen.SendRemindersResponseObject, error) {
	result, err := wh.service.SendReminders(ctx)
	if err != nil {
		return gen.SendReminders500JSONResponse{Error: err.Error()}, nil
	}

	return gen.SendReminders200JSONResponse{
		Sent:          result.Sent,
		Failed:        result.Failed,
		ExpiringDeals: result.ExpiringDeals,
	}, nil
}
