package questions

import (
	_ "embed"
	"encoding/json"
	"math/rand"
)

//go:embed questions.json
var questionFile string

// Структура вопроса
type Question struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

var Questions = func() []Question {
	var questions []Question
	err := json.Unmarshal([]byte(questionFile), &questions)

	if err != nil {
		panic(err)
	}

	return questions
} ()


// Функция для получения случайных вопросов
func GetRandomQuestions(quantity int) []Question {
	if quantity >= len(Questions) {
		return Questions
	}

	perm := rand.Perm(len(Questions))

	randomQuestions := make([]Question, quantity)
	for i := 0; i < quantity; i++ {
		randomQuestions[i] = Questions[perm[i]]
	}

	return randomQuestions
}