package api

import (
	"encoding/json"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
)

const (
	authorizationHeaderKey  = "authorization"
	authorizationTypeBearer = "bearer"
	authorizationPayloadKey = "authorization_payload"
	authorizationSessionKey = "session_key"
)

func authMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		payload, _ := isLoggedIn(ctx)
		ctx.Set(authorizationPayloadKey, payload)
		ctx.Next()
	}
}

func MustLoginMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		u := ctx.MustGet(authorizationPayloadKey)
		if u == nil {
			ctx.Redirect(http.StatusMovedPermanently, "/login")
		}
		ctx.Next()
	}
}

func isLoggedIn(ctx *gin.Context) (*UserResponse, bool) {
	sessionValue, exists := ctx.Get(sessions.DefaultKey)
	if !exists {
		return nil, false
	}
	if sessionValue == nil {
		return nil, false
	}
	session := sessionValue.(sessions.Session)
	bytes := session.Get("user")
	if bytes == nil {
		return nil, false
	}
	var payload UserResponse
	err := json.Unmarshal(bytes.([]byte), &payload)
	if err != nil {
		return nil, false
	}
	if payload.ID < 1 {
		return nil, false
	}
	return &payload, true
}
