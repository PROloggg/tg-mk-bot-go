package main

import (
	"app/db"
	tools "app/handlers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
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
		tools.SendAndLog(bot, msg)

		return
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Привет! 👋 Я помогу тебе найти информацию о наших курсах. Давай начнем!")
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
		text, err := tools.ReadTextFile("data/Инструкция по бронированию.txt")
		if err != nil {
			text = "Не удалось загрузить инструкцию по бронированию."
		}

		msg := tgbotapi.NewMessage(chatID, text)
		msg.ReplyMarkup = ContactKeyboard()
		tools.SendAndLog(bot, msg)

	case data == "needed_tools":
		speakerDir := userSpeakerDir[chatID]
		msg := tgbotapi.NewMessage(chatID, tools.GetToolsText(speakerDir))
		tools.SendAndLog(bot, msg)
	}

	// Отвечаем на callback (через Request, а не через SendAndLog!)
	callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "")
	if _, err := bot.Request(callback); err != nil {
		log.Printf("Ошибка отправки callback: %v", err)
	}

}

// pickSpeaker обрабатывает выбор спикера и запрашивает город и дату обучения
func pickSpeaker(data string, bot *tgbotapi.BotAPI, chatID int64, update tgbotapi.Update) {
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
	tools.SendAndLog(bot, msg)
}

// pickCourse обрабатывает выбор курса и отправляет информацию о программе
func pickCourse(data string, bot *tgbotapi.BotAPI, chatID int64, update tgbotapi.Update) {
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
	speakerDir := programPath
	if strings.Contains(programPath, "/") {
		speakerDir = strings.SplitN(programPath, "/", 2)[0]
	} else if strings.Contains(programPath, "\\") {
		speakerDir = strings.SplitN(programPath, "\\", 2)[0]
	}
	userSpeakerDir[chatID] = speakerDir

	// Отправляем заголовок
	header := tgbotapi.NewMessage(chatID, "Вот полная программа курса 👇")
	tools.SendAndLog(bot, header)

	// Отправляем саму программу (текст или файл)
	err = tools.SendCourseProgram(bot, chatID, course.Program)
	if err != nil {
		log.Print("Ошибка отправки программы курса:", err)
		return
	}

	msg := tgbotapi.NewMessage(chatID, "Выберите действие:")
	msg.ReplyMarkup = CourseActionKeyboard()
	tools.SendAndLog(bot, msg)
}
