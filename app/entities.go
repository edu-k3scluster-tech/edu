package app

import "time"

type UserStatus string

const (
	UserStatusNew         UserStatus = "new"
	UserStatusActive      UserStatus = "active"
	UserStatusDeactivated UserStatus = "deactivated"
)

type User struct {
	Id         int        `db:"id"`
	TgId       *int64     `db:"tg_id"`
	TgUsername *string    `db:"tg_username"`
	Status     UserStatus `db:"status"`
	IsStaff    bool       `db:"is_staff"`
	CreatedAt  time.Time  `db:"created_at"`
}

type AuditLog struct {
	UserId    int       `db:"user_id"`
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

type UserCertificate struct {
	UserId      int       `db:"user_id"`
	Username    string    `db:"username"`
	Certificate string    `db:"certificate"`
	PrivateKey  string    `db:"private_key"`
	CreatedAt   time.Time `db:"created_at"`
}
