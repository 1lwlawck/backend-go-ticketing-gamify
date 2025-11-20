package users

import "time"

// User represents a workspace user profile.
type User struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Username  string    `json:"username"`
	Role      string    `json:"role"`
	AvatarURL string    `json:"avatarUrl"`
	Badges    []string  `json:"badges"`
	Bio       string    `json:"bio"`
	CreatedAt time.Time `json:"createdAt"`
}

// UpdateProfileInput captures editable profile fields.
type UpdateProfileInput struct {
	Name      string `json:"name"`
	Bio       string `json:"bio"`
	AvatarURL string `json:"avatarUrl"`
}
