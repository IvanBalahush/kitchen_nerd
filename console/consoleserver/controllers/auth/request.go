package auth

// RegistrationRequest contains user registration fields.
type RegistrationRequest struct {
	Email            string `json:"email"`
	UserName         string `json:"username"`
	Password         string `json:"password"`
	RepeatedPassword string `json:"repeatedPassword"`
}

// LoginRequest contains user registration fields.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// IsValid check the request for all conditions.
func (uf *RegistrationRequest) IsValid() bool {
	switch {
	case uf.UserName == "":
		return false
	case uf.Email == "":
		return false
	case uf.Password == "":
		return false
	case uf.Password != uf.RepeatedPassword:
		return false
	default:
		return true
	}
}

func (uf *LoginRequest) IsValid() bool {
	switch {
	case uf.Email == "":
		return false
	case uf.Password == "":
		return false
	default:
		return true
	}
}
