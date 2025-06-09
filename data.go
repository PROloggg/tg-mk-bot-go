package main

import (
	"encoding/csv"
	"os"
)

var Speakers []Speaker

func LoadSpeakersFromCSV(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return err
	}

	speakersMap := make(map[string]*Speaker)
	for _, rec := range records[1:] { // пропускаем заголовок
		name, city, date, program := rec[0], rec[1], rec[2], rec[3]
		if speakersMap[name] == nil {
			speakersMap[name] = &Speaker{Name: name}
		}
		speakersMap[name].Courses = append(speakersMap[name].Courses, Course{
			City:    city,
			Date:    date,
			Program: program,
		})
	}
	Speakers = nil
	for _, s := range speakersMap {
		Speakers = append(Speakers, *s)
	}
	return nil
}
