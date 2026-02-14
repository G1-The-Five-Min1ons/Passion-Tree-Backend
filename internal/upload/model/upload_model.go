package model

type UploadImageRequest struct {
	Filename string `json:"filename"`
	Folder   string `json:"folder"`
}

type UploadImageResponse struct {
	UploadURL string `json:"upload_url"`
	PublicURL string `json:"public_url"`
	ExpiresIn string `json:"expires_in"`
	Folder    string `json:"folder"`
}
