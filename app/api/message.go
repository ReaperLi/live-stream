package api

import (
	"context"
	db "github.com/reaper/live-stream/db/sqlc"
	"log"
)

func (server *Server) MessageSave(ctx context.Context, userid, roomid int64, anonym, msg string) {
	arg := db.CreateChatParams{
		UserID:    userid,
		Anonym:    anonym,
		Message:   msg,
		SessionID: "",
		RoomID:    roomid,
	}
	_, err := server.store.CreateChat(ctx, arg)
	if err != nil {
		log.Println(err)
	}
}
