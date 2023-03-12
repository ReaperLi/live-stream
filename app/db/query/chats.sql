-- name: CreateChat :one
INSERT INTO chats (
    user_id,anonym, message,session_id,room_id
) VALUES (
             $1, $2 ,$3,$4,$5
         )
    RETURNING *;
