package main

import (
	"app/db"
	tools "app/handlers"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strconv"
	"strings"
)

var userSpeakerDir = make(map[int64]string)

const (
	bookCourseInfoPath         = "data/Инструкция по бронированию.txt"
	greetingMessage            = "Привет! 👋\nЯ помогу тебе выбрать лучший курс.\nВыбери, что интересно, и мы сразу подберём варианты!"
	contactConfirmationMessage = "Спасибо! 📲 Мы записали твой номер, менеджер скоро свяжется."
	speakerPromptMessage       = "Отличный выбор! 🎯\nТеперь выбери город 📍\nГде тебе будет удобно заниматься?"
	courseHeaderTemplate       = "Отправляю программу курса «%s» — посмотри материалы ниже."
	nextStepMessage            = "Что делаем дальше?"
	bookCourseFallbackMessage  = "Не удалось найти информацию о бронировании курса. Напишите нам, пожалуйста."
)

func HandleMessage(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	user := update.Message.From
	chatID := update.Message.Chat.ID
	phone := ""

	if update.Message.Contact != nil {
		phone = update.Message.Contact.PhoneNumber
	}

	err := db.UpsertUser(
		dbConn,
		chatID,
		phone,
		strings.TrimSpace(user.FirstName+" "+user.LastName),
		"",
		"",
	)
	if err != nil {
		log.Println("failed to upsert user:", err)
	}

	if update.Message.Contact != nil {
		msg := tgbotapi.NewMessage(chatID, contactConfirmationMessage)
		tools.SendAndLog(bot, msg)

		contactName := strings.TrimSpace(user.FirstName + " " + user.LastName)
		if c := update.Message.Contact; c != nil {
			candidate := strings.TrimSpace(c.FirstName + " " + c.LastName)
			if candidate != "" {
				contactName = candidate
			}
		}
		setSessionContact(chatID, phone, contactName)

		trySyncBitrixDeal(bot, chatID)
		return
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, greetingMessage)
	msg.ReplyMarkup = SpeakerKeyboard()
	tools.SendAndLog(bot, msg)
}

func HandleCallback(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	data := update.CallbackQuery.Data
	chatID := update.CallbackQuery.Message.Chat.ID

	switch {
	case strings.HasPrefix(data, "speaker_"):
		pickSpeaker(data, bot, chatID, update)
	case strings.HasPrefix(data, "course_"):
		pickCourse(data, bot, chatID, update)
	case data == "book_course":
		text, err := tools.ReadTextFile(bookCourseInfoPath)
		if err != nil {
			text = bookCourseFallbackMessage
		}

		msg := tgbotapi.NewMessage(chatID, text)
		msg.ReplyMarkup = ContactKeyboard()
		tools.SendAndLog(bot, msg)

	case data == "needed_tools":
		speakerDir := userSpeakerDir[chatID]
		msg := tgbotapi.NewMessage(chatID, tools.GetToolsText(speakerDir))
		tools.SendAndLog(bot, msg)
	}

	callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "")
	if _, err := bot.Request(callback); err != nil {
		log.Printf("failed to answer callback: %v", err)
	}
}

func pickSpeaker(data string, bot *tgbotapi.BotAPI, chatID int64, update tgbotapi.Update) {
	idx, _ := strconv.Atoi(strings.TrimPrefix(data, "speaker_"))
	speaker := Speakers[idx].Name
	user := update.CallbackQuery.From

	err := db.UpsertUser(
		dbConn,
		chatID,
		"",
		strings.TrimSpace(user.FirstName+" "+user.LastName),
		"",
		speaker,
	)
	if err != nil {
		log.Println("failed to update user speaker:", err)
	}

	msg := tgbotapi.NewMessage(chatID, speakerPromptMessage)
	msg.ReplyMarkup = CourseKeyboard(idx)
	tools.SendAndLog(bot, msg)

	setSessionCourse(chatID, speaker, "")
}

func pickCourse(data string, bot *tgbotapi.BotAPI, chatID int64, update tgbotapi.Update) {
	parts := strings.Split(strings.TrimPrefix(data, "course_"), "_")
	if len(parts) < 2 {
		return
	}

	speakerIdx, _ := strconv.Atoi(parts[0])
	courseIdx, _ := strconv.Atoi(parts[1])

	course := Speakers[speakerIdx].Courses[courseIdx]
	city := course.City
	user := update.CallbackQuery.From

	err := db.UpsertUser(
		dbConn,
		chatID,
		"",
		strings.TrimSpace(user.FirstName+" "+user.LastName),
		city,
		"",
	)
	if err != nil {
		log.Println("failed to update user city:", err)
	}

	programPath := strings.TrimPrefix(course.Program, "/")
	speakerDir := programPath
	if strings.Contains(programPath, "/") {
		speakerDir = strings.SplitN(programPath, "/", 2)[0]
	} else if strings.Contains(programPath, "\\") {
		speakerDir = strings.SplitN(programPath, "\\", 2)[0]
	}
	userSpeakerDir[chatID] = speakerDir

	courseTitle := strings.TrimSpace(course.City)
	if courseTitle == "" {
		courseTitle = "курс"
	}

	header := tgbotapi.NewMessage(chatID, fmt.Sprintf(courseHeaderTemplate, courseTitle))
	tools.SendAndLog(bot, header)

	if err := tools.SendCourseProgram(bot, chatID, course.Program); err != nil {
		log.Printf("failed to send course program: %v", err)
		return
	}

	msg := tgbotapi.NewMessage(chatID, nextStepMessage)
	msg.ReplyMarkup = CourseActionKeyboard()
	tools.SendAndLog(bot, msg)

	speakerName := Speakers[speakerIdx].Name
	setSessionCourse(chatID, speakerName, city)

	trySyncBitrixDeal(bot, chatID)
}
