package project

type Project struct {
	Id            int    `json:"id" db:"id"`
	UserId        string `json:"user_id" db:"user_id"`
	Name          string `json:"name" db:"name"`
	SourceCodeUrl string `json:"source_code_url" db:"source_code_url"`
	SourceCode    string `json:"source_code" db:"source_code"`
	Subdomain     string `json:"subdomain" db:"subdomain"`
	CustomDomain  string `json:"custom_domain" db:"custom_domain"`
	Language      string `json:"language" db:"language"`
	IsDockerized  string `json:"is_dockerized" db:"is_dockerized"`
	CreatedAt     string `json:"created_on" db:"created_at"`
}
