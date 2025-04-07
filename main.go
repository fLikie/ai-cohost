package main

import (
	"cohost/internal/audio"
	"cohost/internal/config"
	"cohost/internal/response"
	"cohost/internal/storage"
	"cohost/internal/stream"
	gui "cohost/internal/ui"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"
	"runtime/debug"
	"strings"
	"sync"
	"time"
)

// Очередь сообщений (🔄 буфер для накопления)
var messageQueue []string
var messageQueueLock sync.Mutex                    // 🔒 Блокировка для потокобезопасности
const messageProcessingInterval = 10 * time.Second // ⏳ Раз в 10 секунд

// Функция добавления сообщений в очередь
func queueMessage(username, message string) {
	messageQueueLock.Lock()
	defer messageQueueLock.Unlock()

	messageQueue = append(messageQueue, fmt.Sprintf("%s: %s", username, message))
	log.Println("📥 Добавлено сообщение в очередь:", username, "-", message)
}

// Функция обработки очереди
func processMessageQueue() {
	for {
		time.Sleep(messageProcessingInterval) // ⏳ Ждём 10 секунд

		messageQueueLock.Lock()
		if len(messageQueue) == 0 {
			messageQueueLock.Unlock()
			continue // Если нет сообщений, ничего не делаем
		}

		// 🔥 Объединяем все сообщения в один запрос к AI
		combinedMessages := strings.Join(messageQueue, "\n")
		messageQueue = nil // 🧹 Очищаем очередь

		messageQueueLock.Unlock()

		log.Println("🤖 Генерируем общий ответ для:", combinedMessages)

		// 📢 Генерируем AI-ответ
		//ai.GenerateAIResponse("чат", combinedMessages)
	}
}

// 🏁 Запуск GUI
func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Паника поймана!", r)
			debug.PrintStack()
			os.Exit(1)
		}
	}()

	log.Println("Запуск программы...")

	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️ .env файл не найден, использую переменные среды")
	}

	// ✅ Проверяем, что всё работает
	config.LoadSettings() // 📂 Загружаем голос и громкость
	storage.LoadUsers()   // 📂 Загружаем пользователей
	storage.LoadChatHistory()
	gui.CreateGUI(stream.StartTwitchBot, audio.StartAudioPlayer, response.ListenVoiceCommands, stream.StartTikTokListener)

	log.Println("Программа завершена.")
}
