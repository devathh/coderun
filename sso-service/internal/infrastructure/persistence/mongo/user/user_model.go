package usermongo

type UserModel struct {
	ID           string `bson:"id"`
	Username     string `bson:"username"`
	Email        string `bson:"email"`
	PasswordHash string `bson:"password_hash"`
}
