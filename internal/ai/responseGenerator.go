package ai

import (
	"cohost/internal/audio"
	"cohost/internal/config"
	"cohost/internal/storage"
	gui "cohost/internal/ui"
	"fmt"
	"log"
	"strings"
)

// 🎙 AI-ответ через GPT-4
func GenerateAIResponse(username, message string) string {
	// ✅ Проверяем, чтобы AI не отвечал сам себе
	storage.ChatHistory = append(storage.ChatHistory, map[string]string{"role": "user", "content": fmt.Sprintf("%s: %s", username, message)})
	if len(storage.ChatHistory) > storage.MaxHistory {
		storage.ChatHistory = storage.ChatHistory[1:]
	}
	if strings.HasPrefix(message, "🤖 AI:") {
		return ""
	}

	var response string
	if config.Settings.SelectedAiModel == "OpenAI" {
		response = GenerateOpenAiResponse()
	} else {
		response = GetDeepSeekResponse()
	}

	// ✅ Обновляем UI
	gui.SetChatText("🤖 AI: " + response)

	// ✅ Озвучиваем AI-ответ
	if config.Settings.SelectedTTS == "ElevenLabs" {
		go audio.GenerateVoice(response)
	} else {
		log.Println("нужно добавить второй обработчик")
		//go generateVoiceYandex(response)
	}

	storage.ChatHistory = append(storage.ChatHistory, map[string]string{"role": "assistant", "content": response})
	if len(storage.ChatHistory) > storage.MaxHistory {
		storage.ChatHistory = storage.ChatHistory[1:]
	}

	storage.SaveChatHistory()

	return response
}
