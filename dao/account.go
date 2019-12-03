package dao

import (
	"github.com/vearne/chat/model"
	"github.com/vearne/chat/resource"
)

func GetAccount(accountId uint64) *model.Account {
	var account model.Account
	resource.MySQLClient.Where("id = ?", accountId).First(&account)
	return &account
}
