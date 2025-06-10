package handlers

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"path/filepath"
	"strings"
)

// SendCourseProgram Отправляем информацию по курсу
func SendCourseProgram(bot *tgbotapi.BotAPI, chatID int64, program string) error {
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
