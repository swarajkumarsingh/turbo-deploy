package deployment

type DeploymentBody struct {
	ProjectId string `validate:"required" json:"projectId"`
}

type Deployment struct {
	Id        int    `json:"id" db:"id"`
	UserId    string `json:"user_id" db:"user_id"`
	ProjectId string `json:"project_id" db:"project_id"`
	Duration  string `json:"duration" db:"duration"`
	ReadUrl   string `json:"ready_url" db:"ready_url"`
	LastLog   string `json:"last_log" db:"last_log"`
	Status    string `json:"status" db:"status"`
	CreatedAt string `json:"created_on" db:"created_at"`
	UpdatedAt string `json:"updated_at" db:"updated_at"`
}
