package entities

type Settings struct {
	SettingID          string  `json:"setting_id" gorm:"column:setting_id"`
	UserID             string  `json:"user_id" gorm:"column:user_id"`
	GetHelpPreferences *string `json:"get_help_preferences" gorm:"column:get_help_preferences"`
}
