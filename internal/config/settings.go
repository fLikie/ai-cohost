package config

import (
	"encoding/json"
	"log"
	"os"
)

type Config struct {
	TwitchChannel   string  `json:"twitch_channel"`
	SelectedVoice   string  `json:"selected_voice"`
	VolumeLevel     float64 `json:"volume_level"`
	SelectedAiModel string  `json:"selected_ai_model"`
	SelectedTTS     string  `json:"selected_tts"`
}

var Settings Config

const settingsFile = "settings.json"

// 📂 Загружаем настройки
func LoadSettings() {
	file, err := os.ReadFile(settingsFile)
	if err != nil {
		log.Println("⚠️ Файл настроек не найден, создаём новый...")
		SaveSettings() // Если файла нет, создаём новый
		return
	}

	if err := json.Unmarshal(file, &Settings); err != nil {
		log.Println("❌ Ошибка загрузки настроек:", err)
		return
	}
}

// 💾 Сохраняем настройки в файл
func SaveSettings() {
	data, err := json.MarshalIndent(Settings, "", "  ")
	if err != nil {
		log.Println("❌ Ошибка сохранения настроек:", err)
		return
	}
	err = os.WriteFile(settingsFile, data, 0644)
	if err != nil {
		log.Println("Ошибка записи настроек:", err)
		return
	}
}
