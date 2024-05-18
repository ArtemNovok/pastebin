package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type ReqUserData struct {
	UserData map[string]string
}

const secretkey = "secret"

func (app *Config) ValidateRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenStr := r.Header.Get("Authorization")
		val := ReqUserData{
			UserData: map[string]string{},
		}
		if tokenStr == "" {
			val.UserData["Authorized"] = "0"
			ctx := context.WithValue(r.Context(), "userdata", val)
			next.ServeHTTP(w, r.WithContext(ctx))
			return

		}
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(secretkey), nil
		})
		if err != nil {
			val.UserData["Authorized"] = "0"
			ctx := context.WithValue(r.Context(), "userdata", val)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			if float64(time.Now().Unix()) > claims["exp"].(float64) {
				val.UserData["Authorized"] = "0"
				ctx := context.WithValue(r.Context(), "userdata", val)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
			userId, ok := claims["sub"].(float64)
			if !ok {
				val.UserData["Authorized"] = "0"
				ctx := context.WithValue(r.Context(), "userdata", val)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
			strId := strconv.FormatFloat(userId, 'f', -1, 64)
			user_name, ok := claims["name"].(string)
			if !ok {
				val.UserData["Authorized"] = "0"
				ctx := context.WithValue(r.Context(), "userdata", val)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
			val.UserData["Id"] = strId
			val.UserData["Username"] = user_name
			val.UserData["Authorized"] = "1"
			ctx := context.WithValue(r.Context(), "userdata", val)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		val.UserData["Authorized"] = "0"
		ctx := context.WithValue(r.Context(), "userdata", val)
		next.ServeHTTP(w, r.WithContext(ctx))
	})

}
