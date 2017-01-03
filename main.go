package main

import (
    "fmt"
    "github.com/julienschmidt/httprouter"
    "net/http"
    "log"
    "encoding/json"
    "io/ioutil"
    "io"
    "time"
)

var taskSet = TaskSet{Task{"Test task", 0, 10, 100, 1000, 0, "#dddddd"}}
var settings = Settings{100, "Non-Preemptive"}

func GetSettings(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(settings)
}

func SetSettings(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
    var s Settings
    body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
    if err != nil {
        panic(err)
    }

    if err := r.Body.Close(); err != nil {
        panic(err)
    }

    if err := json.Unmarshal(body, &s); err != nil {
        w.Header().Set("Content-Type", "application/json; charset=UTF-8")
        w.WriteHeader(422) // unprocessable entity
        if err := json.NewEncoder(w).Encode(err); err != nil {
            panic(err)
        }
    }

    old := settings.Type
    settings = s

    switch s.Type {

    default:
        settings.Type = old
        break
    case "Non-Preemptive":
    case "Preemtive":
    case "Threshold":
        break
    }

    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    w.WriteHeader(http.StatusCreated)
    if err := json.NewEncoder(w).Encode(settings); err != nil {
        panic(err)
    }
}

func GetTaskset(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(taskSet)
}

func SetTaskset(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

    var ts TaskSet
    body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
    if err != nil {
        panic(err)
    }

    if err := r.Body.Close(); err != nil {
        panic(err)
    }

    if err := json.Unmarshal(body, &ts); err != nil {
        w.Header().Set("Content-Type", "application/json; charset=UTF-8")
        w.WriteHeader(422) // unprocessable entity
        if err := json.NewEncoder(w).Encode(err); err != nil {
            panic(err)
        }
    }

    taskSet = ts

    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    w.WriteHeader(http.StatusCreated)
    if err := json.NewEncoder(w).Encode(ts); err != nil {
        panic(err)
    }
}

func Ping(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
    fmt.Fprintf(w, "Pong %v", time.Now().UnixNano())
}

func main() {
    router := httprouter.New()

    // GET for Ping, should return "pong" and the server time in microseconds
    router.GET("/ping", Ping)

    // Taskset cruds, only retrieve and set allowed
    router.GET("/taskset", GetTaskset)
    router.PUT("/taskset", SetTaskset)
    router.POST("/taskset", SetTaskset)

    // Setting cruds, only retrieve and set allowed
    router.GET("/settings", GetSettings)
    router.PUT("/settings", SetSettings)
    router.POST("/settings", SetSettings)

    fmt.Println("Server about to start running")
    log.Fatal(http.ListenAndServe("192.168.178.12:1337", router))
}
