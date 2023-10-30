package model

type (
	SwitchMode struct {
		Mode int `json:"mode"` // 0-stop  1-start
	}

	UpdateNftRequest struct {
		Keys   []string `json:"keys"`
		Values []string `json:"values"`
		Type   int      `json:"type"` // 0-delete 1-add
	}
)
