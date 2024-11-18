package handler

import (
	"chat-app/models"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"github.com/labstack/echo/v4"
	"net/http"
	"os"
	"sync"
	"time"
)

var (
	messages []models.Message
	users    = make(map[string]models.User)
	sessions = make(map[string]string) // cookie -> userID
	mu       sync.Mutex
	dataFile = "chat.json"
)

func generateID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
func LoadMessages() {
	mu.Lock()
	defer mu.Unlock()

	if _, err := os.Stat(dataFile); os.IsNotExist(err) {
		return
	}

	data, err := os.ReadFile(dataFile)
	if err != nil {
		panic(err)
	}

	if err := json.Unmarshal(data, &messages); err != nil {
		panic(err)
	}
}

func loadMessagesFromFile() error {
	mu.Lock()
	defer mu.Unlock()

	if _, err := os.Stat(dataFile); os.IsNotExist(err) {
		messages = []models.Message{}
		return nil
	}

	data, err := os.ReadFile(dataFile)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &messages)
}

func SaveMessages() {
	mu.Lock()
	defer mu.Unlock()

	data, err := json.MarshalIndent(messages, "", "  ")
	if err != nil {
		panic(err)
	}

	if err := os.WriteFile(dataFile, data, 0644); err != nil {
		panic(err)
	}
}

func saveMessageToFile(msg models.Message) error {
	mu.Lock()
	messages = append(messages, msg)
	mu.Unlock()

	mu.Lock()
	defer mu.Unlock()

	data, err := json.MarshalIndent(messages, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(dataFile, data, 0644)
}

func ChatPage(c echo.Context) error {
	cookie, err := c.Cookie("session")
	if err != nil || cookie.Value == "" {
		// Show login page if not logged in
		return c.HTML(http.StatusOK, `
            <!DOCTYPE html>
            <html>
            <head>
                <title>Chat App - Login</title>
                <script src="https://unpkg.com/htmx.org@1.9.2"></script>
                <style>
                    body {
                        font-family: Arial, sans-serif;
                        background-color: #e5ddd5;
                        margin: 0;
                        padding: 0;
                        display: flex;
                        flex-direction: column;
                        align-items: center;
                        justify-content: center;
                        height: 100vh;
                    }
                    .login-form {
                        background-color: white;
                        padding: 20px;
                        border-radius: 8px;
                        box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
                    }
                    input[type="text"] {
                        padding: 10px;
                        border: 1px solid #ddd;
                        border-radius: 20px;
                        font-size: 16px;
                        margin-bottom: 10px;
                        width: 200px;
                    }
                    button {
                        padding: 10px 20px;
                        border: none;
                        background-color: #25D366;
                        color: white;
                        border-radius: 20px;
                        cursor: pointer;
                        font-size: 16px;
                        width: 100%;
                    }
                </style>
            </head>
            <body>
                <div class="login-form">
                    <h2>Enter Chat</h2>
                    <form action="/login" method="POST">
                        <input type="text" name="name" placeholder="Enter your name" required>
                        <button type="submit">Join Chat</button>
                    </form>
                </div>
            </body>
            </html>
        `)
	}

	mu.Lock()
	userID, exists := sessions[cookie.Value]
	user := users[userID]
	mu.Unlock()

	if !exists {
		c.SetCookie(&http.Cookie{
			Name:    "session",
			Value:   "",
			Expires: time.Now().Add(-1 * time.Hour),
		})
		return c.Redirect(http.StatusSeeOther, "/")
	}

	return c.HTML(http.StatusOK, `
        <!DOCTYPE html>
        <html>
        <head>
            <title>Chat App</title>
            <script src="https://unpkg.com/htmx.org@1.9.2"></script>
            <style>
                body {
                    font-family: Arial, sans-serif;
                    background-color: #e5ddd5;
                    margin: 0;
                    padding: 0;
                    display: flex;
                    flex-direction: column;
                    align-items: center;
                    justify-content: flex-start;
                    height: 100vh;
                }
                h1 {
                    color: #075E54;
                }
                #chatbox {
                    width: 90%;
                    max-width: 600px;
                    height: 70vh;
                    border: none;
                    border-radius: 8px;
                    background-color: #fff;
                    overflow-y: scroll;
                    padding: 10px;
                    margin-bottom: 10px;
                    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
                }
                .chat-bubble {
                    max-width: 70%;
                    padding: 10px;
                    margin: 5px;
                    border-radius: 10px;
                    font-size: 14px;
                    word-wrap: break-word;
                    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.2);
                }
                .chat-bubble.self {
                    background-color: #DCF8C6;
                    align-self: flex-end;
                    margin-left: auto;
                }
                .chat-bubble.other {
                    background-color: #ECE5DD;
                    align-self: flex-start;
                }
                form {
                    width: 90%;
                    max-width: 600px;
                    display: flex;
                    gap: 10px;
                }
                input[type="text"] {
                    flex: 1;
                    padding: 10px;
                    border: 1px solid #ddd;
                    border-radius: 20px;
                    font-size: 16px;
                }
                button {
                    padding: 10px 20px;
                    border: none;
                    background-color: #25D366;
                    color: white;
                    border-radius: 20px;
                    cursor: pointer;
                    font-size: 16px;
                }
                button:hover {
                    background-color: #1EBE56;
                }
                .user-info {
                    margin-bottom: 10px;
                    color: #075E54;
                }
            </style>
        </head>
        <body>
            <div class="user-info">
                Logged in as: `+user.Name+`
                <form action="/logout" method="POST" style="display: inline;">
                    <button type="submit">Logout</button>
                </form>
            </div>
            <div id="chatbox" hx-get="/messages" hx-trigger="load, every 2s" hx-swap="innerHTML" style="display: flex; flex-direction: column;"></div>
            <form hx-post="/send" hx-target="#chatbox">
                <input type="text" name="message" placeholder="Type a message..." required>
                <button type="submit">Send</button>
            </form>
            <script>
                document.addEventListener('htmx:afterSwap', function(evt) {
                    if (evt.detail.target.id === "chatbox") {
                        evt.detail.target.scrollTop = evt.detail.target.scrollHeight;
                    }
                });
            </script>
        </body>
        </html>
    `)
}

func Login(c echo.Context) error {
	name := c.FormValue("name")
	if name == "" {
		return c.Redirect(http.StatusSeeOther, "/")
	}

	userID := generateID()
	user := models.User{
		ID:        userID,
		Name:      name,
		CreatedAt: time.Now(),
	}

	sessionID := generateID()

	mu.Lock()
	users[userID] = user
	sessions[sessionID] = userID
	mu.Unlock()

	c.SetCookie(&http.Cookie{
		Name:    "session",
		Value:   sessionID,
		Expires: time.Now().Add(24 * time.Hour),
	})

	return c.Redirect(http.StatusSeeOther, "/")
}

func Logout(c echo.Context) error {
	cookie, err := c.Cookie("session")
	if err == nil {
		mu.Lock()
		delete(sessions, cookie.Value)
		mu.Unlock()
	}

	c.SetCookie(&http.Cookie{
		Name:    "session",
		Value:   "",
		Expires: time.Now().Add(-1 * time.Hour),
	})

	return c.Redirect(http.StatusSeeOther, "/")
}

//func SendMessage(c echo.Context) error {
//	cookie, err := c.Cookie("session")
//	if err != nil {
//		return c.Redirect(http.StatusSeeOther, "/")
//	}
//
//	mu.Lock()
//	userID, exists := sessions[cookie.Value]
//	user := users[userID]
//	mu.Unlock()
//
//	if !exists {
//		return c.Redirect(http.StatusSeeOther, "/")
//	}
//
//	content := c.FormValue("message")
//	if content == "" {
//		return GetMessages(c)
//	}
//
//	newMessage := models.Message{
//		ID:        generateID(),
//		UserID:    userID,
//		UserName:  user.Name,
//		Content:   content,
//		CreatedAt: time.Now(),
//	}
//
//	mu.Lock()
//	messages = append(messages, newMessage)
//	mu.Unlock()
//
//	SaveMessages()
//
//	return GetMessages(c)
//}

func SendMessage(c echo.Context) error {
	cookie, err := c.Cookie("session")
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/")
	}

	mu.Lock()
	userID, exists := sessions[cookie.Value]
	user := users[userID]
	mu.Unlock()

	if !exists {
		return c.Redirect(http.StatusSeeOther, "/")
	}

	content := c.FormValue("message")
	if content == "" {
		return GetMessages(c)
	}

	newMessage := models.Message{
		ID:        generateID(),
		UserID:    userID,
		UserName:  user.Name,
		Content:   content,
		CreatedAt: time.Now(),
	}

	// Save the new message
	if err := saveMessageToFile(newMessage); err != nil {
		// Handle error appropriately
		return c.String(http.StatusInternalServerError, "Error saving message")
	}

	return GetMessages(c)
}

