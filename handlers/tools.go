package handlers

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"os"
	"path/filepath"
)

func ReadTextFile(path string) (string, error) {
	bytes, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		log.Println("Ошибка чтения файла:", err)
		return "", err
	}
	return string(bytes), nil
}

// GetToolsText Возвращает текст файла инструментов: сначала ищет в папке спикера, потом общий
func GetToolsText(speakerDir string) string {
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

// SendAndLog Отправляет сообщение и логирует ошибку, если она есть
func SendAndLog(bot *tgbotapi.BotAPI, msg tgbotapi.Chattable) {
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Ошибка отправки сообщения: %v", err)
	}
}
