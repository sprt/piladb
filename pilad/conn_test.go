package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fern4lvarez/piladb/pila"
	"github.com/fern4lvarez/piladb/pkg/uuid"
)

func TestNewConn(t *testing.T) {
	conn := NewConn()

	if conn == nil {
		t.Fatal("conn is nil")
	}

	if conn.Pila == nil {
		t.Error("conn.Pila is nil")
	}

	if conn.Status == nil {
		t.Fatal("conn.Status is nil")
	}

	if conn.Status.Code != "OK" {
		t.Errorf("conn.Status is %s, expected %s", conn.Status.Code, "OK")
	}
}

func TestStatusHandler(t *testing.T) {
	conn := NewConn()
	request, err := http.NewRequest("GET", "/_status", nil)
	if err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()

	conn.statusHandler(response, request)

	if contentType := response.Header().Get("Content-Type"); contentType != "application/json" {
		t.Errorf("Content-Type is %v, expected %v", contentType, "application/json")
	}

	if response.Code != 200 {
		t.Errorf("response code is %v, expected %v", response.Code, 200)
	}

	statusJSON, err := ioutil.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.HasPrefix(string(statusJSON), `{"status":"OK","version"`) {
		t.Errorf("status is %s", string(statusJSON))
	}
}

func TestDatabasesHandler_GET(t *testing.T) {
	db := pila.NewDatabase("db")

	p := pila.NewPila()
	_ = p.AddDatabase(db)

	conn := NewConn()
	conn.Pila = p

	request, err := http.NewRequest("GET", "/databases", nil)
	if err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()

	conn.databasesHandler(response, request)

	if contentType := response.Header().Get("Content-Type"); contentType != "application/json" {
		t.Errorf("Content-Type is %v, expected %v", contentType, "application/json")
	}

	if response.Code != 200 {
		t.Errorf("response code is %v, expected %v", response.Code, 200)
	}

	databases, err := ioutil.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}

	if expected := `{"number_of_databases":1,"databases":[{"id":"8cfa8cb55c92fa403369a13fd12a8e01","name":"db","number_of_stacks":0}]}`; string(databases) != expected {
		t.Errorf("databases are %s, expected %s", string(databases), expected)
	}
}

func TestDatabasesHandler_GET_Empty(t *testing.T) {
	conn := NewConn()
	request, err := http.NewRequest("GET", "/databases", nil)
	if err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()

	conn.databasesHandler(response, request)

	if contentType := response.Header().Get("Content-Type"); contentType != "application/json" {
		t.Errorf("Content-Type is %v, expected %v", contentType, "application/json")
	}

	if response.Code != 200 {
		t.Errorf("response code is %v, expected %v", response.Code, 200)
	}

	databases, err := ioutil.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}

	if string(databases) != `{"number_of_databases":0,"databases":[]}` {
		t.Errorf("databases are %s, expected %s", string(databases), `{"number_of_databases":0,"databases":[]}`)
	}
}

func TestDatabasesHandler_PUT(t *testing.T) {
	// see TestCreateDatabaseHandler* tests for a further spec on
	// how createDatabaseHandler works.
	conn := NewConn()
	request, err := http.NewRequest("PUT", "/databases", nil)
	if err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()

	conn.databasesHandler(response, request)

	if response.Code != 400 {
		t.Errorf("response code is %v, expected %v", response.Code, 400)
	}
}

func TestCreateDatabaseHandler(t *testing.T) {
	conn := NewConn()
	request, err := http.NewRequest("PUT", "/databases?name=db", nil)
	if err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()

	conn.createDatabaseHandler(response, request)

	if contentType := response.Header().Get("Content-Type"); contentType != "application/json" {
		t.Errorf("Content-Type is %v, expected %v", contentType, "application/json")
	}

	if response.Code != 201 {
		t.Errorf("response code is %v, expected %v", response.Code, 201)
	}

	databases, err := ioutil.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}

	if string(databases) != `{"id":"8cfa8cb55c92fa403369a13fd12a8e01","name":"db","number_of_stacks":0}` {
		t.Errorf("databases are %s, expected %s", string(databases), `{"id":"8cfa8cb55c92fa403369a13fd12a8e01","name":"db","number_of_stacks":0}`)
	}
}

func TestCreateDatabaseHandler_NoName(t *testing.T) {
	conn := NewConn()
	request, err := http.NewRequest("PUT", "/databases", nil)
	if err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()

	conn.createDatabaseHandler(response, request)

	if response.Code != 400 {
		t.Errorf("response code is %v, expected %v", response.Code, 400)
	}
}

