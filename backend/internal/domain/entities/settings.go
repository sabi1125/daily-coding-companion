package entities

type Settings struct {
	SettingID          string  `json:"setting_id" gorm:"column:setting_id"`
	UserID             string  `json:"-" gorm:"column:user_id"`
	GetHelpPreferences *string `json:"get_help_preferences" gorm:"column:get_help_preferences"`

	NeedsReauth bool `json:"needs_reauth" gorm:"-"`
}

type UpdateSettingBody struct {
	GetHelpPreferences string `json:"get_help_preferences" validate:"required,min=1,max=1000"`
}
