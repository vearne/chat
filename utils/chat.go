package utils

import (
	"github.com/imroc/req"
)

type IdResult struct {
	IdList []uint64 `json:"id_list"`
}

func GetChatId() uint64 {
	url := "http://id.vearne.cc/v1/nextid?tag=chat&count=1"
	r, _ := req.Get(url)
	var res IdResult
	r.ToJSON(&res)
	return res.IdList[0]
}
