package handler

import (
	"chat-app/internal/handler"
	"github.com/labstack/echo/v4"
	"net/http"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	e := echo.New()

	e.GET("/", handler.ChatPage)
	e.POST("/login", handler.Login)
	e.POST("/logout", handler.Logout)
	e.POST("/send", handler.SendMessage)
	e.GET("/messages", handler.GetMessages)

	handler.LoadMessages()

	e.ServeHTTP(w, r)
}
