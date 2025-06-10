package main

import (
	"encoding/csv"
	"log"
	"os"
	"sort"
	"strings"
)

var Speakers []Speaker

func LoadSpeakersFromCSV(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Println("Ошибка чтения файла курсов:", err)
		}
	}(file)

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return err
	}

	speakersMap := make(map[string]*Speaker)
	for _, rec := range records[1:] { // пропускаем заголовок
		name, cityRaw, program := rec[0], rec[1], rec[2]

		// Разбиваем cityRaw по ";" и убираем пробелы
		cities := strings.Split(cityRaw, ";")
		for _, c := range cities {
			city := strings.TrimSpace(c)
			if city == "" {
				continue // пропустить пустые строки
			}

			if speakersMap[name] == nil {
				speakersMap[name] = &Speaker{Name: name}
			}

			course := Course{
				City:    city,
				Program: program,
			}

			speakersMap[name].Courses = append(speakersMap[name].Courses, course)
		}
	}

	Speakers = nil
	for _, s := range speakersMap {
		Speakers = append(Speakers, *s)
	}

	// Сортировка по алфавиту
	sort.Slice(Speakers, func(i, j int) bool {
		return Speakers[i].Name < Speakers[j].Name
	})

	return nil
}
