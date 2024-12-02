package deployment

type DeploymentBody struct {
	ProjectId  string `validate:"required" json:"projectId"`
}