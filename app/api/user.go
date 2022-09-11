package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	db "github.com/reaper/live-stream/db/sqlc"
	"github.com/reaper/live-stream/util"
	"net/http"
	"time"
)

func (server *Server) signup(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "signup.html", gin.H{
		"title": "signup",
	})
}

func (server *Server) login(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "login.html", gin.H{
		"title": "login",
	})
}

type createUserRequest struct {
	Username        string `json:"username" binding:"required,alphanum,min=6,max=18"`
	Password        string `json:"password" binding:"required,min=6,max=18"`
	PasswordConfirm string `json:"password_confirm" binding:"required"`
}

func (server *Server) createAccount(ctx *gin.Context) {
	var req createUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	if ok := util.IsPasswordConfirmed(req.Password, req.PasswordConfirm); !ok {
		err := errors.New("PasswordConfirm is not correct")
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	HashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	params := db.CreateUserParams{
		Username:       req.Username,
		HashedPassword: HashedPassword,
	}

	user, err := server.store.CreateUser(ctx, params)
	if err != nil {
		//如果是数据库出错
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "unique_violation":
				ctx.JSON(http.StatusForbidden, errorResponse(err))
				return
			}
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
	}
	resp := newUserResponse(user)
	ctx.JSON(http.StatusOK, resp)
}

type UserResponse struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
}

func newUserResponse(user db.User) UserResponse {
	return UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		CreatedAt: user.CreatedAt,
	}
}

type loginUserRequest struct {
	Username string `json:"username" binding:"required,alphanum"`
	Password string `json:"password" binding:"required,min=6"`
}

type loginUserResponse struct {
	AccessToken string
	User        UserResponse
}

func (server *Server) loginUser(ctx *gin.Context) {
	var req loginUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	user, err := server.store.GetUser(ctx, req.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	err = util.CheckPassword(req.Password, user.HashedPassword)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	//create session for user
	// 初始化session对象

	session := sessions.Default(ctx)
	// 通过session.Get读取session值
	// session是键值对格式数据，因此需要通过key查询数据

	//session.Get(")
	jsonUser, err := json.Marshal(newUserResponse(user))
	if err != nil {
		errors.New("json marshal error")
	}
	session.Set("user", jsonUser)
	//session.Delete()
	session.Save()
	// 删除整个session
	// session.Clear()

	//create access token for user
	accessToken, err := server.tokenMaker.CreateToken(
		user.Username,
		server.config.AccessTokenDuration,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	rsp := loginUserResponse{
		AccessToken: accessToken,
		User:        newUserResponse(user),
	}
	ctx.JSON(http.StatusOK, rsp)
}

func (server *Server) user(ctx *gin.Context) {
	payload := ctx.MustGet(authorizationPayloadKey).(*UserResponse)
	ctx.JSON(http.StatusOK, payload)
}
