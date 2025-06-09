package main

import (
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Клавиатура с кнопками Спикеров
func SpeakerKeyboard() tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	for i, s := range Speakers {
		btn := tgbotapi.NewInlineKeyboardButtonData(s.Name, "speaker_"+strconv.Itoa(i))
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(btn))
	}
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// Клавиатура с кнопками Курсов для выбранного Спикера
func CourseKeyboard(speakerIdx int) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	for i, c := range Speakers[speakerIdx].Courses {
		btn := tgbotapi.NewInlineKeyboardButtonData(
			c.City+" ("+c.Date+")",
			"course_"+strconv.Itoa(speakerIdx)+"_"+strconv.Itoa(i),
		)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(btn))
	}
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// Клавиатура с кнопками "Забронировать место" и "Инструменты для обучения"
func CourseActionKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔒 Забронировать место", "book_course"),
			tgbotapi.NewInlineKeyboardButtonData("🧰 Инструменты для обучения", "needed_tools"),
		),
	)
}

// Обычная клавиатура с кнопкой "Отправить телефон"
func ContactKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButtonContact("📞 Отправить телефон"),
		),
	)
}
