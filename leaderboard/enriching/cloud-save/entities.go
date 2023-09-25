package cloud_save

type CloudSaveGetProfilesRequest struct {
	TenantID string   `json:"tenant_id"`
	IDs      []string `json:"public_account_ids"`
}

type CloudSaveGetProfilesResponse struct {
	Documents []CloudSavePlayerEntity `json:"documents"`
}

type CloudSavePlayerEntity struct {
	AccountID string            `json:"accountId"`
	Data      map[string]string `json:"data"`
}

type CloudSavePlayerData struct {
	PictureURL string `json:"picture_url"`
	Name       string `json:"player_name"`
}
