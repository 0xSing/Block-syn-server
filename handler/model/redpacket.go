package model

type (
	GetRPByHashReq struct {
		Hash string `json:"hash"`
	}

	GetRPByIdReq struct {
		Id string `json:"id"`
	}

	GetRPShareUri struct {
		Hash string `json:"hash"`
	}

	CheckRPStates struct {
		Hash string `json:"hash"`
	}

	GetRandomAmountReq struct {
		Id string `json:"id"`
	}

	GetClaimSignReq struct {
		Id       string `json:"id"`
		Receiver string `json:"receiver""`
		Amount   string `json:"amount"`
	}

	GetCreateRPsReq struct {
		Owner string `json:"owner"`
	}

	GetClaimRPsReq struct {
		Owner string `json:"owner"`
	}

	GetCreateRPRByIdReq struct {
		Owner string `json:"owner"`
		Id    string `json:"id"`
	}

	GetClaimRPRByIdReq struct {
		Owner string `json:"owner"`
		Id    string `json:"id"`
	}
)
