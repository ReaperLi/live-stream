package api

import (
	"encoding/json"
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	db "github.com/reaper/live-stream/db/sqlc"
	"github.com/reaper/live-stream/token"
	"github.com/reaper/live-stream/util"
	"log"
	"net/http"
	"strconv"
	"time"
)

var UserNo int = 0

type Server struct {
	config     util.Config
	store      db.Store
	Chatroom   *ChatRoom
	Router     *gin.Engine
	tokenMaker token.Maker
	Upgrader   websocket.Upgrader
}

func NewServer(config util.Config, store db.Store) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)

	if err != nil {
		return nil, fmt.Errorf("cannot create token maker:%w", err)
	}
	server := &Server{
		config:     config,
		store:      store,
		Chatroom:   NewChatRoom(),
		tokenMaker: tokenMaker,
		Upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			// 解决跨域问题
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
	//if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
	//	v.RegisterValidation("password_confirm", ValidPasswordConfirm)
	//}

	server.setupRouter()
	return server, nil
}

func (server *Server) setupRouter() {
	router := gin.Default()
	store := cookie.NewStore([]byte("111111"))
	store.Options(sessions.Options{
		Path:     "/",
		Domain:   "192.168.1.5",
		MaxAge:   86400,
		Secure:   false,
		HttpOnly: false,
		SameSite: 0,
	})
	//	sessions.Sessions("session001", store)
	router.Use(sessions.Sessions("session001", store))
	authRoutes := router.Group("/").Use(authMiddleware())
	MustLogin := authRoutes.Use(MustLoginMiddleware())

	router.Static("./css", "./resources/css")
	router.Static("./js", "./resources/js")
	router.Static("./img", "./resources/img")
	router.Static("./fonts", "./resources/fonts")
	router.LoadHTMLGlob("./resources/views/*")
	authRoutes.GET("/room/:id", server.room)
	authRoutes.GET("/chatroom/:id", server.wsHandler)
	router.GET("/login", server.login)
	router.GET("/signup", server.signup)
	router.POST("/create_account", server.createAccount)
	authRoutes.POST("/login", server.loginUser)
	MustLogin.GET("/user", server.user)

	server.Router = router
}

func (server *Server) Start(address string) error {
	return server.Router.Run(address)
}

type WsResponse struct {
	FromUser string `json:"from_user"`
	DataType uint32 `json:"data_type"`
	Message  string `json:"message"`
	Status   uint32 `json:"status"`
}

var countWS int

func (server *Server) wsHandler(c *gin.Context) {
	var (
		wsConn *websocket.Conn
		err    error
		conn   *Connection
		recv   []byte
	)
	UserNo++
	username := "user" + strconv.Itoa(UserNo)
	countWS++
	log.Println("---------WS请求成功次数", countWS)
	//upgrade connection to websocket
	wsConn, err = server.Upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	//将该连接绑定到对应房间ID
	IDSerial := c.Param("id")
	sessionID, err := strconv.Atoi(IDSerial)

	//get/create connManager
	connManager := server.Chatroom.GetChat(uint32(sessionID))
	//create new ws conn
	conn, err = NeWConnection(wsConn, connManager)
	//bind the conn to the connManager
	connManager.Register(conn)
	//set closeHandler
	//conn.wsConn.SetCloseHandler(conn.closeHandler)
	if err != nil {
		log.Println(err.Error())
		conn.Close()
		return
	}

	//心跳设置
	go func() {
		var err error
		resp, _ := Package(1, "matrix", "ping")
		for {
			err = conn.WriteMessage(resp)
			if err != nil {
				log.Println(err.Error())
				conn.Close()
				return
			}
			time.Sleep(time.Second * 5)
		}
	}()

	//kafka配置
	producer, err := sarama.NewSyncProducer([]string{"192.168.1.5:19092"}, nil)

	//只允许登录用户发送消息
	payload, _ := c.Get(authorizationPayloadKey)
	user := payload.(*UserResponse)
	if user == nil {
		for {
			//接收客户端推送来的消息
			recv, err = conn.ReadMessage()
			if err != nil {
				goto ERR
			}
			//-----------生产者
			rec, _ := Package(2, username, string(recv))
			msg := &sarama.ProducerMessage{Topic: IDSerial, Value: sarama.StringEncoder(rec)}
			_, _, err := producer.SendMessage(msg)
			log.Println("----------------produced message")
			if err != nil {
				goto ERR
			}

		}
	}
	//for {
	//	//接收客户端推送来的消息
	//	recv, err = conn.ReadMessage()
	//	if err != nil {
	//		goto ERR
	//	}
	//	go server.MessageSave(c, user.ID, int64(sessionID), "", string(recv))
	//	rec, _ := Package(2, user.Username, string(recv))
	//	//将接收到的消息推送给当前房间所有客户端
	//	for i := range server.Chatroom.GetChat(uint32(sessionID)).GetConns() {
	//		err = i.WriteMessage(rec)
	//		if err != nil {
	//			goto ERR
	//		}
	//	}
	//}

	//TODO:解绑sessionID并注销连接
ERR:
	log.Println("-------------ERR------------------")
	conn.Close()
}

func Package(dataType uint32, fromUser, msg string) ([]byte, error) {
	var res = WsResponse{
		FromUser: fromUser,
		DataType: dataType, //心跳
		Message:  msg,
		Status:   0, //0 success
	}
	return json.Marshal(&res)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
