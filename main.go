package main

import (
    "fmt"
    "github.com/julienschmidt/httprouter"
    "net/http"
    "log"
    "sync"
)

var taskSet TaskSet

var settings Visualize
var cs CyclerSettings

var TaskSetC (chan TaskSet)
var ResetsC (chan bool)
var COaPId (chan uint16)

var jobColorMap = map[string] string{}
var jobColorMapMutex = sync.Mutex{}

var lampMap = map[string] string{
    "lamp1" : "10.1.0.20",
    "lamp2" : "10.1.0.21",
    "lamp3" : "10.1.0.22",
}

func seedValues(){
    var t = Task{"Test task", 10, 10, 100, 1000, 0, "#013370"}
    var ts = make([]Task, 1)
    ts[0] = t

    taskSet = TaskSet{ts, "Default"}

    settings = Visualize{1000, 50, 30000, "Preemtive", []LampDef{
        {"lamp1", "free"},
        {"lamp2", "free"},
        {"lamp3", "free"},
    }}

    cs = CyclerSettings{
        50,
        []string{},
    }

    fmt.Println(settings)

}

var lampC (chan LampAction)

func main() {

    seedValues()
    lampC = initLamp()

    runningSettinsChan = showLamps()

    runningSettinsChan <- CyclerSettings{
        50,
        []string{"lamp1"},
    }

    // Seed the visualizers
    //var vs = Visualizer{}
    //
    //vs.Visual(&HueVisual{})
    //vs.Visual(&GraspVisual{})
    //
    //TaskSetC, ResetsC = vs.Init()
    //
    //TaskSetC <- taskSet

    // COaP identifier generation
    COaPId = make(chan uint16)

    go func() {
        var messageId uint16 = 1
        for {
            COaPId <- messageId
            messageId++
        }
    }()

    br := Broadcaster{}

    br.GetBroadcast()

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
    router.GET("/settings/visualizers", GetVisualSettings)
    router.PUT("/settings/restart", SetSettings)

    // Set a lamp to a value
    router.POST("/command", SetLamp)

    // Restart the visualisation
    router.PATCH("/restart", ForceRestart)

    // Get the running lamp amount
    router.GET("/lamps", LampAmount)

    // Get the running lamp amount
    router.POST("/lamps/identify/:dur", LampsIdentify)

    router.GET("/grasp", GetGraspScript)

    router.GET("/schedulers", Schedulers)
    //router.POST("/schedulers", SetSchedulers)

    fmt.Println("Server about to start running")
    log.Fatal(http.ListenAndServe(":1337", router))
}
