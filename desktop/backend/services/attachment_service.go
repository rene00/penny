package services

import (
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"pennyapp/backend/config"
	"pennyapp/backend/internal/dberror"
	"pennyapp/backend/logwrap"
	"pennyapp/backend/model"
	"pennyapp/backend/types"
	"strings"
	"time"

	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type attachmentService struct {
	ctx    context.Context
	db     *sql.DB
	conf   config.Config
	logger *logwrap.LogWrap
}

func Attachment() *attachmentService {
	return &attachmentService{}
}

func (a *attachmentService) Start(ctx context.Context, conf config.Config, db *sql.DB, logger *logwrap.LogWrap) {
	a.ctx = ctx
	a.db = db
	a.conf = conf
	a.logger = logger
}

func (a *attachmentService) Shutdown(ctx context.Context, conf config.Config, logger *logwrap.LogWrap) {
	// Delete temporary attachments from tmpdir.
	filepath.Walk(conf.TmpDir, func(path string, info os.FileInfo, err error) error {
		filename := filepath.Base(path)
		if strings.HasPrefix(filename, "attachment_") && !info.IsDir() {
			if err := os.Remove(path); err != nil {
				a.logger.Error(fmt.Sprintf("failed to remove tmp attachment %s: %s", path, err.Error()))
			}
		}
		return nil
	})
}

func (a *attachmentService) CreateAttachment(transactionID int64, name, data string) types.AttachmentResponse {
	var resp types.AttachmentResponse
	var err error

	tx, err := a.db.BeginTx(a.ctx, nil)
	if err != nil {
		resp.Msg = err.Error()
		a.logger.Error(resp.Msg)
		return resp
	}
	defer tx.Rollback()

	transaction, err := model.FindTransaction(a.ctx, tx, transactionID)
	if err != nil {
		resp.Msg = err.Error()
		a.logger.Error(resp.Msg)
		return resp
	}

	sArray := strings.Split(data, ",")
	if len(sArray) != 2 {
		resp.Msg = "Failed to parse Attachment"
		a.logger.Error(resp.Msg)
		return resp
	}

	attachmentBlob := sArray[1]

	decoded, err := base64.StdEncoding.DecodeString(attachmentBlob)
	if err != nil {
		resp.Msg = fmt.Sprintf("Failed to decode: %s", err.Error())
		a.logger.Error(resp.Msg)
		return resp
	}

	attachment := &model.Attachment{
		Name:      name,
		SZ:        int64(len(decoded)),
		Data:      []byte(attachmentBlob),
		CreatedAt: time.Now().UTC().Unix(),
	}

	err = attachment.Insert(a.ctx, tx, boil.Infer())
	switch {
	case dberror.IsUniqueConstraint(err):
		attachment, err = model.Attachments(qm.Where("name=? AND data=?", name, []byte(attachmentBlob))).One(a.ctx, tx)
		if err != nil {
			resp.Msg = err.Error()
			a.logger.Error(resp.Msg)
			return resp
		}
	case err != nil:
		resp.Msg = err.Error()
		a.logger.Error(resp.Msg)
		return resp
	}

	attachmentTransaction := model.AttachmentTransaction{
		AttachmentID:   attachment.ID,
		TransactionsID: transaction.ID,
	}
	if err = attachmentTransaction.Insert(a.ctx, tx, boil.Infer()); err != nil {
		resp.Msg = err.Error()
		a.logger.Error(resp.Msg)
		return resp
	}

	if err := tx.Commit(); err != nil {
		resp.Msg = err.Error()
		a.logger.Error(resp.Msg)
		return resp
	}

	resp.Success = true
	return resp
}

func (a *attachmentService) GetTransactionAttachments(id int64) types.AttachmentsResponse {
	var resp types.AttachmentsResponse

	transaction, err := model.FindTransaction(a.ctx, a.db, id)
	if err != nil {
		resp.Msg = err.Error()
		a.logger.Error(resp.Msg)
		return resp
	}

	q := []qm.QueryMod{
		qm.Where("transactions_id=?", transaction.ID),
		qm.InnerJoin("attachment ON attachment_transactions.attachment_id = attachment.id"),
		qm.Load("Attachment"),
	}
	attachmentTransactions, err := model.AttachmentTransactions(q...).All(a.ctx, a.db)
	if err != nil {
		resp.Msg = err.Error()
		a.logger.Error(resp.Msg)
		return resp
	}

	resp.Data = []types.Attachment{}
	for _, i := range attachmentTransactions {
		attachment, err := types.NewAttachment(i.R.Attachment)
		if err != nil {
			resp.Msg = err.Error()
			a.logger.Error(resp.Msg)
			return resp
		}
		resp.Data = append(resp.Data, attachment)
	}

	resp.Success = true
	return resp
}

func (a *attachmentService) GetAttachments() types.AttachmentsResponse {
	var resp types.AttachmentsResponse
	resp.Success = true
	return resp
}

func (a *attachmentService) GetAttachment(id int64) types.AttachmentResponse {
	var resp types.AttachmentResponse

	attachment, err := model.FindAttachment(a.ctx, a.db, id)
	if err != nil {
		resp.Msg = err.Error()
		a.logger.Error(resp.Msg)
		return resp
	}

	resp.Data, err = types.NewAttachment(attachment)
	if err != nil {
		resp.Msg = err.Error()
		a.logger.Error(resp.Msg)
		return resp
	}

	filePath := filepath.Join(a.conf.TmpDir, fmt.Sprintf("attachment_%d", resp.Data.ID))
	f, err := os.Create(filePath)
	if err != nil {
		resp.Msg = err.Error()
		a.logger.Error(resp.Msg)
		return resp
	}
	defer f.Close()

	if err = os.Chmod(filePath, 0600); err != nil {
		resp.Msg = err.Error()
		a.logger.Error(resp.Msg)
		return resp
	}

	// resp.Data.Data is base64 encoded so will need to decode to []byte and write.
	data, err := base64.StdEncoding.DecodeString(string(resp.Data.Data))
	if err != nil {
		resp.Msg = err.Error()
		a.logger.Error(resp.Msg)
		return resp
	}

	if _, err := f.Write(data); err != nil {
		resp.Msg = err.Error()
		a.logger.Error(resp.Msg)
		return resp
	}

	cmd := exec.Command("/usr/bin/open", f.Name())
	if err := cmd.Run(); err != nil {
		resp.Msg = err.Error()
		a.logger.Error(resp.Msg)
		return resp
	}

	resp.Success = true
	return resp
}
