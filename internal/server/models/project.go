package models

type CreateProjectReq struct {
	Path string `json:"path"`
	Name string `json:"name"`
}
