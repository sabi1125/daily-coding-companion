package response

type SettingResponse struct {
	SettingId          string  `json:"setting_id"`
	Email              string  `json:"email"`
	SolvedCount        int64   `json:"solved_count"`
	UnsolvedCount      int64   `json:"unsolved_count"`
	GetHelpPreferences *string `json:"get_help_preferences"`
	NeedsReauth        bool    `json:"needs_reauth"`
}
