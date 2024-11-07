package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"github.com/gorilla/mux"
)
type Student struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Age      int    `json:"age"`
	Course   string `json:"course"`
	Email    string `json:"email"`
}
var (
	students   = make(map[int]Student)
	mu         sync.Mutex
	currentID  = 0
)
func getNextID() int {
	mu.Lock()
	defer mu.Unlock()
	currentID++
	return currentID
}

func createStudent(w http.ResponseWriter, r *http.Request) {
	var newStudent Student
	if err := json.NewDecoder(r.Body).Decode(&newStudent); err != nil {
		http.Error(w, "Invalid input data", http.StatusBadRequest)
		return
	}
	newStudent.ID = getNextID()
	mu.Lock()
	students[newStudent.ID] = newStudent
	mu.Unlock()

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newStudent)
}

func getAllStudents(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	studentList := make([]Student, 0, len(students))
	for _, student := range students {
		studentList = append(studentList, student)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(studentList)
}

func getStudentByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "Invalid student ID", http.StatusBadRequest)
		return
	}

	mu.Lock()
	student, exists := students[id]
	mu.Unlock()

	if !exists {
		http.Error(w, "Student not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(student)
}

func updateStudentByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "Invalid student ID", http.StatusBadRequest)
		return
	}

	var updatedStudent Student
	if err := json.NewDecoder(r.Body).Decode(&updatedStudent); err != nil {
		http.Error(w, "Invalid input data", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()
	student, exists := students[id]
	if !exists {
		http.Error(w, "Student not found", http.StatusNotFound)
		return
	}

	updatedStudent.ID = student.ID
	students[id] = updatedStudent

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedStudent)
}

func deleteStudentByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "Invalid student ID", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()
	if _, exists := students[id]; !exists {
		http.Error(w, "Student not found", http.StatusNotFound)
		return
	}

	delete(students, id)
	w.WriteHeader(http.StatusNoContent)
}

func generateSummary(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "Invalid student ID", http.StatusBadRequest)
		return
	}

	mu.Lock()
	student, exists := students[id]
	mu.Unlock()

	if !exists {
		http.Error(w, "Student not found", http.StatusNotFound)
		return
	}

	// Call the Ollama API to generate a summary
	ollamaSummary := fmt.Sprintf(
		"Student %s, aged %d, is enrolled in the course %s and can be contacted at %s.",
		student.Name, student.Age, student.Course, student.Email,
	)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(ollamaSummary))
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/students", createStudent).Methods("POST")
	r.HandleFunc("/students", getAllStudents).Methods("GET")
	r.HandleFunc("/students/{id:[0-9]+}", getStudentByID).Methods("GET")
	r.HandleFunc("/students/{id:[0-9]+}", updateStudentByID).Methods("PUT")
	r.HandleFunc("/students/{id:[0-9]+}", deleteStudentByID).Methods("DELETE")
	r.HandleFunc("/students/{id:[0-9]+}/summary", generateSummary).Methods("GET")
	fmt.Println("Server is running on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", r))
}
