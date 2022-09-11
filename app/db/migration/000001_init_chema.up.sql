CREATE TABLE "users" (
                         "id" bigserial PRIMARY KEY,
                         "username" varchar UNIQUE NOT NULL,
                         "hashed_password" varchar NOT NULL,
                         "created_at" timestamp NOT NULL DEFAULT (now())
);

CREATE TABLE "chats" (
                         "id" bigserial PRIMARY KEY,
                         "user_id" bigserial NOT NULL,
                         "message" varchar NOT NULL,
                         "created_at" timestamp NOT NULL DEFAULT (now())
);

CREATE TABLE "rooms" (
                         "id" bigserial PRIMARY KEY,
                         "user_id" bigserial NOT NULL,
                         "room_name" varchar UNIQUE NOT NULL,
                         "created_at" timestamp NOT NULL DEFAULT (now())
);

ALTER TABLE "chats" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id");

ALTER TABLE "rooms" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id");