//func GetMessages(c echo.Context) error {
//	cookie, err := c.Cookie("session")
//	if err != nil {
//		return c.Redirect(http.StatusSeeOther, "/")
//	}
//
//	mu.Lock()
//	userID := sessions[cookie.Value]
//	messagesCopy := make([]models.Message, len(messages))
//	copy(messagesCopy, messages)
//	mu.Unlock()
//
//	var chatContent string
//	for _, msg := range messagesCopy {
//		class := "other"
//		if msg.UserID == userID {
//			class = "self"
//		}
//		chatContent += `<div class="chat-bubble ` + class + `"><strong>` + msg.UserName + `:</strong> ` + msg.Content + `</div>`
//	}
//
//	return c.HTML(http.StatusOK, chatContent)
//}

func GetMessages(c echo.Context) error {
	cookie, err := c.Cookie("session")
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/")
	}

	// Load messages from file
	if err := loadMessagesFromFile(); err != nil {
		return c.String(http.StatusInternalServerError, "Error loading messages")
	}

	mu.Lock()
	userID := sessions[cookie.Value]
	messagesCopy := make([]models.Message, len(messages))
	copy(messagesCopy, messages)
	mu.Unlock()

	var chatContent string
	for _, msg := range messagesCopy {
		class := "other"
		if msg.UserID == userID {
			class = "self"
		}
		chatContent += `<div class="chat-bubble ` + class + `"><strong>` + msg.UserName + `:</strong> ` + msg.Content + `</div>`
	}

	return c.HTML(http.StatusOK, chatContent)
}
