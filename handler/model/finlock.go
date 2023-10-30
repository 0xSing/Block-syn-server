package model

type (
	GetOwnerLocks struct {
		Owner string `json:"owner"`
	}

	ChangeLockStates struct {
		Owner string `json:"owner"`
		Id    string `json:"id"`
	}
)
