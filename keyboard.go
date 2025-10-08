package main

import (
	"fmt"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func SpeakerKeyboard() tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	for i, s := range Speakers {
		label := fmt.Sprintf("🎓 %s ✂️", s.Name)
		btn := tgbotapi.NewInlineKeyboardButtonData(label, "speaker_"+strconv.Itoa(i))
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(btn))
	}
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func CourseKeyboard(speakerIdx int) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	for i, c := range Speakers[speakerIdx].Courses {
		btn := tgbotapi.NewInlineKeyboardButtonData(
			c.City,
			"course_"+strconv.Itoa(speakerIdx)+"_"+strconv.Itoa(i),
		)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(btn))
	}
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func CourseActionKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📝 Оставить заявку", "book_course"),
			tgbotapi.NewInlineKeyboardButtonData("❓ Как оплатить", "needed_tools"),
		),
	)
}

func ContactKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButtonContact("📱 Поделиться номером 📱"),
		),
	)
}
