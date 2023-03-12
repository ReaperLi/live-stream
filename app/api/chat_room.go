package api

import (
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
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
	//创建房间同时创建消费者，房间id与topic相等
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": "PLAINTEXT://192.168.1.5:19092",
		"security.protocol": "PLAINTEXT",
		"sasl.mechanisms":   "PLAIN",
		"acks":              "all",
		"group.id":          "chat01",
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		fmt.Printf("Failed to create consumer: %s", err)
	}
	go chatRoom.ConsumeMessages(c, roomID)
	return chatRoom.chats[roomID]
}

func (chatRoom *ChatRoom) ConsumeMessages(c *kafka.Consumer, roomID uint32) {
	topic := strconv.Itoa(int(roomID))
	err := c.SubscribeTopics([]string{topic}, nil)
	if err != nil {
		fmt.Printf("Failed to subscribe topic: %s", err)
	}
	//defer c.Close()
	for {
		msg, err := c.ReadMessage(100 * time.Millisecond)
		if err != nil {
			//fmt.Printf("consuming messages failed :%s", err)
			continue
		}
		//consuming logic
		for i := range chatRoom.GetChat(roomID).GetConns() {
			err = i.WriteMessage(msg.Value)
			if err != nil {
				log.Println("-------------消费时ERR------------------")
				i.Close()
			}
		}
		fmt.Printf("Consumed event from topic %s: partition %d : key = %-10s value = %s\n",
			*msg.TopicPartition.Topic, msg.TopicPartition.Partition, string(msg.Key), string(msg.Value))
	}

}

//	func (chatRoom *ChatRoom) SendMsg(c *gin.Context) {
//		var sendMsg sendMsgRequest
//		if err := c.ShouldBindJSON(&sendMsg); err != nil {
//			c.JSON(http.StatusBadRequest, gin.H{
//				"error": err.Error(),
//			})
//			return
//		}
//
//		roomID, err := strconv.Atoi(sendMsg.RoomID)
//		if err != nil {
//			c.JSON(http.StatusInternalServerError, gin.H{
//				"error": err.Error(),
//			})
//			return
//		}
//
//		connManager := chatRoom.GetChat(uint32(roomID))
//		for _, conn := range connManager.connections {
//			err = conn.write(websocket.TextMessage, []byte(sendMsg.Message))
//			if err != nil {
//				c.JSON(http.StatusInternalServerError, gin.H{
//					"error": err.Error(),
//				})
//				return
//			}
//		}
//
//		c.JSON(http.StatusOK, gin.H{
//			"ok": sendMsg.RoomID,
//		})
//	}
var count int

func (server *Server) room(c *gin.Context) {
	payload, _ := c.Get(authorizationPayloadKey)
	user := payload.(*UserResponse)
	var username string
	if user != nil {
		username = user.Username
	} else {
		username = "匿名用户"
	}
	count++
	log.Println("---------请求成功次数", count)
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
