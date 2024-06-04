package main

import (
    "encoding/json"
    "log"
    "net/http"
)

// Структура для примера данных
type ExampleData struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}

// Точка /hello
func helloHandler(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("Hello, World!"))
}

// Точка /data
func dataHandler(w http.ResponseWriter, r *http.Request) {
    data := ExampleData{ID: 1, Name: "Sample Data"}
    jsonData, err := json.Marshal(data)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.Write(jsonData)
}

func main() {
    http.HandleFunc("/hello", helloHandler)
    http.HandleFunc("/data", dataHandler)

    log.Println("Starting server on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
