package main

import (
	"app/db"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var userSpeakerDir = make(map[int64]string)

func HandleMessage(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	user := update.Message.From
	chatID := update.Message.Chat.ID
	phone := ""

	if update.Message.Contact != nil {
		phone = update.Message.Contact.PhoneNumber
	}

	// Сохраняем пользователя в базу
	err := db.UpsertUser(
		dbConn,
		chatID,
		phone, // телефон можно запросить позже
		user.FirstName+" "+user.LastName,
		"", // город можно запросить позже
		"", // куратор можно запросить позже
	)
	if err != nil {
		log.Println("Ошибка сохранения пользователя:", err)
	}

	if update.Message.Contact != nil {
		// Если пользователь отправил контакт, благодарим его
		msg := tgbotapi.NewMessage(chatID, "💇‍♀️ Спасибо, ждем вас на мастер-классе! 💇🏻")
		bot.Send(msg)
		return
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Привет! 👋 Я помогу тебе найти информацию о наших курсах. Давай начнем!")
	msg.ReplyMarkup = SpeakerKeyboard()
	bot.Send(msg)
}

func HandleCallback(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	data := update.CallbackQuery.Data
	chatID := update.CallbackQuery.Message.Chat.ID
	switch {
	case strings.HasPrefix(data, "speaker_"):
		idx, _ := strconv.Atoi(strings.TrimPrefix(data, "speaker_"))
		speaker := Speakers[idx].Name
		user := update.CallbackQuery.From
		// Сохраняем спикера
		err := db.UpsertUser(
			dbConn,
			chatID,
			"", // телефон не меняем
			user.FirstName+" "+user.LastName,
			"", // город не меняем
			speaker,
		)
		if err != nil {
			log.Println("Ошибка сохранения куратора:", err)
		}
		msg := tgbotapi.NewMessage(chatID, "В каком городе и когда ты хочешь пройти обучение?")
		msg.ReplyMarkup = CourseKeyboard(idx)
		bot.Send(msg)
	case strings.HasPrefix(data, "course_"):
		parts := strings.Split(strings.TrimPrefix(data, "course_"), "_")
		speakerIdx, _ := strconv.Atoi(parts[0])
		courseIdx, _ := strconv.Atoi(parts[1])
		course := Speakers[speakerIdx].Courses[courseIdx]
		city := course.City
		user := update.CallbackQuery.From

		// Сохраняем город
		err := db.UpsertUser(
			dbConn,
			chatID,
			"", // телефон не меняем
			user.FirstName+" "+user.LastName,
			city,
			"", // куратор не меняем
		)
		if err != nil {
			log.Println("Ошибка сохранения города:", err)
		}

		// Сохраняем имя папки спикера для этого пользователя
		programPath := strings.TrimPrefix(course.Program, "/")
		speakerDir := strings.SplitN(programPath, "/", 2)[0]
		userSpeakerDir[chatID] = speakerDir

		// Отправляем заголовок
		header := tgbotapi.NewMessage(chatID, "Вот полная программа курса 👇")
		bot.Send(header)

		// Отправляем саму программу (текст или файл)
		sendCourseProgram(bot, chatID, course.Program)

		msg := tgbotapi.NewMessage(chatID, "Выберите действие:")
		msg.ReplyMarkup = CourseActionKeyboard()
		bot.Send(msg)
	case data == "book_course":
		text, err := readTextFile("data/Инструкция по бронированию.txt")
		if err != nil {
			text = "Не удалось загрузить инструкцию по бронированию."
		}

		msg := tgbotapi.NewMessage(chatID, text)
		msg.ReplyMarkup = ContactKeyboard()
		bot.Send(msg)
	case data == "needed_tools":
		speakerDir := userSpeakerDir[chatID]
		msg := tgbotapi.NewMessage(chatID, getToolsText(speakerDir))
		bot.Send(msg)
	}

	bot.Send(tgbotapi.NewCallback(update.CallbackQuery.ID, ""))
}

func sendCourseProgram(bot *tgbotapi.BotAPI, chatID int64, program string) error {
	ext := strings.ToLower(filepath.Ext(program))
	baseDir := "data"

	switch ext {
	case ".jpg", ".jpeg", ".png":
		filePath := filepath.Join(baseDir, program)
		photo := tgbotapi.NewPhoto(chatID, tgbotapi.FilePath(filePath))
		photo.Caption = "Программа курса"
		_, err := bot.Send(photo)
		if err != nil {
			log.Println("Ошибка отправки фото:", err)
		}
		return err
	case ".pdf":
		filePath := filepath.Join(baseDir, program)
		doc := tgbotapi.NewDocument(chatID, tgbotapi.FilePath(filePath))
		doc.Caption = "Программа курса"
		_, err := bot.Send(doc)
		if err != nil {
			log.Println("Ошибка отправки PDF:", err)
		}
		return err
	default:
		msg := tgbotapi.NewMessage(chatID, program)
		_, err := bot.Send(msg)
		if err != nil {
			log.Println("Ошибка отправки текста:", err)
		}
		return err
	}
}

func readTextFile(path string) (string, error) {
	bytes, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		log.Println("Ошибка чтения файла:", err)
		return "", err
	}
	return string(bytes), nil
}

// Возвращает текст файла инструментов: сначала ищет в папке спикера, потом общий
func getToolsText(speakerDir string) string {
	// Путь к файлу в папке спикера
	speakerTools := filepath.Join("data", speakerDir, "Список инструментов.txt")
	if _, err := os.Stat(speakerTools); err == nil {
		text, err := os.ReadFile(speakerTools)
		if err == nil {
			return string(text)
		}
		log.Println("Ошибка чтения файла спикера:", err)
	}
	// Если нет — берем общий
	commonTools := filepath.Join("data", "Список инструментов.txt")
	text, err := os.ReadFile(commonTools)
	if err != nil {
		log.Println("Ошибка чтения общего файла инструментов:", err)
		return "Не удалось загрузить список инструментов."
	}
	return string(text)
}
