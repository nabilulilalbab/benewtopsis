// models/userModel.go
package models

// User represents a user in the system
type User struct {
	Id              int64  `gorm:"primaryKey;autoIncrement" json:"id"               example:"1"`
	NameLengkap     string `                                json:"nama_lengkap"     example:"John Doe"`
	Email           string `gorm:"unique"                   json:"email"            example:"john@example.com"`
	Password        string `                                json:"password"         example:"hashed_password"`
	ConfirmPassword string `                                json:"confirm_password" example:"hashed_password"`
}
