package tree

type TreeAlbum struct {
	ID        int    `json:"album_id"`
	Title     string `json:"title"`
	IconURL   string `json:"icon_url"`
	IsDeleted bool   `json:"-"`
}

type Tree struct {
	ID        int    `json:"tree_id"`
	AlbumID   int    `json:"album_id"`
	Name      string `json:"name"`
	Variety   string `json:"variety"`
	PlantedAt string `json:"planted_at"`
}

type TreeNode struct {
	ID         int    `json:"node_id"`
	TreeID     int    `json:"tree_id"`
	Title      string `json:"title"`
	Reflection string `json:"reflection"`
	ImageURL   string `json:"image_url"`
}

type CreateAlbumRequest struct {
	Title   string `json:"title" binding:"required"`
	IconURL string `json:"icon_url"`
}

type UpdateAlbumRequest struct {
	Title   string `json:"title"`
	IconURL string `json:"icon_url"`
}

type CreateTreeRequest struct {
	Name    string `json:"name" binding:"required"`
	Variety string `json:"variety"`
}

type CreateNodeRequest struct {
	Title      string `json:"title" binding:"required"`
	Reflection string `json:"reflection"`
	ImageURL   string `json:"image_url"`
}

type UpdateNodeRequest struct {
	Title      string `json:"title"`
	Reflection string `json:"reflection"`
	ImageURL   string `json:"image_url"`
}