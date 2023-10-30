package model

type (
	CreateWalletRequest struct {
		Creator     string `json:"creator"`
		OwnersCount int64  `json:"owners_count"`
		Threshold   int64  `json:"threshold"`
		Name        string `json:"name"`
	}

	AddOwnerRequest struct {
		Creator string `json:"id"`
		Wallet  string `json:"wallet"`
	}

	InitWalletRequest struct {
		Creator string `json:"id"`
	}
)