func TestCreateDatabaseHandler_Duplicated(t *testing.T) {
	conn := NewConn()

	request, err := http.NewRequest("PUT", "/databases?name=db", nil)
	if err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()

	conn.createDatabaseHandler(response, request)

	if response.Code != 201 {
		t.Errorf("response code is %v, expected %v", response.Code, 201)
	}

	request, err = http.NewRequest("PUT", "/databases?name=db", nil)
	if err != nil {
		t.Fatal(err)
	}
	response = httptest.NewRecorder()

	conn.createDatabaseHandler(response, request)

	if response.Code != 409 {
		t.Errorf("response code is %v, expected %v", response.Code, 409)
	}
}

func TestDatabaseHandler_GET(t *testing.T) {
	s := pila.NewStack("stack")
	s.Push("foo")

	db := pila.NewDatabase("mydb")
	_ = db.AddStack(s)

	p := pila.NewPila()
	_ = p.AddDatabase(db)

	conn := NewConn()
	conn.Pila = p

	request, err := http.NewRequest("GET",
		fmt.Sprintf("/databases/%s",
			db.ID.String()),
		nil)
	if err != nil {
		t.Fatal(err)
	}

	response := httptest.NewRecorder()

	databaseHandle := conn.databaseHandler(db.ID.String())
	databaseHandle.ServeHTTP(response, request)

	if contentType := response.Header().Get("Content-Type"); contentType != "application/json" {
		t.Errorf("Content-Type is %v, expected %v", contentType, "application/json")
	}

	if response.Code != 200 {
		t.Errorf("response code is %v, expected %v", response.Code, 200)
	}

	database, err := ioutil.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}

	if expected := `{"id":"c13cec0e70876381c78c616ee2d809eb","name":"mydb","number_of_stacks":1,"stacks":["240d019e07a0ec2d7eeca6c7c00b82a4"]}`; string(database) != expected {
		t.Errorf("database is %v, expected %v", string(database), expected)
	}
}

func TestDatabaseHandler_GET_Name(t *testing.T) {
	s := pila.NewStack("stack")
	s.Push("foo")

	db := pila.NewDatabase("mydb")
	_ = db.AddStack(s)

	p := pila.NewPila()
	_ = p.AddDatabase(db)

	conn := NewConn()
	conn.Pila = p

	request, err := http.NewRequest("GET",
		fmt.Sprintf("/databases/%s",
			db.Name),
		nil)
	if err != nil {
		t.Fatal(err)
	}

	response := httptest.NewRecorder()

	databaseHandle := conn.databaseHandler(db.Name)
	databaseHandle.ServeHTTP(response, request)

	if contentType := response.Header().Get("Content-Type"); contentType != "application/json" {
		t.Errorf("Content-Type is %v, expected %v", contentType, "application/json")
	}

	if response.Code != 200 {
		t.Errorf("response code is %v, expected %v", response.Code, 200)
	}

	database, err := ioutil.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}

	if expected := `{"id":"c13cec0e70876381c78c616ee2d809eb","name":"mydb","number_of_stacks":1,"stacks":["240d019e07a0ec2d7eeca6c7c00b82a4"]}`; string(database) != expected {
		t.Errorf("database is %v, expected %v", string(database), expected)
	}
}

func TestDatabaseHandler_GET_Gone(t *testing.T) {
	conn := NewConn()

	request, err := http.NewRequest("GET",
		fmt.Sprintf("/databases/%s",
			"nodb"),
		nil)
	if err != nil {
		t.Fatal(err)
	}

	response := httptest.NewRecorder()

	databaseHandle := conn.databaseHandler(uuid.UUID("nodb").String())
	databaseHandle.ServeHTTP(response, request)

	if response.Code != 410 {
		t.Errorf("response code is %v, expected %v", response.Code, 410)
	}

	database, err := ioutil.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}

	if expected := "database nodb is Gone"; string(database) != expected {
		t.Errorf("database is %v, expected %v", string(database), expected)
	}

}

func TestNotFoundHandler_WrongEndpoint(t *testing.T) {
	conn := NewConn()
	request, err := http.NewRequest("GET", "/_statuss", nil)
	if err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()

	conn.notFoundHandler(response, request)

	if response.Code != 404 {
		t.Errorf("response code is %v, expected %v", response.Code, 404)
	}
}

func TestNotFoundHandler_WrongType(t *testing.T) {
	conn := NewConn()
	request, err := http.NewRequest("POST", "/_status", nil)
	if err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()

	conn.notFoundHandler(response, request)

	if response.Code != 404 {
		t.Errorf("response code is %v, expected %v", response.Code, 404)
	}
}

