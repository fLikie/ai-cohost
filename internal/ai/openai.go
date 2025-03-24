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

func GenerateOpenAiResponse() string {
	client := resty.New()

	body := map[string]interface{}{
		"model":      "gpt-4o-mini",
		"max_tokens": 50,
		"messages": append([]map[string]string{
			{"role": "system", "content": fmt.Sprintf("Ты AI-соведущий стрима по игре %s. Помни контекст чата и поддерживай разговор. Если ты видишь текст от юзера local, значит ты видишь текст с экрана, Отвечай коротко и ёмко", gui.SelectedGame)},
		}, storage.ChatHistory...),
	}

	openAIKey := os.Getenv("OPENAI_KEY")

	resp, err := client.R().
		SetHeader("Authorization", "Bearer "+openAIKey).
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		Post("https://api.openai.com/v1/chat/completions")

	if err != nil {
		log.Println("❌ Ошибка запроса к OpenAI:", err)
		return "Ошибка AI: нет ответа"
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		log.Println("❌ Ошибка парсинга JSON:", err)
		return "Ошибка AI: JSON"
	}

	// ✅ Проверяем, есть ли поле "choices"
	choices, ok := result["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		log.Println("❌ OpenAI вернул неожиданный ответ:", result)
		return "Ошибка AI: пустой ответ"
	}

	// ✅ Проверяем, есть ли поле "message"
	messageData, ok := choices[0].(map[string]interface{})["message"].(map[string]interface{})
	if !ok {
		log.Println("❌ Ошибка: отсутствует поле 'message' в JSON:", choices)
		return "Ошибка AI: нет сообщения"
	}

	// ✅ Извлекаем ответ AI
	response, ok := messageData["content"].(string)
	if !ok {
		log.Println("❌ Ошибка: 'content' не является строкой:", messageData)
		return "Ошибка AI: пустой контент"
	}

	return response
}
