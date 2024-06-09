package services

import (
	"context"
	"database/sql"
	"pennyapp/backend/model"
	"pennyapp/backend/types"

	"github.com/volatiletech/sqlboiler/v4/boil"
)

type entityService struct {
	ctx context.Context
	db  *sql.DB
}

var entity *entityService

func Entity() *entityService {
	entity = &entityService{}
	return entity
}

func (e *entityService) Start(ctx context.Context, db *sql.DB) {
	e.ctx = ctx
	e.db = db
}

func (e *entityService) ListEntities() (resp types.JSResp) {
	j := types.JSResp{
		Success: true,
		Msg:     "list entities",
		Data:    []string{"foo", "bar"},
	}
	resp.Success = true
	resp.Data = j
	return
}

func (e *entityService) SaveEntity(name string) (resp types.JSResp) {
	var err error
	entity := model.Entity{Name: name}
	if err = entity.Insert(e.ctx, e.db, boil.Infer()); err != nil {
		resp.Msg = err.Error()
		return
	}
	resp.Success = true
	return
}