func TestPopStackHandler(t *testing.T) {
	s := pila.NewStack("stack")
	s.Push("foo")

	db := pila.NewDatabase("db")
	_ = db.AddStack(s)

	p := pila.NewPila()
	_ = p.AddDatabase(db)

	conn := NewConn()
	conn.Pila = p

	request, err := http.NewRequest("GET",
		fmt.Sprintf("/databases/%s/stacks/%s",
			db.ID.String(),
			s.ID.String()),
		nil)
	if err != nil {
		t.Fatal(err)
	}

	response := httptest.NewRecorder()

	params := map[string]string{
		"database_id": db.ID.String(),
		"stack_id":    s.ID.String(),
	}

	popStackHandle := conn.popStackHandler(params)
	popStackHandle.ServeHTTP(response, request)

	if contentType := response.Header().Get("Content-Type"); contentType != "application/json" {
		t.Errorf("Content-Type is %v, expected %v", contentType, "application/json")
	}

	if response.Code != 200 {
		t.Errorf("response code is %v, expected %v", response.Code, 200)
	}

	elementJSON, err := ioutil.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}

	if string(elementJSON) != "\"foo\"" {
		t.Errorf("popped element is %v, expected %v", string(elementJSON), "\"foo\"")
	}
}

func TestPopStackHandler_EmptyStack(t *testing.T) {
	s := pila.NewStack("stack")

	db := pila.NewDatabase("db")
	_ = db.AddStack(s)

	p := pila.NewPila()
	_ = p.AddDatabase(db)

	conn := NewConn()
	conn.Pila = p

	request, err := http.NewRequest("GET",
		fmt.Sprintf("/databases/%s/stacks/%s",
			db.ID.String(),
			s.ID.String()),
		nil)
	if err != nil {
		t.Fatal(err)
	}

	response := httptest.NewRecorder()

	params := map[string]string{
		"database_id": db.ID.String(),
		"stack_id":    s.ID.String(),
	}

	popStackHandle := conn.popStackHandler(params)
	popStackHandle.ServeHTTP(response, request)

	if response.Code != 204 {
		t.Errorf("response code is %v, expected %v", response.Code, 204)
	}
}

func TestPopStackHandler_NoStackFound(t *testing.T) {
	s := pila.NewStack("stack")

	db := pila.NewDatabase("db")

	p := pila.NewPila()
	_ = p.AddDatabase(db)

	conn := NewConn()
	conn.Pila = p

	request, err := http.NewRequest("GET",
		fmt.Sprintf("/databases/%s/stacks/%s",
			db.ID.String(),
			s.ID.String()),
		nil)
	if err != nil {
		t.Fatal(err)
	}

	response := httptest.NewRecorder()

	params := map[string]string{
		"database_id": db.ID.String(),
		"stack_id":    s.ID.String(),
	}

	popStackHandle := conn.popStackHandler(params)
	popStackHandle.ServeHTTP(response, request)

	if response.Code != 410 {
		t.Errorf("response code is %v, expected %v", response.Code, 410)
	}

	message, err := ioutil.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}

	if string(message) != fmt.Sprintf("stack %s is Gone", s.ID.String()) {
		t.Errorf("message is %s, expected %s",
			string(message),
			fmt.Sprintf("stack %s is Gone", s.ID.String()))
	}
}

func TestPopStackHandler_NoDatabaseFound(t *testing.T) {
	s := pila.NewStack("stack")
	db := pila.NewDatabase("db")

	p := pila.NewPila()

	conn := NewConn()
	conn.Pila = p

	request, err := http.NewRequest("GET",
		fmt.Sprintf("/databases/%s/stacks/%s",
			db.ID.String(),
			s.ID.String()),
		nil)
	if err != nil {
		t.Fatal(err)
	}

	response := httptest.NewRecorder()

	params := map[string]string{
		"database_id": db.ID.String(),
		"stack_id":    s.ID.String(),
	}

	popStackHandle := conn.popStackHandler(params)
	popStackHandle.ServeHTTP(response, request)

	if response.Code != 410 {
		t.Errorf("response code is %v, expected %v", response.Code, 410)
	}

	message, err := ioutil.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}

	if string(message) != fmt.Sprintf("database %s is Gone", db.ID.String()) {
		t.Errorf("message is %s, expected %s",
			string(message),
			fmt.Sprintf("database %s is Gone", db.ID.String()))
	}
}
