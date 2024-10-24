package auth

import (
	"fmt"
	"gophermart/internal/app/storage"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/rs/zerolog/log"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

type Auth struct {
	Storage storage.Storage
}

const tokenExp = time.Hour * 3
const secretKey = "supersecretkey"

func BuildJWTString(userID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenExp)),
		},
		UserID: userID,
	})

	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func GetUserID(tokenString string) (string, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(secretKey), nil
		})
	if err != nil {
		return "", err
	}

	if !token.Valid {
		fmt.Println("Token is not valid")
		return "", fmt.Errorf("invalid token")
	}

	return claims.UserID, nil
}

func AddAuth(w http.ResponseWriter, userID string) error {
	token, err := BuildJWTString(userID)
	if err != nil {
		return err
	}

	cookie := &http.Cookie{
		Path:   "/",
		Name:   "jwt_auth",
		Value:  token,
		MaxAge: 300,
	}

	http.SetCookie(w, cookie)
	return nil
}

func (a Auth) WithAuth(h http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Info().Msg("da")
		var token string
		jwtAuth, err := r.Cookie("jwt_auth")
		if err != nil {
			if err == http.ErrNoCookie {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			http.Error(w, "Unexpected error while getting auth cookie", http.StatusInternalServerError)
			return
		}
		token = jwtAuth.Value

		userID, err := GetUserID(token)
		if err != nil {
			log.Error().Err(err).Msg("Error while get user_id from token")
			http.Error(w, "Unexpected error while get user_id from token", http.StatusInternalServerError)
			return
		}
		r.Header.Set("user-id-auth", userID)
		log.Info().Str("user_id", userID).Msg("")
		h.ServeHTTP(w, r)
	})
}
