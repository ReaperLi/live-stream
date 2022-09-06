package api

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"strconv"
	"time"
)

type Server struct {
	Chatroom *ChatRoom
	Router   *gin.Engine
	Upgrader websocket.Upgrader
}

func NewServer() *Server {
	server := &Server{
		Chatroom: NewChatRoom(),
		Upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			// 解决跨域问题
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
	router := gin.Default()
	router.Static("./css", "./resources/css")
	router.Static("./js", "./resources/js")
	router.Static("./img", "./resources/img")
	router.LoadHTMLGlob("./resources/views/*")
	router.GET("/room/:id", server.room)
	//router.POST("/room/sendMsg", server.Chatroom.SendMsg)
	router.GET("/chatroom/:id", server.wsHandler)

	server.Router = router
	return server
}

func (server *Server) Start(address string) error {
	return server.Router.Run(address)
}

type WsResponse struct {
	DataType uint32 `json:"data_type"`
	Message  string `json:"message"`
	Status   uint32 `json:"status"`
}

func (server *Server) wsHandler(c *gin.Context) {
	var (
		wsConn *websocket.Conn
		err    error
		conn   *Connection
		recv   []byte
	)
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
		resp, _ := Package(1, "ping")
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

	for {

		log.Println("for begins")
		recv, err = conn.ReadMessage()
		if err != nil {
			goto ERR
		}
		log.Println(string(recv))
		rec, _ := Package(2, string(recv))

		for i, _ := range server.Chatroom.GetChat(uint32(sessionID)).GetConns() {
			err = i.WriteMessage(rec)
			if err != nil {
				goto ERR
			}
		}
	}

	//TODO:解绑sessionID并注销连接
ERR:
	log.Println("-------------ERR------------------")
	conn.Close()
}

func Package(dataType uint32, msg string) ([]byte, error) {
	var res = WsResponse{
		DataType: dataType, //心跳
		Message:  msg,
		Status:   0, //0 success
	}
	return json.Marshal(&res)
}
