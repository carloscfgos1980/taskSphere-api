package utils

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type TokenType string

const (
	// TokenTypeAccess -
	TokenTypeAccess TokenType = "taskSphere-access-token"
)

// ErrNoAuthHeaderIncluded -
var ErrNoAuthHeaderIncluded = errors.New("no auth header included in request")

// HashPassword -
func HashPassword(password string) (string, error) {
	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return "", err
	}
	return hash, nil
}

// CheckPasswordHash -
func CheckPasswordHash(password, hash string) (bool, error) {
	match, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		return false, err
	}
	return match, nil
}

// MakeJWT -
func MakeJWT(
	userID uuid.UUID,
	tokenSecret string,
	expiresIn time.Duration,
) (string, error) {
	signingKey := []byte(tokenSecret)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    string(TokenTypeAccess),
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(expiresIn)),
		Subject:   userID.String(),
	})
	return token.SignedString(signingKey)
}

// ValidateJWT -
func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	claimsStruct := jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(
		tokenString,
		&claimsStruct,
		func(token *jwt.Token) (interface{}, error) { return []byte(tokenSecret), nil },
	)
	if err != nil {
		return uuid.Nil, err
	}

	userIDString, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.Nil, err
	}

	issuer, err := token.Claims.GetIssuer()
	if err != nil {
		return uuid.Nil, err
	}
	if issuer != string(TokenTypeAccess) {
		return uuid.Nil, errors.New("invalid issuer")
	}

	id, err := uuid.Parse(userIDString)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid user ID: %w", err)
	}
	return id, nil
}

// GetBearerToken -
func GetBearerToken(c *gin.Context) (string, error) {
	// Extract the Authorization header from the incoming request
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
		c.Abort()
		return "", ErrNoAuthHeaderIncluded
	}
	// Check if the Authorization header is in the correct format (e.g., "Bearer <token>")
	splitAuth := strings.Split(authHeader, " ")
	if len(splitAuth) < 2 || splitAuth[0] != "Bearer" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "malformed authorization header"})
		c.Abort()
		return "", errors.New("malformed authorization header")
	}
	return splitAuth[1], nil
}

// MakeRefreshToken makes a random 256 bit token
// encoded in hex
func MakeRefreshToken() string {
	token := make([]byte, 32)
	rand.Read(token)
	return hex.EncodeToString(token)
}

// strong password function that checks for length and complexity
func IsStrongPassword(password string) error {
	if len(password) < 8 {
		return errors.New("password is too short")
	}
	var hasUpper, hasLower, hasNumber, hasSpecial bool
	for _, char := range password {
		switch {
		case 'A' <= char && char <= 'Z':
			hasUpper = true
		case 'a' <= char && char <= 'z':
			hasLower = true
		case '0' <= char && char <= '9':
			hasNumber = true
		case strings.ContainsRune("!@#$%^&*()-_=+[]{}|;:,.<>?/", char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return errors.New("password must contain at least one uppercase letter")
	}
	if !hasLower {
		return errors.New("password must contain at least one lowercase letter")
	}
	if !hasNumber {
		return errors.New("password must contain at least one number")
	}
	if !hasSpecial {
		return errors.New("password must contain at least one special character")
	}
	return nil
}

// IsValidEmail checks if an email has a valid format.
func IsValidEmail(email string) error {
	email = strings.TrimSpace(email)
	if email == "" {
		return errors.New("email is required")
	}

	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Address != email {
		return errors.New("invalid email format")
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return errors.New("invalid email format")
	}

	if !strings.Contains(parts[1], ".") {
		return errors.New("invalid email format")
	}

	return nil
}

func CheckPriority(priority string) (resultPriority string, err error) {
	switch priority {
	case "":
		return "medium", nil
	case "low", "medium", "high", "urgent":
		return priority, nil
	default:
		return "", fmt.Errorf("Invalid priority value: %s", priority)
	}
}

func CheckState(state string) (resultState string, err error) {
	switch state {
	case "":
		return "pending", nil
	case "pending", "in progress", "done", "cancelled":
		return state, nil
	default:
		return "", fmt.Errorf("Invalid state value: %s", state)
	}
}

func CheckTag(tag string) (resultTag string, err error) {
	switch tag {
	case "":
		return "private", nil
	case "private", "collaborative", "public":
		return tag, nil
	default:
		return "", fmt.Errorf("Invalid tag value: %s", tag)
	}
}
