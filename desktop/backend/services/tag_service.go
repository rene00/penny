package services

import (
	"context"
	"database/sql"
	"fmt"
	"pennyapp/backend/internal/dberror"
	"pennyapp/backend/logwrap"
	"pennyapp/backend/model"
	"pennyapp/backend/types"
	"regexp"

	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type tagService struct {
	ctx         context.Context
	db          *sql.DB
	logger      *logwrap.LogWrap
	TagRequestC chan TagRequest
}

type TagRequest struct{}

func Tag() *tagService {
	return &tagService{}
}

func (t *tagService) Start(ctx context.Context, db *sql.DB, logger *logwrap.LogWrap) {
	t.ctx = ctx
	t.db = db
	t.logger = logger
	t.TagRequestC = make(chan TagRequest)

	go func() { t.matcher(t.ctx, t.db, t.logger) }()
}

func (t *tagService) matcher(ctx context.Context, db *sql.DB, logger *logwrap.LogWrap) {
	var q []qm.QueryMod

	for range t.TagRequestC {
		tags, err := model.Tags().All(ctx, db)
		if err != nil {
			logger.Error("Failed to find Tags")
		}

		q = []qm.QueryMod{
			qm.Load("Splits"),
			qm.Load("Splits.Account"),
		}
		transactions, err := model.Transactions(q...).All(ctx, db)
		if err != nil {
			logger.Error(err.Error())
			continue
		}

		for _, tag := range tags {
			tagDeleted, err := model.TagAttributes(qm.Where("tag_id=? AND name=? AND value=?", tag.ID, "deleted", "true")).Exists(ctx, db)
			if err != nil {
				logger.Error(err.Error())
				continue
			}
			if tagDeleted {
				continue
			}

			tagRegexes, err := model.TagRegexes(qm.Where("tag_id=?", tag.ID)).All(ctx, db)
			if err != nil {
				logger.Error(err.Error())
				continue
			}

			for _, transaction := range transactions {
				transactionDeleted, err := model.TransactionsAttributes(qm.Where("transactions_id=? AND name=? AND value=?", transaction.ID, "deleted", "true")).Exists(ctx, db)
				if err != nil {
					t.logger.Error(err.Error())
					continue
				}
				if transactionDeleted {
					continue
				}

				for _, regex := range tagRegexes {
					r, err := regexp.Compile(fmt.Sprintf("(?i)%s", regex.Regex))
					if err != nil {
						logger.Error(fmt.Sprintf("regex compile %d: %s", regex.ID, err.Error()))
						continue
					}
					result := r.FindStringSubmatch(transaction.Memo)
					if len(result) > 0 {
						tagTransaction := model.TagTransaction{
							TagID:          tag.ID,
							TransactionsID: transaction.ID,
						}
						if err := tagTransaction.Insert(ctx, db, boil.Infer()); err != nil {
							logger.Error(err.Error())
							continue
						}
					}
				}
			}
		}
	}
}

func (t *tagService) SendTagRequest() types.TagResponse {
	var resp types.TagResponse
	t.TagRequestC <- TagRequest{}
	resp.Success = true
	resp.Msg = "Request sent"
	return resp
}

func (t *tagService) GetTag(id int64) types.TagResponse {
	var resp types.TagResponse

	tag, err := model.Tags(qm.Where("id=?", id)).One(t.ctx, t.db)
	switch {
	case dberror.IsNoRowsFound(err):
		resp.Msg = "Tag not found"
		t.logger.Error(resp.Msg)
		return resp
	case err != nil:
		resp.Msg = fmt.Sprintf("Failed to find Tag: %s", err.Error())
		t.logger.Error(resp.Msg)
		return resp
	}

	resp.Data, err = types.NewTag(t.ctx, t.db, tag)
	if err != nil {
		resp.Msg = err.Error()
		t.logger.Error(resp.Msg)
		return resp
	}

	resp.Success = true
	return resp
}

func (t *tagService) CreateTag(name, description string) types.TagResponse {
	var resp types.TagResponse

	tx, err := t.db.BeginTx(t.ctx, nil)
	if err != nil {
		resp.Msg = err.Error()
		t.logger.Error(resp.Msg)
		return resp
	}
	defer tx.Rollback()

	tag := &model.Tag{Name: name}
	if err := tag.Upsert(t.ctx, tx, true, []string{"name"}, boil.Whitelist("name"), boil.Infer()); err != nil {
		resp.Msg = err.Error()
		t.logger.Error(resp.Msg)
		return resp
	}

	if len(description) > 0 {
		attr := model.TagAttribute{
			TagID: tag.ID,
			Name:  "description",
			Value: description,
		}
		if err := attr.Insert(t.ctx, tx, boil.Infer()); err != nil {
			resp.Msg = err.Error()
			t.logger.Error(resp.Msg)
			return resp
		}
	}

	if err := tx.Commit(); err != nil {
		resp.Msg = err.Error()
		t.logger.Error(resp.Msg)
		return resp
	}

	resp.Data, err = types.NewTag(t.ctx, t.db, tag)
	if err != nil {
		resp.Msg = err.Error()
		t.logger.Error(resp.Msg)
		return resp
	}

	resp.Success = true
	resp.Msg = "Tag created"
	return resp
}

func (t *tagService) CreateTagRegex(id int64, regex string) types.TagRegexResponse {
	var resp types.TagRegexResponse
	tag, err := model.FindTag(t.ctx, t.db, id)
	if err != nil {
		resp.Msg = err.Error()
		t.logger.Error(resp.Msg)
		return resp
	}

	r := model.TagRegex{
		TagID: tag.ID,
		Regex: regex,
	}

	if err := r.Insert(t.ctx, t.db, boil.Infer()); err != nil {
		resp.Msg = err.Error()
		t.logger.Error(resp.Msg)
		return resp
	}

	resp.Success = true
	return resp
}

func (t *tagService) UndeleteTag(id int64) types.TagResponse {
	var resp types.TagResponse
	attr := model.TagAttribute{
		TagID: id,
		Name:  "deleted",
		Value: "false",
	}
	if err := attr.Upsert(t.ctx, t.db, true, []string{"tag_id", "name"}, boil.Whitelist("value"), boil.Infer()); err != nil {
		resp.Msg = err.Error()
		t.logger.Error(resp.Msg)
		return resp
	}
	resp.Success = true
	resp.Msg = "Tag undeleted"
	return resp
}

func (t *tagService) DeleteTag(id int64) types.TagResponse {
	var resp types.TagResponse
	attr := model.TagAttribute{
		TagID: id,
		Name:  "deleted",
		Value: "true",
	}
	if err := attr.Upsert(t.ctx, t.db, true, []string{"tag_id", "name"}, boil.Whitelist("value"), boil.Infer()); err != nil {
		resp.Msg = err.Error()
		t.logger.Error(resp.Msg)
		return resp
	}
	resp.Success = true
	resp.Msg = "Tag deleted"
	return resp
}

func (t *tagService) UpdateTag(id int64, name, description string) types.TagResponse {
	var resp types.TagResponse

	tx, err := t.db.BeginTx(t.ctx, nil)
	if err != nil {
		resp.Msg = err.Error()
		t.logger.Error(resp.Msg)
		return resp
	}
	defer tx.Rollback()

	tag, err := model.FindTag(t.ctx, tx, id)
	if err != nil {
		resp.Msg = err.Error()
		t.logger.Error(resp.Msg)
		return resp
	}
	tag.Name = name

	if _, err = tag.Update(t.ctx, tx, boil.Whitelist("name")); err != nil {
		resp.Msg = err.Error()
		t.logger.Error(resp.Msg)
		return resp
	}

	attr := model.TagAttribute{
		TagID: tag.ID,
		Name:  "description",
		Value: description,
	}
	if err = attr.Upsert(t.ctx, tx, true, []string{"tag_id", "name"}, boil.Whitelist("value"), boil.Infer()); err != nil {
		resp.Msg = err.Error()
		t.logger.Error(resp.Msg)
		return resp
	}

	if err := tx.Commit(); err != nil {
		resp.Msg = err.Error()
		t.logger.Error(resp.Msg)
		return resp
	}

	resp.Success = true
	resp.Msg = "Tag updated"

	return resp
}

func (t *tagService) GetTags() types.TagsResponse {
	var resp types.TagsResponse

	tags, err := model.Tags().All(t.ctx, t.db)
	if err != nil {
		resp.Msg = err.Error()
		t.logger.Error(resp.Msg)
		return resp
	}

	resp.Data = []types.Tag{}
	for _, i := range tags {
		tag, err := types.NewTag(t.ctx, t.db, i)
		if err != nil {
			resp.Msg = err.Error()
			t.logger.Error(resp.Msg)
			return resp
		}
		if tag.Deleted {
			continue
		}
		resp.Data = append(resp.Data, tag)
	}

	resp.Success = true
	return resp
}

func (t *tagService) DeleteTagRegex(id int64) types.TagRegexResponse {
	var resp types.TagRegexResponse
	if _, err := model.TagRegexes(qm.Where("id=?", id)).DeleteAll(t.ctx, t.db); err != nil {
		resp.Msg = "Failed to delete Tag Regex"
		t.logger.Error(resp.Msg)
		return resp
	}
	resp.Success = true
	resp.Msg = "Tag Filter deleted"
	return resp
}

func (t *tagService) UpdateTagRegex(id int64, regex string) types.TagRegexResponse {
	var resp types.TagRegexResponse
	tagRegex, err := model.FindTagRegex(t.ctx, t.db, id)
	if err != nil {
		resp.Msg = "Failed to find Tag Regex"
		t.logger.Error(resp.Msg)
		return resp
	}
	tagRegex.Regex = regex
	if _, err := tagRegex.Update(t.ctx, t.db, boil.Infer()); err != nil {
		resp.Msg = "Failed to update Tag Regex"
		t.logger.Error(resp.Msg)
		return resp
	}

	resp.Success = true
	resp.Msg = "Tag Filter updated"
	return resp
}
