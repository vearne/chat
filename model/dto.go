package model

type CommonCmd struct {
	Cmd string `json:"cmd"`
}

type CmdCreateAccountReq struct {
	Cmd      string `json:"cmd"`
	NickName string `json:"nickName"`
}

type CmdCreateAccountResp struct {
	Cmd       string `json:"cmd"`
	NickName  string `json:"nickName"`
	AccountId uint64 `json:"accountId"`
}

type CmdMatchReq struct {
	Cmd       string `json:"cmd"`
	AccountId uint64 `json:"accountId"`
}

type CmdMatchResp struct {
	Cmd         string `json:"cmd"`
	PartnerId   uint64 `json:"partnerId,omitempty"`
	PartnerName string `json:"partnerName,omitempty"`
	SessionId   uint64 `json:"sessionId,omitempty"`
	Code        int32 `json:"code"`
}
