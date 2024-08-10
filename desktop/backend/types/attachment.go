package types

import "pennyapp/backend/model"

type AttachmentResponse struct {
	Success bool       `json:"success"`
	Msg     string     `json:"msg"`
	Data    Attachment `json:"data"`
}

type AttachmentsResponse struct {
	Success bool         `json:"success"`
	Msg     string       `json:"msg"`
	Data    []Attachment `json:"data"`
}

type Attachment struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Size int64  `json:"size"`
	Data []byte `json:"data"`
}

func NewAttachment(a *model.Attachment) (Attachment, error) {
	attachment := Attachment{
		ID:   a.ID,
		Name: a.Name,
		Size: a.SZ,
		Data: a.Data,
	}
	return attachment, nil
}
