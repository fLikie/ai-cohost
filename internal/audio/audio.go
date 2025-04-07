package audio

import (
	"bytes"
	"cohost/internal/common"
	"cohost/internal/config"
	"fmt"
	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
	"github.com/go-resty/resty/v2"
	"io"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"time"
)

const AudioQueueMax = 10

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
func playAudio(audioData []byte) {
	if len(audioData) == 0 {
		log.Println("❌ Ошибка: Пустой аудиофайл!")
		return
	}

	reader := bytes.NewReader(audioData)
	buf := make([]byte, 12)
	if _, err := reader.Read(buf); err != nil {
		log.Println("❌ Не удалось прочитать заголовок аудиофайла:", err)
		return
	}
	reader.Seek(0, io.SeekStart) // Сброс к началу

	var streamer beep.StreamSeekCloser
	var format beep.Format
	var err error

	switch {
	case bytes.HasPrefix(buf, []byte("RIFF")):
		// WAV-файл
		log.Println("📀 Определён формат: WAV")
		streamer, format, err = wav.Decode(reader)
	case bytes.HasPrefix(buf, []byte("\xFF\xFB")) || bytes.HasPrefix(buf, []byte("\x49\x44\x33")):
		// MP3-файл (начинается с FF FB или ID3)
		log.Println("📀 Определён формат: MP3")
		streamer, format, err = mp3.Decode(reader)
	default:
		log.Println("❌ Неизвестный формат аудио")
		saveMP3Debug(audioData)
		return
	}

	if err != nil {
		log.Println("❌ Ошибка декодирования:", err)
		saveMP3Debug(audioData)
		return
	}

	if !initialized {
		speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
		initialized = true
	}

	if volumeControl != nil {
		speaker.Lock()
		volumeControl.Streamer = streamer
		speaker.Unlock()
	} else {
		volumeControl = &effects.Volume{
			Streamer: streamer,
			Base:     2,
			Volume:   (config.Settings.VolumeLevel * 2) - 2,
			Silent:   false,
		}
	}

	done := make(chan struct{})
	speaker.Play(beep.Seq(volumeControl, beep.Callback(func() {
		close(done)
	})))
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
	voiceID := common.Voices[config.Settings.SelectedVoice] // Получаем ID выбранного голоса
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

func cleanTextForTTS(text string) string {
	re := regexp.MustCompile(`[^\p{L}\p{N}\p{P}\p{Z}]`) // убираем эмодзи и странные символы
	return re.ReplaceAllString(text, "")
}

func GenerateSileroVoice(text string) {
	client := resty.New()
	url := "http://185.21.142.27/speak"

	cleaned := cleanTextForTTS(text)

	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody([]string{cleaned}).
		Post(url)

	if err != nil {
		log.Println("Ошибка Silero TTS:", err)
		return
	}

	log.Println("✅ MP3 успешно получен, размер:", len(resp.Body()), "байт, добавляем в очередь")
	QueueAudio(resp.Body()) // Добавляем звук в очередь
}
