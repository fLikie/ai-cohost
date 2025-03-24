package audio

import (
	"bytes"
	"cohost/internal/config"
	"fmt"
	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/go-resty/resty/v2"
	"io"
	"io/ioutil"
	"log"
	"os"
	"time"
)

const AudioQueueMax = 10

var Voices = map[string]string{
	"Sarah":         "EXAVITQu4vr4xnSDxMaL",
	"Victoria":      "FZGeNF7bE3syeQOynDKC",
	"Oleg Krugliak": "m2gtxNsYBaIRqPBA5vU5",
	"Denophine":     "M1CSR3PJBsfWU6ZquG3C",
}

var audioQueue = make(chan []byte, AudioQueueMax) // Очередь звуков (до 10 сообщений)
var initialized = false
var volumeControl *effects.Volume // Глобальная переменная громкости
var volumeLevel = 1.0             // По умолчанию громкость 100%

// 🎵 Фоновый поток для воспроизведения очереди
func StartAudioPlayer() {
	go func() {
		for audioData := range audioQueue {
			playAudio(audioData) // Воспроизводим звук из очереди
		}
	}()
}

// 🔊 Функция добавления в очередь
func QueueAudio(audioData []byte) {
	select {
	case audioQueue <- audioData:
		log.Println("🎵 Добавлен звук в очередь")
	default:
		log.Println("⚠️ Очередь переполнена, пропускаем звук")
	}
}

// 🔊 Функция воспроизведения с регулировкой громкости
func playAudio(mp3Data []byte) {
	if len(mp3Data) == 0 {
		log.Println("❌ Ошибка: Пустой MP3-файл!")
		return
	}

	reader := io.NopCloser(bytes.NewReader(mp3Data))

	// ✅ Декодируем MP3
	streamer, format, err := mp3.Decode(reader)
	if err != nil {
		log.Println("❌ Ошибка декодирования MP3:", err)
		saveMP3Debug(mp3Data)
		return
	}

	// ✅ Инициализируем динамики только один раз
	if !initialized {
		speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
		initialized = true
	}

	// 🔥 Если `volumeControl` уже существует, заменяем `Streamer`
	if volumeControl != nil {
		speaker.Lock() // Блокируем поток перед обновлением
		volumeControl.Streamer = streamer
		speaker.Unlock() // Разблокируем поток
	} else {
		// 🔥 Создаём новый контроллер громкости
		volumeControl = &effects.Volume{
			Streamer: streamer,
			Base:     2,
			Volume:   volumeLevel, // Используем глобальную переменную громкости
			Silent:   false,
		}
	}

	// ✅ Воспроизводим с корректным завершением потока
	done := make(chan struct{})
	speaker.Play(beep.Seq(volumeControl, beep.Callback(func() {
		close(done)
	})))

	// ✅ Блокируем поток, пока звук не отыграет
	<-done
	log.Println("🎵 Аудио воспроизведение завершено!")
}

func saveMP3Debug(mp3Data []byte) {
	err := ioutil.WriteFile("test_raw.txt", mp3Data, 0644)
	if err != nil {
		log.Println("❌ Ошибка сохранения MP3 в текстовый формат:", err)
	} else {
		log.Println("✅ Сырые данные MP3 сохранены как test_raw.txt")
	}
}

func UpdateVolume(newVolume float64) {
	if volumeControl != nil {
		speaker.Lock() // Блокируем поток перед изменением громкости

		// 🔥 Пересчитываем громкость: -5 = почти тишина, 0 = нормальная, +2 = громче
		volumeControl.Volume = (newVolume * 2) - 2 // Переводим 0-1 в -5 до +2

		speaker.Unlock() // Разблокируем поток
		log.Println("🔊 Громкость обновлена:", volumeControl.Volume)
	}
}

// 🔊 Генерация аудио через ElevenLabs
func GenerateVoice(text string) {
	client := resty.New()
	voiceID := Voices[config.Settings.SelectedVoice] // Получаем ID выбранного голоса
	url := fmt.Sprintf("https://api.elevenlabs.io/v1/text-to-speech/%s", voiceID)

	elevenLabsKey := os.Getenv("ELEVENLABS_KEY")

	resp, err := client.R().
		SetHeader("xi-api-key", elevenLabsKey).
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]interface{}{
			"text":             text, // 🔥 Новый голос "Sarah"
			"model_id":         "eleven_multilingual_v2",
			"language_id":      "ru",
			"stability":        volumeLevel, // 📢 Громкость влияет на стабильность звука
			"similarity_boost": 0.8,         // Оптимальное значение
		}).
		Post(url)

	if err != nil {
		log.Println("Ошибка ElevenLabs:", err)
		return
	}

	log.Println("✅ MP3 успешно получен, размер:", len(resp.Body()), "байт, добавляем в очередь")
	QueueAudio(resp.Body()) // Добавляем звук в очередь
}
