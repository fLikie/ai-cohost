package stream

import (
	"cohost/internal/config"
	"cohost/internal/response"
	storage "cohost/internal/storage"
	gui "cohost/internal/ui"
	"fmt"
	"github.com/gempir/go-twitch-irc/v4"
	"log"
	"os"
)

var twitchBotStarted = false

// 🎥 Подключение к Twitch
func StartTwitchBot() {
	if twitchBotStarted {
		return
	}
	twitchBotStarted = true

	twitchToken := os.Getenv("TWITCH_TOKEN")

	client := twitch.NewClient("cohost_bot", twitchToken)

	client.OnPrivateMessage(func(message twitch.PrivateMessage) {
		username := message.User.DisplayName
		text := message.Message
		msg := fmt.Sprintf("💬 Twitch %s: %s", message.User.DisplayName, message.Message)
		fmt.Println(msg)
		gui.SetUsersText(msg)
		gui.AppendToChatHistory(msg)

		storage.UpdateUser(username, text)

		// Проверяем, представился ли он
		storage.DetectName(username, text)

		response := response.GenerateAIResponse(text, username)
		gui.SetChatText(fmt.Sprintf("🤖 AI: %s", response))
	})

	log.Println("Присоединемся к каналу " + config.Settings.TwitchChannel)
	client.Join(config.Settings.TwitchChannel)
	err := client.Connect()
	if err != nil {
		log.Fatal("Ошибка Twitch:", err)
	}
}
