package main

import (
	"log"
	"net/http"

	"github.com/Anna-Tregub/go_final_project/internal/handlers"
	"github.com/Anna-Tregub/go_final_project/internal/storage"

	_ "github.com/mattn/go-sqlite3"
	_ "modernc.org/sqlite"
)

func main() {

	dataBase := storage.OpenDataBase()
	defer dataBase.Close()
	store := storage.NewStore(dataBase)

	fileServer := http.FileServer(http.Dir("./web"))

	http.Handle("/", fileServer)

	http.HandleFunc("/api/nextdate", handlers.NextDateHandler)
	http.HandleFunc("GET /api/task", handlers.TaskGetHandler(store))
	http.HandleFunc("POST /api/task", handlers.TaskPostHandler(store))
	http.HandleFunc("PUT /api/task", handlers.TaskPutHandler(store))
	http.HandleFunc("DELETE /api/task", handlers.TaskDeleteHandler(store))
	http.HandleFunc("/api/tasks", handlers.TasksGetHandler(store))
	http.HandleFunc("/api/task/done", handlers.TaskDoneHandler(store))

	port := ":7540"
	log.Println("Приложение запущено на порту", port)

	err := http.ListenAndServe(port, nil)
	if err != nil {
		panic(err)
	}
}
