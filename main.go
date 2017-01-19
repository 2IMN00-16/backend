package main

import (
    "fmt"
    "github.com/julienschmidt/httprouter"
    "net/http"
    "log"
)

var taskSet = TaskSet{Task{"Test task", 0, 10, 100, 1000, 0, "#dddddd"}}
var settings = Settings{100, "Non-Preemptive"}

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

    fmt.Println("Server about to start runnnning")
    log.Fatal(http.ListenAndServe("192.168.178.12:1337", router))
}
