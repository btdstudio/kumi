package models

import (
	"time"

	"github.com/uptrace/bun"
)

type User struct {
	bun.BaseModel `bun:"table:users,alias:u"`

	ID           string    `bun:"id,pk,type:uuid,default:gen_random_uuid()"`
	Email        string    `bun:"email,unique,notnull"`
	PasswordHash string    `bun:"password_hash,notnull"`
	DisplayName  string    `bun:"display_name"`
	AvatarURL    string    `bun:"avatar_url"`
	Theme        string    `bun:"theme,notnull,default:'catppuccin-mocha'"`
	PendingEmail string    `bun:"pending_email"`
	CreatedAt    time.Time `bun:"created_at,notnull,default:current_timestamp"`
}
