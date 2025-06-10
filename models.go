package main

type Course struct {
	SpeakerIdx int
	City       string
	Program    string
}

type Speaker struct {
	Name    string
	Courses []Course
}
