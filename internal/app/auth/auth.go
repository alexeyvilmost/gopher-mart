package auth

import (
	"fmt"
	"gophermart/internal/app/storage"
	"hash/fnv"
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
	Storage   storage.Storage
	SecretKey string
}

const tokenExp = time.Hour * 3

func Hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

func (a Auth) BuildJWTString(userID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenExp)),
		},
		UserID: userID,
	})

	tokenString, err := token.SignedString([]byte(a.SecretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (a Auth) GetUserID(tokenString string) (string, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(a.SecretKey), nil
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

func (a Auth) AddAuth(w http.ResponseWriter, userID string) error {
	token, err := a.BuildJWTString(userID)
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

func (a Auth) WithAuth(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

		userID, err := a.GetUserID(token)
		if err != nil {
			log.Error().Err(err).Msg("Error while get user_id from token")
			http.Error(w, "Unexpected error while get user_id from token", http.StatusInternalServerError)
			return
		}
		_, err = a.Storage.GetUser(r.Context(), userID)
		if err != nil {
			if err == storage.ErrEmpty {
				log.Error().Err(err).Msg("No such user")
				http.Error(w, "No such user", http.StatusUnauthorized)
				return
			}
			log.Error().Err(err).Msg("Error while checking user existance")
			http.Error(w, "Unexpected error while checking user existance", http.StatusInternalServerError)
			return
		}
		r.Header.Set("x-user-id", userID)
		log.Info().Str("user_id", userID).Msg("")
		h.ServeHTTP(w, r)
	})
}
