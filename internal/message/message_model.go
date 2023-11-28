package message

import "gorm.io/gorm"

type Message struct {
	gorm.Model
	Text   string `json:"text"`
	Sender string `json:"sender"`
	Room   string `json:"room"`
}
