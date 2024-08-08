package types

import (
	"pennyapp/backend/model"
)

type NoteResponse struct {
	Success bool   `json:"success"`
	Msg     string `json:"msg"`
	Data    Note   `json:"data"`
}

type NotesResponse struct {
	Success bool   `json:"success"`
	Msg     string `json:"msg"`
	Data    []Note `json:"data"`
}

type Note struct {
	ID   int64  `json:"id"`
	Note string `json:"note"`
}

func NewNote(n *model.Note) Note {
	note := Note{ID: n.ID, Note: n.Note}
	return note
}
