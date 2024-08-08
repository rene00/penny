package services

import (
	"context"
	"database/sql"
	"pennyapp/backend/internal/dberror"
	"pennyapp/backend/logwrap"
	"pennyapp/backend/model"
	"pennyapp/backend/types"

	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type noteService struct {
	ctx    context.Context
	db     *sql.DB
	logger *logwrap.LogWrap
}

func Note() *noteService {
	return &noteService{}
}

func (n *noteService) Start(ctx context.Context, db *sql.DB, logger *logwrap.LogWrap) {
	n.ctx = ctx
	n.db = db
	n.logger = logger
}

func (n *noteService) GetTransactionNote(id int64) types.NoteResponse {
	var resp types.NoteResponse

	transaction, err := model.FindTransaction(n.ctx, n.db, id)
	if err != nil {
		resp.Msg = err.Error()
		n.logger.Error(resp.Msg)
		return resp
	}

	resp.Data = types.Note{}

	q := []qm.QueryMod{
		qm.Where("transactions_id=?", transaction.ID),
		qm.InnerJoin("note on note_transactions.note_id = note.id"),
		qm.Load("Note"),
	}
	noteTransaction, err := model.NoteTransactions(q...).One(n.ctx, n.db)
	switch {
	case dberror.IsNoRowsFound(err):
	case err != nil:
		resp.Msg = err.Error()
		n.logger.Error(resp.Msg)
		return resp
	default:
		resp.Data = types.NewNote(noteTransaction.R.Note)
	}

	resp.Success = true
	return resp
}

func (n *noteService) DeleteTransactionNote(id int64) types.NoteResponse {
	var resp types.NoteResponse
	var err error

	tx, err := n.db.BeginTx(n.ctx, nil)
	if err != nil {
		resp.Msg = err.Error()
		n.logger.Error(resp.Msg)
		return resp
	}
	defer tx.Rollback()

	note, err := model.FindNote(n.ctx, tx, id)
	if err != nil {
		resp.Msg = err.Error()
		n.logger.Error(resp.Msg)
		return resp
	}

	noteTransaction, err := model.NoteTransactions(qm.Where("note_id=?", note.ID)).One(n.ctx, tx)
	if err != nil {
		resp.Msg = err.Error()
		n.logger.Error(resp.Msg)
		return resp
	}

	if _, err = noteTransaction.Delete(n.ctx, tx); err != nil {
		resp.Msg = err.Error()
		n.logger.Error(resp.Msg)
		return resp
	}

	if _, err = note.Delete(n.ctx, tx); err != nil {
		resp.Msg = err.Error()
		n.logger.Error(resp.Msg)
		return resp
	}

	if err = tx.Commit(); err != nil {
		resp.Msg = err.Error()
		n.logger.Error(resp.Msg)
		return resp
	}

	resp.Success = true
	return resp
}

func (n *noteService) UpdateTransactionNote(id int64, s string) types.NoteResponse {
	var resp types.NoteResponse
	var err error

	transaction, err := model.FindTransaction(n.ctx, n.db, id)
	if err != nil {
		resp.Msg = err.Error()
		n.logger.Error(resp.Msg)
		return resp
	}

	noteTransaction, err := model.NoteTransactions(qm.Where("transactions_id=?", transaction.ID)).One(n.ctx, n.db)
	if err != nil {
		resp.Msg = err.Error()
		n.logger.Error(resp.Msg)
		return resp
	}

	note, err := model.FindNote(n.ctx, n.db, noteTransaction.NoteID)
	if err != nil {
		resp.Msg = err.Error()
		n.logger.Error(resp.Msg)
		return resp
	}

	note.Note = s
	if _, err = note.Update(n.ctx, n.db, boil.Whitelist("note")); err != nil {
		resp.Msg = err.Error()
		n.logger.Error(resp.Msg)
		return resp
	}

	resp.Success = true
	return resp
}

func (n *noteService) CreateTransactionNote(id int64, note string) types.NoteResponse {
	var resp types.NoteResponse
	var err error

	tx, err := n.db.BeginTx(n.ctx, nil)
	if err != nil {
		resp.Msg = err.Error()
		n.logger.Error(resp.Msg)
		return resp
	}
	defer tx.Rollback()

	transaction, err := model.FindTransaction(n.ctx, tx, id)
	if err != nil {
		resp.Msg = err.Error()
		n.logger.Error(resp.Msg)
		return resp
	}

	newNote := model.Note{Note: note}
	if err = newNote.Insert(n.ctx, tx, boil.Infer()); err != nil {
		resp.Msg = err.Error()
		n.logger.Error(resp.Msg)
		return resp
	}

	noteTransaction := model.NoteTransaction{
		NoteID:         newNote.ID,
		TransactionsID: transaction.ID,
	}
	if err = noteTransaction.Insert(n.ctx, tx, boil.Infer()); err != nil {
		resp.Msg = err.Error()
		n.logger.Error(resp.Msg)
		return resp
	}

	if err = tx.Commit(); err != nil {
		resp.Msg = err.Error()
		n.logger.Error(resp.Msg)
		return resp
	}

	resp.Success = true
	return resp
}
