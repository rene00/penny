package ofximport

import (
	"context"
	"database/sql"

	"github.com/aclindsa/ofxgo"
)

type OfxImport struct {
	resp *ofxgo.Response
}

func NewOfxImport(resp *ofxgo.Response) OfxImport {
	return OfxImport{resp: resp}
}

func (o OfxImport) processBank(ctx context.Context, db *sql.DB, msg ofxgo.Message) error {
	return nil
}

func (o OfxImport) Do(ctx context.Context, db *sql.DB) error {
	for _, i := range o.resp.Bank {
		if err := o.processBank(ctx, db, i); err != nil {
			return err
		}
	}
	return nil
}
