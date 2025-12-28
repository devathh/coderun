package auth

type JWTManager interface {
	Validate(tokenString string) (*CoderunClaims, error)
}
