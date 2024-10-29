package storage

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/Anna-Tregub/go_final_project/internal/tasks"
	"github.com/Anna-Tregub/go_final_project/models"
)

type Store struct {
	db *sql.DB
}

// Открываем/создаем  БД
func OpenDataBase() *sql.DB {
	appPath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	dbFile := filepath.Join(filepath.Dir(appPath), "scheduler.db")
	_, err = os.Stat(dbFile)

	var install bool
	if err != nil {
		install = true
	}

	db, err := sql.Open("sqlite", "scheduler.db")
	if err != nil {
		log.Fatal(err)
	}

	if install {
		_, err = db.Exec(`CREATE TABLE IF NOT EXISTS scheduler (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	date CHAR(8) NOT NULL DEFAULT '',
	title VARCHAR(128) NOT NULL DEFAULT '',
	comment VARCHAR(256) NOT NULL DEFAULT '',
	repeat VARCHAR(128) NOT NULL DEFAULT ''
	);`)

		if err != nil {
			log.Fatal(err)
		}

		_, err = db.Exec(`CREATE INDEX IF NOT EXISTS scheduler_date ON scheduler (date);`)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("База данных создана")

	} else {
		log.Println("База данных была создана ранее")
	}
	return db
}

func NewStore(db *sql.DB) Store {
	return Store{db: db}
}

// Добавление задачи
func (s *Store) AddTask(t models.Task) (string, error) {
	var err error

	if t.Title == "" {
		return "", fmt.Errorf(`{"error":"Не указан заголовок задачи"}`)
	}

	if t.Date == "" {
		t.Date = time.Now().Format(models.DateFormat)
	}

	_, err = time.Parse(models.DateFormat, t.Date)
	if err != nil {
		return "", fmt.Errorf(`{"error":"Некорректный формат даты"}`)
	}

	if t.Date < time.Now().Format(models.DateFormat) {
		if t.Repeat != "" {
			nextDate, err := tasks.NextDate(time.Now(), t.Date, t.Repeat)
			if err != nil {
				return "", fmt.Errorf(`{"error":"Некорректное правило повторения"}`)
			}
			t.Date = nextDate
		} else {
			t.Date = time.Now().Format(models.DateFormat)
		}
	}

	query := `INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`
	result, err := s.db.Exec(query, t.Date, t.Title, t.Comment, t.Repeat)
	if err != nil {
		return "", fmt.Errorf(`{"error":"Не удалось добавить задачу"}`)
	}

	// Возвращаем идентификатор добавленной задачи
	id, err := result.LastInsertId()
	if err != nil {
		return "", fmt.Errorf(`{"error":"Не удалось вернуть id новой задачи"}`)
	}
	return fmt.Sprintf("%d", id), nil
}

// Получаем список задач по фильтрам
func (s *Store) GetTasks(search string) ([]models.Task, error) {
	var t models.Task
	var tasks []models.Task
	var rows *sql.Rows
	var err error
	if search == "" {
		rows, err = s.db.Query("SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date LIMIT ?", models.MaxTasks)
	} else if date, error := time.Parse("02.01.2006", search); error == nil {
		query := "SELECT id, date, title, comment, repeat FROM scheduler WHERE date = ? ORDER BY date LIMIT ?"
		rows, err = s.db.Query(query, date.Format(models.DateFormat), models.MaxTasks)

	} else {
		search = "%%%" + search + "%%%"
		query := "SELECT id, date, title, comment, repeat FROM scheduler WHERE title LIKE ? OR comment LIKE ? ORDER BY date LIMIT ?"
		rows, err = s.db.Query(query, search, search, models.MaxTasks)
	}
	if err != nil {
		return []models.Task{}, fmt.Errorf(`{"error":"ошибка запроса"}`)
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&t.ID, &t.Date, &t.Title, &t.Comment, &t.Repeat)
		if err = rows.Err(); err != nil {
			return []models.Task{}, fmt.Errorf(`{"error":"Ошибка распознавания данных"}`)
		}
		tasks = append(tasks, t)
	}
	if len(tasks) == 0 {
		tasks = []models.Task{}
	}

	return tasks, nil
}

// Получение задачи по id
func (s *Store) GetTask(id string) (models.Task, error) {
	var t models.Task
	if id == "" {
		return models.Task{}, fmt.Errorf(`{"error":"Не указан id"}`)
	}
	row := s.db.QueryRow("SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?", id)
	err := row.Scan(&t.ID, &t.Date, &t.Title, &t.Comment, &t.Repeat)
	if err != nil {
		return models.Task{}, fmt.Errorf(`{"error":"Задача не найдена"}`)
	}
	return t, nil
}

// Редактирование задачи
func (s *Store) UpdateTask(t models.Task) error {

	if t.ID == "" {
		return fmt.Errorf(`{"error":"Не указан id"}`)
	}

	if t.Title == "" {
		return fmt.Errorf(`{"error":"Не указан заголовок задачи"}`)
	}

	if t.Date == "" {
		t.Date = time.Now().Format(models.DateFormat)
	}

	_, err := time.Parse(models.DateFormat, t.Date)
	if err != nil {
		return fmt.Errorf(`{"error":"Некорректный формат даты"}`)
	}

	if t.Date < time.Now().Format(models.DateFormat) {
		if t.Repeat != "" {
			nextDate, err := tasks.NextDate(time.Now(), t.Date, t.Repeat)
			if err != nil {

				return fmt.Errorf(`{"error":"Некорректное правило повторения"}`)
			}
			t.Date = nextDate
		} else {
			t.Date = time.Now().Format(models.DateFormat)
		}
	}

	// Обновляем задачу в базе
	query := `UPDATE scheduler SET date=?, title=?, comment=?, repeat=? WHERE id=?`
	result, err := s.db.Exec(query, t.Date, t.Title, t.Comment, t.Repeat, t.ID)
	if err != nil {

		return fmt.Errorf(`{"error":"Задача с таким id не найдена"}`)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf(`{"error":"Не удалось посчитать измененные строки"}`)
	}

	if rowsAffected == 0 {
		return fmt.Errorf(`{"error":"Задача с таким id не найдена"}`)
	}

	return nil
}

// Выполнение задачи
func (s *Store) TaskDone(id string) error {
	var t models.Task

	t, err := s.GetTask(id)
	if err != nil {
		return err
	}
	if t.Repeat == "" {

		err := s.DeleteTask(id)
		if err != nil {
			return err
		}

	} else {
		next, err := tasks.NextDate(time.Now(), t.Date, t.Repeat)
		if err != nil {
			return err
		}
		t.Date = next
		err = s.UpdateTask(t)
		if err != nil {
			return err
		}
	}

	return nil
}

// Удаление задачи из БД
func (s *Store) DeleteTask(id string) error {

	if id == "" {
		return fmt.Errorf(`{"error":"Не указан id"}`)
	}
	query := "DELETE FROM scheduler WHERE id = ?"
	result, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf(`{"error":"Не удалось удалить задачу"}`)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {

		return fmt.Errorf(`{"error":"Не удалось посчитать измененные строки"}`)
	}

	if rowsAffected == 0 {

		return fmt.Errorf(`{"error":"Задача с таким id не найдена"}`)
	}

	return nil
}
