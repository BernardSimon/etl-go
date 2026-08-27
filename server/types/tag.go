package types

// ── 标签 ──────────────────────────────────────────────────────────────────────

type AddTagRequest struct {
	Name  string `json:"name" binding:"required"`
	Color string `json:"color"`
}

type UpdateTagBody struct {
	Name  string `json:"name" binding:"required"`
	Color string `json:"color"`
}

type GetTagListRequest struct {
	Search string `form:"search"`
}
