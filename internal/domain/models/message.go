package models

type Message struct {
	ID       int64
	Content  string
	UserID   int64
	Type     string
	DateTime string
}
