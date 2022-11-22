package chat

type Message struct {
	ID      string `bson:"_id"`
	Content string
	ChatID  string
	User    User
}
