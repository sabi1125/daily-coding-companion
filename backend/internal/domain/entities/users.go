package entities

type Users struct {
	UserId    string `json:"user_id" gorm:"column:user_id"`
	Email     string `json:"email" gorm:"column:email"`
	FirstName string `json:"firstname" gorm:"column:firstname"`
	LastName  string `json:"lastname" gorm:"column:lastname"`
}
