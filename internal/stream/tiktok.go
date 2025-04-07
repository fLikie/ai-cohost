package stream

import (
	"bufio"
	"cohost/internal/config"
	"cohost/internal/response"
	gui "cohost/internal/ui"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
)

func ensureTikTokLogin() error {
	// Проверяем наличие папки с авторизацией
	if _, err := os.Stat("tiktok_profile"); os.IsNotExist(err) {
		fmt.Println("🔐 Авторизация TikTok не найдена. Запускаем логин...")
		cmd := exec.Command("python", "internal/stream/tiktok_login.py")

		// Подключаем stdout/stderr к терминалу Go
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("ошибка запуска tiktok_login.py: %w", err)
		}
		fmt.Println("✅ Авторизация завершена.")
	} else {
		fmt.Println("✅ Авторизация уже есть, пропускаем вход.")
	}

	return nil
}

func StartTikTokListener() {

	if err := ensureTikTokLogin(); err != nil {
		log.Println("❌ Нет логина:", err)
		return
	}

	log.Println("Запускаем тикток листенер")
	cmd := exec.Command("python", "internal/stream/tiktok_listener.py", config.Settings.TikTokUsername)

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		log.Println("❌ Не удалось запустить TikTok listener:", err)
		return
	}

	// Вывод ошибок
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			log.Println("🐍 TikTok stderr:", scanner.Text())
		}
	}()

	// Обработка комментариев
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			var data map[string]string
			if err := json.Unmarshal([]byte(line), &data); err == nil {
				user := data["user"]
				comment := data["comment"]
				log.Printf("💬 TikTok %s: %s\n", user, comment)

				// Отправляем в AI
				aiResponse := response.GenerateAIResponse(user, comment)
				gui.SetChatText(fmt.Sprintf("🤖 AI: %s", aiResponse))
				gui.SetUsersText(fmt.Sprintf("💬 TikTok %s: %s\n", user, comment))
				gui.AppendToChatHistory(fmt.Sprintf("🤖 AI: %s", aiResponse))
				gui.AppendToChatHistory(fmt.Sprintf("%s: %s", user, comment))

			}
		}
	}()
}
