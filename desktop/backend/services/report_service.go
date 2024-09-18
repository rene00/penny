package services

import (
	"context"
	"database/sql"
	"pennyapp/backend/logwrap"
	"pennyapp/backend/types"
)

type reportService struct {
	ctx    context.Context
	db     *sql.DB
	logger *logwrap.LogWrap
}

func Report() *reportService {
	return &reportService{}
}

func (r *reportService) Start(ctx context.Context, db *sql.DB, logger *logwrap.LogWrap) {
	r.ctx = ctx
	r.db = db
	r.logger = logger
}

func (r *reportService) GetProfitLossReport() types.ReportProfitLossResponse {
	var err error
	var resp types.ReportProfitLossResponse

	if resp.Data, err = types.NewReportProfitLoss(r.ctx, r.db); err != nil {
		resp.Msg = err.Error()
		r.logger.Error(resp.Msg)
		return resp
	}

	resp.Success = true

	return resp
}
