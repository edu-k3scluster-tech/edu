package app

import "time"

type User struct {
	Id         int     `db:"id"`
	TgId       *int64  `db:"tg_id"`
	TgUsername *string `db:"tg_username"`
}

type AuditLog struct {
	UserId    string    `db:"user_id"`
	Action    string    `db:"action"`
	CreatedAt time.Time `db:"created_at"`
}

type AuthToken struct {
	UserId    int       `db:"user_id"`
	Token     string    `db:"token"`
	CreatedAt time.Time `db:"created_at"`
}

type TgOneTimeToken struct {
	Token     string    `db:"token"`
	UserId    *int      `db:"user_id"`
	CreatedAt time.Time `db:"created_at"`
}
