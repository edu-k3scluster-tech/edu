package app

import "time"

type User struct {
	Id    string `db:"id"`
	TgId  string `db:"tg_id"`
	Token string `db:"auth_token"`
}

type OneTimeToken struct {
	UserId string `db:"user_id"`
	Token  string `db:"token"`
}

type AuditLog struct {
	UserId    string    `db:"user_id"`
	Action    string    `db:"action"`
	CreatedAt time.Time `db:"created_at"`
}
