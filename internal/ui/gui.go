package ui

import (
	"cohost/internal/audio"
	"cohost/internal/common"
	"cohost/internal/config"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"log"
	"os"
	"strings"
)

var (
	chatBox           *widget.Label
	aiResponseBox     *widget.Label
	gameEntry         *widget.Entry
	SelectedGame      string
	aiModelSelector   *widget.Select
	voiceSelector     *widget.Select
	volumeSlider      *widget.Slider
	volumeLabel       *widget.Label // Отображает текущую громкость
	botRunning        = false
	aiModels          = []string{"OpenAI", "DeepSeek"} // Доступные модели
	chatHistoryWindow fyne.Window
	chatHistoryText   *widget.Entry
)

func CreateGUI(
	twichBot func(),
	audioPlayer func(),
	listenVoiceCommands func(),
	listenTikTokListener func(),
) {
	log.Println("Создание GUI...")
	myApp := app.New()
	myWindow := myApp.NewWindow("AI Соведущий")

	log.Println("Инициализация UI элементов...")
	log.Println("✅ Пользователи загружены!")

	keys := make([]string, 0, len(common.Voices))
	for k := range common.Voices {
		keys = append(keys, k)
	}

	voiceSelector = widget.NewSelect(keys, func(value string) {
		config.Settings.SelectedVoice = value
		log.Println("🎙 Выбран голос:", config.Settings.SelectedVoice)
		config.SaveSettings() // 💾 Сохраняем в файл
	})
	voiceSelector.SetSelected(config.Settings.SelectedVoice) // 🎯 Устанавливаем загруженный голос

	// 🔊 Ползунок громкости
	volumeSlider = widget.NewSlider(0.0, 1.0)          // Диапазон от 0% до 100%
	volumeSlider.Step = 0.05                           // 🔥 Более точные шаги
	volumeSlider.SetValue(config.Settings.VolumeLevel) // 50% громкости = стандартное значение

	// 📢 Label для отображения текущей громкости
	volumeLabel = widget.NewLabel(fmt.Sprintf("🔊 Громкость: %d%%", int(config.Settings.VolumeLevel*100)))

	volumeSlider.OnChanged = func(value float64) {
		config.Settings.VolumeLevel = value
		volumeLabel.SetText(fmt.Sprintf("🔊 Громкость: %d%%", int(config.Settings.VolumeLevel*100)))
		log.Println("🔊 Громкость изменена:", config.Settings.VolumeLevel)
		audio.UpdateVolume(value)
		config.SaveSettings() // 💾 Сохраняем в файл
	}

	volumeSlider.SetValue(config.Settings.VolumeLevel) // 🎯 Устанавливаем загруженную громкость
	chatBox = widget.NewLabel("Чат подключён...")
	chatBox.Wrapping = fyne.TextWrapWord // 🔥 Перенос строк

	aiResponseBox = widget.NewLabel("🤖 AI здесь!")
	aiResponseBox.Wrapping = fyne.TextWrapWord // 🔥 Перенос строк

	gameEntry = widget.NewEntry()
	gameEntry.SetPlaceHolder("Введите название игры...")

	startButton := widget.NewButton("Запустить AI-Соведущего", func() {
		if botRunning {
			log.Println("⚠️ AI-соведущий уже работает!")
			return
		}
		botRunning = true

		SelectedGame = gameEntry.Text // Получаем введённое название игры
		if SelectedGame == "" {
			SelectedGame = "неизвестная игра" // По умолчанию, если игра не указана
		}
		log.Println("🎮 Выбрана игра:", SelectedGame)
		log.Println("🚀 AI Соведущий запущен с голосом:", config.Settings.SelectedVoice)
		go twichBot()
		go listenTikTokListener()
		listenVoiceCommands()
		audioPlayer()
		SetChatText(fmt.Sprintf("🤖 Соведущий активен! Играем в: %s", SelectedGame))
	})

	stopButton := widget.NewButton("🛑 Остановить бота", func() {
		if botRunning {
			log.Println("🛑 Останавливаем AI-соведущего...")
			botRunning = false // Выключаем флаг работы
		} else {
			log.Println("⚠️ Бот уже остановлен")
		}
	})

	chatContainer := container.NewVScroll(chatBox)
	chatContainer.SetMinSize(fyne.NewSize(0, 150)) // 150 пикселей по высоте

	aiResponseContainer := container.NewVScroll(aiResponseBox)
	aiResponseContainer.SetMinSize(fyne.NewSize(0, 150)) // 150 пикселей по высоте

	settingsButton := widget.NewButton("⚙ Настройки", func() {
		openSettingsWindow(myApp)
	})

	exitButton := widget.NewButton("❌ Выйти", func() {
		log.Println("🚪 Выход из программы...")
		os.Exit(0)
	})

	aiModelSelector = widget.NewSelect(aiModels, func(value string) {
		config.Settings.SelectedAiModel = value
		log.Println("🤖 Выбрана модель AI:", config.Settings.SelectedAiModel)
		config.SaveSettings() // 💾 Сохраняем настройки
	})
	aiModelSelector.SetSelected(config.Settings.SelectedAiModel) // 🎯 Загружаем сохранённую модель

	historyButton := widget.NewButton("📜 История чата", func() {
		openChatHistoryWindow(myApp)
	})

	myWindow.SetContent(container.NewVBox(
		widget.NewLabelWithStyle("🔴 AI Соведущий для стримов", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabel("Выберите голос:"),
		voiceSelector,
		widget.NewLabel("Настройка громкости:"),
		volumeSlider, // 🔊 Ползунок громкости
		volumeLabel,
		gameEntry,
		settingsButton,
		startButton,
		historyButton,
		chatContainer,
		aiResponseContainer,

		stopButton,
		exitButton,
	))

	myWindow.Resize(fyne.NewSize(500, 400))
	log.Println("Запуск GUI...")
	myWindow.ShowAndRun()
}

func openSettingsWindow(app fyne.App) {
	settingsWindow := app.NewWindow("⚙ Настройки")

	// Поле для ввода Twitch-канала
	twitchChannelEntry := widget.NewEntry()
	twitchChannelEntry.SetPlaceHolder("Введите Twitch-канал")
	twitchChannelEntry.SetText(config.Settings.TwitchChannel)

	aiModelSelector = widget.NewSelect(aiModels, func(value string) {
		config.Settings.SelectedAiModel = value
		log.Println("🤖 Выбрана модель AI:", config.Settings.SelectedAiModel)
		config.SaveSettings() // 💾 Сохраняем настройки
	})
	aiModelSelector.SetSelected(config.Settings.SelectedAiModel) // 🎯 Загружаем сохранённую модель

	// Выпадающий список TTS-сервисов
	ttsSelector := widget.NewSelect([]string{"ElevenLabs", "Yandex SpeechKit"}, func(value string) {
		config.Settings.SelectedTTS = value
		config.SaveSettings() // 💾 Сохраняем настройку
		log.Println("🎙 Выбран сервис TTS:", config.Settings.SelectedTTS)
	})
	ttsSelector.SetSelected(config.Settings.SelectedTTS) // Загружаем последнее сохраненное значение

	wakeWordEntry := widget.NewEntry()
	wakeWordEntry.SetPlaceHolder("Например: пятница")
	wakeWordEntry.SetText(config.Settings.WakeWord)

	tiktokEntry := widget.NewEntry()
	tiktokEntry.SetPlaceHolder("Tiktok username")
	tiktokEntry.SetText(config.Settings.TikTokUsername)

	// Кнопка сохранения настроек
	saveButton := widget.NewButton("💾 Сохранить", func() {
		config.Settings.TwitchChannel = twitchChannelEntry.Text
		config.Settings.TikTokUsername = tiktokEntry.Text
		config.Settings.WakeWord = wakeWordEntry.Text
		config.SaveSettings() // 💾 Сохраняем в файл
		settingsWindow.Close()
		log.Println("✅ Twitch-канал сохранён:", config.Settings.TwitchChannel)
	})

	settingsWindow.SetContent(container.NewVBox(
		widget.NewLabel("🔧 Twitch-настройки"),
		widget.NewLabel("Название канала:"),
		twitchChannelEntry,
		widget.NewLabel("Выбор AI модели:"),
		aiModelSelector,
		widget.NewLabel("Выбор Синтезатора голоса:"),
		ttsSelector,
		widget.NewLabel("Выбор кодового слова для активации соведущего"),
		wakeWordEntry,
		widget.NewLabel("Выбор канала тикток"),
		tiktokEntry,
		saveButton,
	))

	settingsWindow.Resize(fyne.NewSize(400, 150))
	settingsWindow.Show()
}

func openChatHistoryWindow(app fyne.App) {
	chatHistoryWindow = app.NewWindow("🗨 История чата")
	chatHistoryText = widget.NewMultiLineEntry()
	chatHistoryText.SetMinRowsVisible(20)
	chatHistoryText.Wrapping = fyne.TextWrapWord
	chatHistoryText.Disable() // только для чтения

	scroll := container.NewVScroll(chatHistoryText)

	chatHistoryWindow.SetContent(scroll)
	chatHistoryWindow.Resize(fyne.NewSize(500, 300))
	chatHistoryWindow.Show()
}

func AppendToChatHistory(text string) {
	if chatHistoryText != nil {
		chatHistoryText.SetText(chatHistoryText.Text + "\n" + text)
	}
}

func SetUsersText(text string) {
	formattedText := strings.ReplaceAll(text, "\n", " ") // Убираем лишние переносы
	chatBox.SetText(formattedText)
}

func SetChatText(text string) {
	formattedText := strings.ReplaceAll(text, "\n", " ") // Убираем лишние переносы
	aiResponseBox.SetText(formattedText)                 // Используем форматирование Fyne
}
