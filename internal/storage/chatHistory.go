package storage

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

// 🔥 Файл, в котором хранится история
const historyFile = "chat_history.json" // Максимум 10 последних сообщений

var ChatHistory []map[string]string // Хранит историю сообщений
const MaxHistory = 100              // Максимум 100 последних сообщений

// 📜 Загружаем историю сообщений при запуске
func LoadChatHistory() {
	if _, err := os.Stat(historyFile); os.IsNotExist(err) {
		log.Println("📂 Файл истории не найден, создаём новый...")
		return
	}

	data, err := ioutil.ReadFile(historyFile)
	if err != nil {
		log.Println("❌ Ошибка чтения истории:", err)
		return
	}

	err = json.Unmarshal(data, &ChatHistory)
	if err != nil {
		log.Println("❌ Ошибка парсинга JSON:", err)
		ChatHistory = []map[string]string{} // Если файл сломан, сбрасываем историю
	}
	log.Println("✅ История сообщений загружена! Количество записей:", len(ChatHistory))
}

// 💾 Сохраняем историю сообщений в файл
func SaveChatHistory() {
	data, err := json.MarshalIndent(ChatHistory, "", "  ")
	if err != nil {
		log.Println("❌ Ошибка сохранения истории:", err)
		return
	}

	err = ioutil.WriteFile(historyFile, data, 0644)
	if err != nil {
		log.Println("❌ Ошибка записи в файл:", err)
	}
}
