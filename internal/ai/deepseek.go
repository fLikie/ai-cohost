package ai

import (
	"cohost/internal/storage"
	gui "cohost/internal/ui"
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"log"
	"os"
)

func GetDeepSeekResponse() string {
	client := resty.New()
	body := map[string]interface{}{
		"model": "deepseek-chat",
		"messages": append([]map[string]string{
			{"role": "system", "content": fmt.Sprintf("Ты AI-соведущий стрима по игре %s.", gui.SelectedGame)},
		}, storage.ChatHistory...),
		"stream": false,
	}

	deepseekKey := os.Getenv("DEEPSEEK_KEY")

	resp, err := client.R().
		SetHeader("Authorization", "Bearer "+deepseekKey).
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		Post("https://api.deepseek.com/v1/chat/completions")

	if err != nil {
		log.Println("❌ Ошибка DeepSeek:", err)
		return "Ошибка AI: нет ответа"
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		log.Println("❌ Ошибка парсинга JSON:", err)
		return "Ошибка AI: JSON"
	}

	choices, ok := result["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		log.Println("❌ AI пустой ответ:")
		return "Ошибка AI: пустой ответ"
	}

	messageData, ok := choices[0].(map[string]interface{})["message"].(map[string]interface{})
	if !ok {
		log.Println("❌ Ошибка: отсутствует поле 'message' в JSON:", choices)
		return "Ошибка AI: нет сообщения"
	}

	// ✅ Проверяем, что ответ не "Ошибка AI: пустой ответ"
	response, ok := messageData["content"].(string)
	if !ok || response == "Ошибка AI: пустой ответ" {
		log.Println("❌ AI сгенерировал ошибку вместо ответа")
		return "Ошибка AI: нет полезного ответа"
	}

	log.Println("✅ AI ответ:", response)
	return response
}
