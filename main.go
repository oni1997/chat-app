package main

import (
	"chat-app/internal/handler"
	"github.com/labstack/echo/v4"
)

func main() {
	handler.LoadMessages()

	e := echo.New()
	e.GET("/", handler.ChatPage)
	e.POST("/login", handler.Login)
	e.POST("/logout", handler.Logout)
	e.POST("/send", handler.SendMessage)
	e.GET("/messages", handler.GetMessages)

	e.Logger.Fatal(e.Start(":8080"))
}
