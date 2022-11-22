package chat

type User struct {
	ID   string `bson:"_id"`
	Name string
}