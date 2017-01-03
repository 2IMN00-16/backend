package main

import (
    "encoding/json"
    "net/http"
    "github.com/julienschmidt/httprouter"
    "io/ioutil"
    "io"
    "fmt"
    "time"
    "strings"
)

func GetSettings(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
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

    if strings.HasSuffix(r.URL.String(), "restart") {
        restart()
    }
}

func GetTaskset(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
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

    if strings.HasSuffix(r.URL.String(), "restart") {
        restart()
    }
}

func Ping(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
    fmt.Fprintf(w, "Pong %v", time.Now().UnixNano())
}

func restart(){
    fmt.Println("Restarting visualizers")
    ResetsC <- true
    fmt.Println("Restarted visualizers")
}

func ForceRestart(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
    fmt.Fprint(w, "Will do!")
    restart()
}
