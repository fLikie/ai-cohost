package response

import (
	"bufio"
	"cohost/internal/ai"
	"cohost/internal/audio"
	"cohost/internal/config"
	"cohost/internal/storage"
	gui "cohost/internal/ui"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strings"
)

func ListenVoiceCommands() {
	log.Println("🚀 Запуск voice_listener.py")
	cmd := exec.Command("python", "-u", "internal/audio/voice_listener.py", config.Settings.WakeWord)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Println("❌ Ошибка получения stdout:", err)
		return
	}
	stderr, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		log.Println("❌ Ошибка запуска voice_listener.py:", err)
		return
	}

	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			log.Println("🐍 stderr:", scanner.Text())
		}
	}()

	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "{") {
				var result map[string]string
				if err := json.Unmarshal([]byte(line), &result); err == nil {
					if command := result["command"]; command != "" {
						log.Println("🎤 Голосовая команда:", command)
						GenerateAIResponse("Вы", command) // или что-то своё
					}
				}
			} else {
				log.Println("🔈", line)
			}
		}
	}()
}

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
		response = ai.GenerateOpenAiResponse()
	} else {
		response = ai.GetDeepSeekResponse()
	}

	// ✅ Обновляем UI
	gui.SetChatText("🤖 AI: " + response)
	gui.AppendToChatHistory("🤖 AI: " + response)

	// ✅ Озвучиваем AI-ответ
	if config.Settings.SelectedTTS == "ElevenLabs" {
		go audio.GenerateVoice(response)
	} else {
		log.Println("нужно добавить второй обработчик")
		go audio.GenerateSileroVoice(response)
	}

	storage.ChatHistory = append(storage.ChatHistory, map[string]string{"role": "assistant", "content": response})
	if len(storage.ChatHistory) > storage.MaxHistory {
		storage.ChatHistory = storage.ChatHistory[1:]
	}

	storage.SaveChatHistory()

	return response
}
