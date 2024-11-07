package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
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

	prompt := fmt.Sprintf(
		"Please generate a detailed summary of the student with the following information: Name: %s, Age: %d, Course: %s, Email: %s. Make sure the summary is clear and informative.",
		student.Name, student.Age, student.Course, student.Email,
	)

	ollamaAPIURL := "http://localhost:11434/api/generate" 
	ollamaModel := "llama3.2" 
	requestBody := fmt.Sprintf(`{
		"model": "%s",
		"prompt": "%s"
	}`, ollamaModel, prompt)

	client := &http.Client{}
	req, err := http.NewRequest("POST", ollamaAPIURL, strings.NewReader(requestBody))
	if err != nil {
		http.Error(w, "Error creating request", http.StatusInternalServerError)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Error calling Ollama API", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var apiResponse map[string]interface{}
	var fullResponse string

	for {
		if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
			http.Error(w, "Error parsing Ollama API response", http.StatusInternalServerError)
			return
		}
		log.Printf("Ollama API Partial Response: %+v", apiResponse)

		if response, ok := apiResponse["response"].(string); ok {
			fullResponse += response
		}

		if done, ok := apiResponse["done"].(bool); ok && done {
			break
		}
	}

	if fullResponse == "" {
		http.Error(w, "No response from Ollama API", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fullResponse))
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
