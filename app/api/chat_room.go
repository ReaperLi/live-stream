package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"sync"
)

type ChatRoom struct {
	chats map[uint32]*ConnManager
	mutex sync.Mutex
}

func NewChatRoom() *ChatRoom {
	return &ChatRoom{
		chats: make(map[uint32]*ConnManager),
	}
}

func (chatRoom *ChatRoom) GetChat(roomID uint32) *ConnManager {
	chatRoom.mutex.Lock()
	defer chatRoom.mutex.Unlock()
	if chat, ok := chatRoom.chats[roomID]; ok {
		return chat
	}
	chatRoom.chats[roomID] = NewConnManager()
	return chatRoom.chats[roomID]
}

//func (chatRoom *ChatRoom) SendMsg(c *gin.Context) {
//	var sendMsg sendMsgRequest
//	if err := c.ShouldBindJSON(&sendMsg); err != nil {
//		c.JSON(http.StatusBadRequest, gin.H{
//			"error": err.Error(),
//		})
//		return
//	}
//
//	roomID, err := strconv.Atoi(sendMsg.RoomID)
//	if err != nil {
//		c.JSON(http.StatusInternalServerError, gin.H{
//			"error": err.Error(),
//		})
//		return
//	}
//
//	connManager := chatRoom.GetChat(uint32(roomID))
//	for _, conn := range connManager.connections {
//		err = conn.write(websocket.TextMessage, []byte(sendMsg.Message))
//		if err != nil {
//			c.JSON(http.StatusInternalServerError, gin.H{
//				"error": err.Error(),
//			})
//			return
//		}
//	}
//
//	c.JSON(http.StatusOK, gin.H{
//		"ok": sendMsg.RoomID,
//	})
//}

func (server *Server) room(c *gin.Context) {
	payload, _ := c.Get(authorizationPayloadKey)
	user := payload.(*UserResponse)
	var username string
	if user != nil {
		username = user.Username
	} else {
		username = "匿名用户"
	}

	c.HTML(http.StatusOK, "chatroom.html", gin.H{
		"room_id":  c.Param("id"),
		"username": username,
	})
}

//func (server *Server) chatRoom(c *gin.Context) {
//	var conn, err = server.Upgrader.Upgrade(c.Writer, c.Request, nil)
//	if err != nil {
//		c.JSON(http.StatusInternalServerError, err.Error())
//		return
//	}
//
//	for {
//		t, msg, err := conn.ReadMessage()
//		if err != nil {
//			c.JSON(http.StatusInternalServerError, err.Error())
//			break
//		}
//		conn.WriteMessage(t, msg)
//	}
//	roomSerial := c.Param("id")
//	roomID, err := strconv.Atoi(roomSerial)
//
//	connManager := server.Chatroom.GetChat(uint32(roomID))
//	ConnID++
//	newConn := NewConnection(conn, ConnID)
//	connManager.Register(newConn)
//	//defer connManager.Remove(ConnID)
//
//	for k := range connManager.conns {
//		log.Println("-----------------------", k)
//	}
//
//}
