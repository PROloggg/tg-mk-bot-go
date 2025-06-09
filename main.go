package main

import (
	"app/db"
	"database/sql"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"log"
	"os"
)

var dbConn *sql.DB

func main() {
	_ = godotenv.Load()
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_TOKEN"))
	if err != nil {
		log.Panic(err)
	}

	err = LoadSpeakersFromCSV("data/courses.csv")
	if err != nil {
		panic(err)
	}

	dbConn, err = db.InitDB()
	if err != nil {
		log.Fatal(err)
	}
	defer dbConn.Close()

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			HandleMessage(bot, update)
		}
		if update.CallbackQuery != nil {
			HandleCallback(bot, update)
		}
	}
}
