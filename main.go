package main

import (
    "fmt"
    "github.com/julienschmidt/httprouter"
    "net/http"
    "log"
)

var taskSet = TaskSet{Task{"Test task", 0, 10, 100, 1000, 0, "#dddddd"}}
var settings = Settings{100, "Non-Preemptive"}

var TaskSetC (chan TaskSet)
var ResetsC (chan bool)

func main() {
    // Seed the visualizers
    var vs = Visualizer{}

    hv := HueVisual{}
    gv := GraspVisual{}

    vs.Visual(&hv)
    vs.Visual(&gv)
    TaskSetC, ResetsC = vs.Init()

    TaskSetC <- taskSet

    router := httprouter.New()

    // GET for Ping, should return "pong" and the server time in microseconds
    router.GET("/ping", Ping)

    // Taskset cruds, only retrieve and set allowed
    router.GET("/taskset", GetTaskset)
    router.PUT("/taskset", SetTaskset)
    router.PUT("/taskset/restart", SetTaskset)

    // Setting cruds, only retrieve and set allowed
    router.GET("/settings", GetSettings)
    router.PUT("/settings", SetSettings)
    router.PUT("/settings/restart", SetSettings)

    router.PATCH("/restart", ForceRestart)

    fmt.Println("Server about to start running")
    log.Fatal(http.ListenAndServe(":1337", router))
}
