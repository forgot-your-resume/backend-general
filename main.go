package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/google/uuid"
	"github.com/rs/cors"
)

// Пользовательская структура
type User struct {
	ID       string `json:"id"`
	Login    string `json:"login"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

// Структура для входа
type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// Структура конференции
type Conference struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	DateTime  int64      `json:"dateTime"`
	Questions []Question `json:"questions"`
	Token     string     `json:"token"`
}

// Структура вопроса
type Question struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

// Структура для запроса токена
type TokenRequest struct {
	TokenType string `json:"tokenType"`
	Channel   string `json:"channel"`
	Role      string `json:"role"`
	UID       string `json:"uid"`
	Expire    int    `json:"expire"`
}

// Структура для ответа токена
type TokenResponse struct {
	Token string `json:"token"`
}

var (
	users       = make(map[string]User)
	conferences = make(map[string]Conference)
	usersMutex  sync.Mutex
	confsMutex  sync.Mutex
	dataDir     = "./data"
)

var AgoraAddr = ""

func main() {
	AgoraAddr = os.Getenv("AGORA_ADDR")
	if AgoraAddr == "" {
		log.Fatal("AGORA_ADDR not defined")
	}

	mux := http.NewServeMux()

	if err := os.MkdirAll(dataDir, os.ModePerm); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}

	loadUsers()
	loadConferences()

	mux.HandleFunc("/ping", pingHandler)
	mux.HandleFunc("/register", registerHandler)
	mux.HandleFunc("/login", loginHandler)
	mux.HandleFunc("/conferences", conferencesHandler)
	mux.HandleFunc("/add_question", addQuestionHandler)
	mux.HandleFunc("/get_questions", getQuestionsHandler)
	mux.HandleFunc("/create_conference", createConferenceHandler)

	handler := cors.Default().Handler(mux)

	log.Println("Server is running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("pong"))
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if user.Role != "соискатель" && user.Role != "рекрутер" && user.Role != "эксперт" {
		http.Error(w, "Invalid role", http.StatusBadRequest)
		return
	}

	userID := uuid.New().String()
	user.ID = userID

	usersMutex.Lock()
	users[user.Login] = user
	usersMutex.Unlock()

	saveUsers()

	w.WriteHeader(http.StatusCreated)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var loginReq LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	usersMutex.Lock()
	user, exists := users[loginReq.Login]
	usersMutex.Unlock()

	if !exists || user.Password != loginReq.Password {
		http.Error(w, "Invalid login or password", http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func conferencesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	confsMutex.Lock()
	confList := make([]Conference, 0, len(conferences))
	for _, conf := range conferences {
		confList = append(confList, conf)
	}
	confsMutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(confList)
}

func addQuestionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var data struct {
		ConfID   string `json:"confID"`
		Question string `json:"question"`
		Answer   string `json:"answer"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	confsMutex.Lock()
	conf, exists := conferences[data.ConfID]
	if exists {
		conf.Questions = append(conf.Questions, Question{Question: data.Question, Answer: data.Answer})
		conferences[data.ConfID] = conf
	}
	confsMutex.Unlock()

	if !exists {
		http.Error(w, "Conference not found", http.StatusNotFound)
		return
	}

	saveConferences()

	w.WriteHeader(http.StatusOK)
}

func getQuestionsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	confID := r.URL.Query().Get("confID")
	if confID == "" {
		http.Error(w, "Conference ID is required", http.StatusBadRequest)
		return
	}

	confsMutex.Lock()
	conf, exists := conferences[confID]
	confsMutex.Unlock()

	if !exists {
		http.Error(w, "Conference not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(conf.Questions)
}

func createConferenceHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var data struct {
		DateTime int64  `json:"dateTime"`
		Name     string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Создаем запрос для получения токена
	tokenReq := TokenRequest{
		TokenType: "rtc",
		Channel:   "my-video-channel",
		Role:      "publisher",
		UID:       "user123",
		Expire:    3600,
	}

	token, err := getAgoraToken(tokenReq)
	if err != nil {
		http.Error(w, "Failed to get token", http.StatusInternalServerError)
		return
	}

	confID := uuid.New().String()

	confsMutex.Lock()
	conferences[confID] = Conference{
		ID:       confID,
		Name:     data.Name,
		DateTime: data.DateTime,
		Token:    token,
	}
	confsMutex.Unlock()

	saveConferences()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(conferences[confID])
}

func getAgoraToken(req TokenRequest) (string, error) {
	jsonReq, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	resp, err := http.Post(AgoraAddr+"/getToken", "application/json", bytes.NewBuffer(jsonReq))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get token, status code: %d", resp.StatusCode)
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", err
	}

	return tokenResp.Token, nil
}

func loadUsers() {
	file, err := os.ReadFile(filepath.Join(dataDir, "users.json"))
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		log.Fatalf("Failed to load users: %v", err)
	}

	if err := json.Unmarshal(file, &users); err != nil {
		log.Fatalf("Failed to parse users: %v", err)
	}
}

func saveUsers() {
	file, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		log.Fatalf("Failed to encode users: %v", err)
	}

	if err := os.WriteFile(filepath.Join(dataDir, "users.json"), file, 0644); err != nil {
		log.Fatalf("Failed to save users: %v", err)
	}
}

func loadConferences() {
	file, err := os.ReadFile(filepath.Join(dataDir, "conferences.json"))
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		log.Fatalf("Failed to load conferences: %v", err)
	}

	if err := json.Unmarshal(file, &conferences); err != nil {
		log.Fatalf("Failed to parse conferences: %v", err)
	}
}

func saveConferences() {
	file, err := json.MarshalIndent(conferences, "", "  ")
	if err != nil {
		log.Fatalf("Failed to encode conferences: %v", err)
	}

	if err := os.WriteFile(filepath.Join(dataDir, "conferences.json"), file, 0644); err != nil {
		log.Fatalf("Failed to save conferences: %v", err)
	}
}
