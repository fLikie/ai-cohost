package storage

import (
	"encoding/json"
	"log"
	"os"
	"strings"
	"time"
)

// Структура пользователя
type User struct {
	Name         string `json:"name"`
	FirstSeen    string `json:"first_seen"`
	MessageCount int    `json:"message_count"`
	LastMessage  string `json:"last_message"`
}

// База пользователей
var users = make(map[string]*User)

// Файл хранения пользователей
const userDBFile = "users.json"

// Загрузка данных из файла
func LoadUsers() {
	file, err := os.Open(userDBFile)
	if err != nil {
		log.Println("⚠️ Не удалось загрузить пользователей, создаем новый файл:", err)
		return
	}
	defer file.Close()

	err = json.NewDecoder(file).Decode(&users)
	if err != nil {
		log.Println("⚠️ Ошибка парсинга users.json:", err)
	}
}

// Сохранение данных в файл
func saveUsers() {
	file, err := os.Create(userDBFile)
	if err != nil {
		log.Println("❌ Ошибка сохранения users.json:", err)
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ") // Красивый вывод
	err = encoder.Encode(users)
	if err != nil {
		log.Println("❌ Ошибка записи в файл:", err)
	}
}

// Обновление информации о пользователе
func UpdateUser(nickname, message string) {
	if user, exists := users[nickname]; exists {
		user.MessageCount++
		user.LastMessage = message
		DetectName(nickname, message)
	} else {
		users[nickname] = &User{
			Name:         "", // Будем заполнять позже, если представится
			FirstSeen:    time.Now().Format(time.RFC3339),
			MessageCount: 1,
			LastMessage:  message,
		}
		DetectName(nickname, message)
	}
	saveUsers()
}

// Функция для определения имени (если представился)
func DetectName(nickname, message string) {
	if users[nickname] != nil && users[nickname].Name == "" {
		message = strings.ToLower(strings.TrimSpace(message)) // Убираем лишние пробелы и приводим к нижнему регистру

		if strings.HasPrefix(message, "меня зовут") {
			name := strings.TrimSpace(strings.TrimPrefix(message, "меня зовут"))
			if len(name) > 1 {
				users[nickname].Name = name
				log.Println("👤 Пользователь представился:", users[nickname].Name)
				saveUsers()
			}
		} else if strings.HasPrefix(message, "я -") {
			name := strings.TrimSpace(strings.TrimPrefix(message, "я -"))
			if len(name) > 1 {
				users[nickname].Name = name
				log.Println("👤 Пользователь представился:", users[nickname].Name)
				saveUsers()
			}
		}
	}
}
