package dao

import (
	"github.com/vearne/chat/resource"
	"time"
)

func CreatOrUpdateViewedAck(sessionId uint64, accountId uint64, msgId uint64) {
	sql := "INSERT INTO `view_ack` (`session_id`, `account_id`, `msg_id`, `created_at`) "
	sql += "VALUES (?, ?, ?, ?) ON DUPLICATE KEY UPDATE msg_id = ?, created_at = ?; "
	resource.MySQLClient.Exec(sql, sessionId, accountId, msgId, time.Now(), msgId, time.Now())
}
