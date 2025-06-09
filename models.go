package main

type Course struct {
	SpeakerIdx int
	City       string
	Date       string
	Program    string
}

type Speaker struct {
	Name    string
	Courses []Course
}
