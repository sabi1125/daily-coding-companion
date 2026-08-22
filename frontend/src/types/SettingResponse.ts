export interface SettingResponse {
  setting_id: string
  email: string
  solved_count: number
  unsolved_count: number
  get_help_preferences: string | null
  needs_reauth: boolean
}

export interface PatchSettingRequest {
  get_help_preferences: string
}
