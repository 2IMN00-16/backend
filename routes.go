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
    "github.com/dustin/go-coap"
    "github.com/lucasb-eyer/go-colorful"
    "strconv"
    "sort"
    "os/exec"
    "sync"
)

func GetSettings(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(settings)
}

func SetSettings(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
    var s Visualize
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

    //old := settings.Scheduler
    settings = s

    rs := CyclerSettings{
        settings.CycleRate,
        []string{},
    }

    ps := CyclerSettings{
        settings.CycleRate,
        []string{},
    }

    as := CyclerSettings{
        settings.CycleRate,
        []string{},
    }
    //
    for _, q := range settings.Lights {
        switch q.Value {
        // \"off\", \"Running\",\"Active\",\"Preempted\
        case "Running":
            rs.Lamps = append(rs.Lamps, q.Name)
            break
        case "Active":
            as.Lamps = append(as.Lamps, q.Name)
            break
        case "Preempted":
            ps.Lamps = append(ps.Lamps, q.Name)
            break
        }
    }

    activeSettingsChan <- as
    preemptedSettingsChan <- ps
    runningSettinsChan <- rs

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

func GetVisualSettings(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
    w.Header().Set("Content-Type", "application/json")
    fmt.Fprintln(w, "[\"off\", \"Running\",\"Active\",\"Preempted\"]")
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

    jobColorMap = make(map[string]string)

    for _, v := range ts.Tasks {
        jobColorMap[v.Name] = v.Color
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

func restart() {
    fmt.Println("Restarting visualizers")

    // Set all stuff for scheduler

    var parsed []Task

    fmt.Println("Computing with scheduler: " + settings.Scheduler)

    for _, v := range taskSet.Tasks {
        local := v

        switch settings.Scheduler {
        case "Non-Preemptive":
            local.Threshold = 10000
            break
        case "Preemtive":
            local.Threshold = local.Priority
            break
        case "Threshold":
            break
        default:
            panic("Unkown scheduler " + settings.Scheduler)
        }

        parsed = append(parsed, local)
    }

    fmt.Printf("Parsed %+v\n", parsed)

    val, _ := json.Marshal(parsed)

    // Invoke renderer (Python Script)
    taskSetSetter <- runScheduler(ScheduleTransimitter{
        parsed,
        settings.Duration,
    })

    fmt.Println(string(val))

    fmt.Println("Restarted visualizers")
}

func initCycler(s CyclerSettings) (chan VisEvent, chan VisEvent, chan bool, chan CyclerSettings) {

    receiver := make(chan VisEvent)
    deleter := make(chan VisEvent)
    updater := make(chan CyclerSettings)
    reseter := make(chan bool)

    go func(s CyclerSettings, r chan VisEvent, d chan VisEvent, re chan bool, u chan CyclerSettings) {
        l := map[string]VisEvent{}

        timer := time.NewTimer(time.Millisecond * 0)
        off := 0
        hasColor := true
        for {
            select {
            case n := <-r:

                l[n.Job] = n

                if len(l) == 1 {
                    timer.Reset(time.Millisecond * 0)
                }

                break
            case de := <-d:

                // Delete when it is the active job
                delete(l, de.Job)

                timer.Reset(time.Millisecond * 0)

                break
            case <-re:
                l = map[string]VisEvent{}

                timer.Reset(time.Millisecond * 0)
                break
            case ss := <-u:
                s = ss

                fmt.Printf("Lamps: %v\n", s.Lamps)

                timer.Reset(time.Millisecond * time.Duration(s.Duration))
                break

            case <-timer.C:
                timer.Reset(time.Millisecond * time.Duration(s.Duration))

                le := len(l)

                if le == 0 {

                    if !hasColor {
                        continue
                    }

                    hasColor = false

                    for _, la := range s.Lamps {
                        go func() {
                            lampC <- LampAction{
                                la,
                                "#000000",
                            }
                        }()
                    }

                    break
                }

                hasColor = true

                n := len(s.Lamps)

                atJob := 0

                for _, v := range l {

                    lid := -1

                    if atJob - off >= 0 && atJob - off < n {
                        lid = atJob - off
                    } else if off + n > le && atJob < (off + n) % le {
                        lid = atJob - off + le
                    }

                    if lid >= 0 {
                        //jobColorMapMutex.Lock()
                        color := jobColorMap[v.Task]
                        //jobColorMapMutex.Unlock()
                        go func(li int, c string) {
                            lampC <- LampAction{
                                s.Lamps[li],
                                c,
                            }
                        }(lid, color)
                    }

                    atJob++
                }

                if le > 0 {
                    off = (off + 1) % le
                } else {
                    off = 0
                }

                break
            }
        }
    }(s, receiver, deleter, reseter, updater)

    return receiver, deleter, reseter, updater
}

var preemptedSettingsChan chan CyclerSettings
var activeSettingsChan chan CyclerSettings
var runningSettinsChan chan CyclerSettings
var taskSetSetter chan VisEvents

func showLamps() chan CyclerSettings {

    taskSetSetter = make(chan VisEvents)

    timer := time.NewTimer(time.Millisecond * 0)

    var t VisEvents

    addPreemted, removePreemted, resetp, ac := initCycler(cs)
    addActive, removeActive, reseta, bc := initCycler(cs)

    preemptedSettingsChan = ac
    activeSettingsChan = bc

    localSettings := cs
    receiver := make(chan CyclerSettings)

    go func() {
        var running *VisEvent
        for {
            select {
            case ss := <-receiver:
                localSettings = ss
                break
            case a := <-taskSetSetter:
                t = a

                taskDeadlineMap := make(map[string]int)
                for _, v := range taskSet.Tasks {
                    taskDeadlineMap[v.Name] = v.Deadline
                }

                // Just provide a grace period

                timer.Reset(time.Second * 1)

                sort.Sort(t)

                temp := []VisEvent{}

                jobDeadlineMap := make(map[string]int)

                resetp <- true
                reseta <- true

                for _, v := range a {

                    if v.Event == "jobArrived" {
                        jobDeadlineMap[v.Job] = taskDeadlineMap[v.Task] + v.Time
                    }

                    if v.Event == "jobCompleted" {
                        j := jobDeadlineMap[v.Job]

                        if v.Time > j {
                            temp = append(temp, VisEvent{
                                j,
                                "deadlineMissed",
                                v.Task,
                                v.Job,
                            })

                            fmt.Printf("Deadline miss for %v\n", v.Task)
                        }
                    }
                }

                for _, v := range settings.Lights {
                    lampC <- LampAction{
                        v.Name,
                        "#000000",
                    }
                }


                t = append(t, temp...)

                sort.Sort(t)

                time.Sleep(time.Second * 2)
                timer.Reset(time.Second)

                break;
            case <-timer.C:
                fmt.Println("Vis thingy")

                if len(t) == 0 {

                    fmt.Println("No penging events")
                    timer.Reset(time.Second * 3)
                    break
                }

                co := true
                var lastTime int

                // All stuff to do in this timestep
                for co {

                    a := len(t)

                    // Shift element from slice
                    e := t[0]
                    t = t[1:]
                    b := len(t)

                    if a == b {
                        fmt.Println("damn")
                        panic("ar")
                    }
                    lastTime = e.Time

                    fmt.Printf("Going to handle %+v %v, %v\n", e, a, b)

                    switch e.Event {
                    case "jobArrived":
                        go func() {
                            addActive <- e
                        }()

                        break
                    case "jobResumed":

                        go func() {
                            removeActive <- e
                        }()
                        go func() {
                            removePreemted <- e
                        }()

                        running = &e

                        break
                    case "jobCompleted":
                        go func() {
                            removeActive <- e
                            removePreemted <- e
                        }()

                        running = nil

                        break
                    case "jobPreempted":
                        go func() {
                            addPreemted <- e
                        }()
                        break

                    case "deadlineMissed":

                        fmt.Println("DEADLINE!!!!!!!!!!")
                        fmt.Println("Setting color")
                        jobColorMapMutex.Lock()
                        fmt.Printf("Running: %v\n", running)
                        fmt.Println(jobColorMap)
                        color := jobColorMap[e.Task]
                        jobColorMapMutex.Unlock()

                        for _, h := range localSettings.Lamps {
                            fmt.Println("Goind go set lamp " + h)
                            go func(c string, l string) {
                                lampC <- LampAction{
                                    l,
                                    c,
                                }

                            }("#ffffff", h)
                        }

                        time.Sleep(time.Second * 1)


                        for _, h := range localSettings.Lamps {
                            fmt.Println("Goind go set lamp " + h)
                            go func(c string, l string) {
                                lampC <- LampAction{
                                    l,
                                    c,
                                }

                            }(color, h)
                        }

                        time.Sleep(time.Second * 2)

                        for _, h := range localSettings.Lamps {
                            fmt.Println("Goind go set lamp " + h)
                            go func(c string, l string) {
                                lampC <- LampAction{
                                    l,
                                    c,
                                }

                            }("#ffffff", h)
                        }

                        time.Sleep(time.Second * 1)

                        for _, h := range localSettings.Lamps {
                            fmt.Println("Goind go set lamp " + h)
                            go func(c string, l string) {
                                lampC <- LampAction{
                                    l,
                                    c,
                                }

                            }(color, h)
                        }

                        time.Sleep(time.Second * 2)

                        for _, h := range localSettings.Lamps {
                            fmt.Println("Goind go set lamp " + h)
                            go func(c string, l string) {
                                lampC <- LampAction{
                                    l,
                                    c,
                                }

                            }("#ffffff", h)
                        }

                        time.Sleep(time.Second * 1)

                        break
                    }

                    var color string

                    fmt.Printf("Value for running: %v", running)

                    if running == nil {
                        color = "#000000"
                    } else {

                        fmt.Println("Setting color")
                        jobColorMapMutex.Lock()
                        fmt.Printf("Running: %v\n", running)
                        fmt.Println(jobColorMap)
                        color = jobColorMap[(*running).Task]
                        jobColorMapMutex.Unlock()
                    }

                    fmt.Println("Color " + color)

                    for _, h := range localSettings.Lamps {
                        fmt.Println("Goind go set lamp " + h)
                        go func(c string, l string) {
                            lampC <- LampAction{
                                l,
                                c,
                            }

                        }(color, h)
                    }

                    co = len(t) > 0 && t[0].Time == e.Time
                }

                // Decide to append an event

                if len(t) > 0 {
                    fmt.Printf("Waiting for %v ms\n", settings.TimeFactor*(t[0].Time-lastTime))
                    timer.Reset(time.Millisecond * time.Duration(settings.TimeFactor * (t[0].Time - lastTime)))
                } else {
                    fmt.Println("No more events")

                    timer.Reset(time.Second)
                }

                break
            }
        }
    }()

    return receiver
}

func ForceRestart(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
    fmt.Fprint(w, "Will do!")
    restart()
}

func LampAmount(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    fmt.Fprint(w, "[\"lamp1\",\"lamp2\",\"lamp3\"]")
    //log.Printf("Response payload: %s", message)
}

func LampsIdentify(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
    pal1, err1 := colorful.SoftPalette(3)
    if err1 != nil {
        fmt.Fprint(w, "Something went wrong")
    }

    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    fmt.Fprint(w, "{")

    matches := []string{
        "lamp1",
        "lamp2",
        "lamp3",
    }

    for i := 0; i < 3; i++ {
        color := pal1[i].Hex()
        fmt.Fprintf(w, "\"%s\" : \"%s\"", matches[i], color)

        if i < len(matches)-1 {
            fmt.Fprint(w, ",")
        }

        go func(m string) {
            lampC <- LampAction{
                m,
                color,
            }
        }(matches[i])
    }

    fmt.Fprint(w, "}")

    duration, _ := strconv.Atoi(p.ByName("dur"))

    timer2 := time.NewTimer(time.Duration(duration) * time.Millisecond)

    go func() {
        <-timer2.C

        for i := 0; i < len(matches); i++ {
            color := pal1[i].Hex()
            fmt.Fprintf(w, "\"%s\" : \"%s\"", matches[i], color)

            if i < len(matches)-1 {
                fmt.Fprint(w, ",")
            }
            fmt.Println("Hierrr")

            lampC <- LampAction{
                matches[i],
                "#000000",
            }
        }
    }()

}

func Schedulers(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    fmt.Fprintln(w, "[\"Non-Preemptive\",\"Preemtive\",\"Threshold\"]")
}

func SetSchedulers(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    fmt.Fprintln(w, "[\"Non-Preemptive\",\"Preemtive\",\"Threshold\"]")
}

func SetLamp(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

    var l LampAction;
    body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
    if err != nil {
        panic(err)
    }

    if err := r.Body.Close(); err != nil {
        panic(err)
    }

    if err := json.Unmarshal(body, &l); err != nil {
        w.Header().Set("Content-Type", "application/json; charset=UTF-8")
        w.WriteHeader(422) // unprocessable entity
        if err := json.NewEncoder(w).Encode(err); err != nil {
            panic(err)
        }
    }

    lampC <- l

    fmt.Fprint(w, "Sent")
}

func initLamp() (chan LampAction) {
    lChan := make(chan LampAction, 10000)

    connMap := make(map[string]*coap.Conn)

    for k, v := range lampMap {

        fmt.Printf("Trying to connect to %v for lamp %v\n", v, k)

        co, err := coap.Dial("udp", v+":5683")
        if err != nil {
            fmt.Println("Cannot connect to lamp: " + k);
        }

        connMap[k] = co
    }

    go func() {
        mapMutex := sync.Mutex{}

        for {
            select {
            case a := <-lChan:
                mapMutex.Lock()
                co := connMap[a.Lamp]
                mapMutex.Unlock()

                fmt.Println("Sending to " + a.Lamp)

                if a.Color == "#000000" {
                    co.Send(buildMessage(a.Lamp, "on=False"))
                } else {
                    c, err := colorful.Hex(a.Color)
                    if err != nil {
                        fmt.Println("ars")
                        //log.Fatal(err)
                        fmt.Println("ars")
                    }

                    h, s, v := c.Hsl()

                    var e error
                    //go func() {
                    //go func() {
                    _, e = co.Send(buildMessage(a.Lamp, "on=True"))
                    if e != nil {
                        fmt.Println("LampError")
                        fmt.Println(e)
                    }
                    //}()
                    //go func() {
                    _, e = co.Send(buildMessage(a.Lamp, "hue="+strconv.Itoa(int(h / 360 * 65535))))
                    if e != nil {
                        fmt.Println("LampError")
                        fmt.Println(e)
                    } //}()
                    //go func() {
                    _, e = co.Send(buildMessage(a.Lamp, "sat="+strconv.Itoa(int(s * 254))))
                    if e != nil {
                        fmt.Println("LampError")
                        fmt.Println(e)
                    } //}()
                    //go func() {
                    _, e = co.Send(buildMessage(a.Lamp, "bri="+strconv.Itoa(int(1 + 243*v))))
                    if e != nil {
                        fmt.Println("LampError")
                        fmt.Println(e)
                    } //}()
                    //}()
                }

                break;
            }
        }
    }()

    return lChan
}

func buildMessage(lamp string, command string) coap.Message {

    fmt.Println("Sending: " + command)
    req := coap.Message{
        Type:      coap.NonConfirmable,
        Code:      coap.PUT,
        MessageID: <-COaPId,
        Payload:   []byte(command),
    }

    req.SetOption(coap.ETag, "weetag")
    req.SetOption(coap.MaxAge, 3)
    req.SetPathString(lamp)

    return req
}

func runScheduler(set ScheduleTransimitter) VisEvents {

    js, err := json.Marshal(set)

    if err != nil {
        panic(err)
    }

    err = ioutil.WriteFile("input.json", js, 0644)

    if err != nil {
        panic(err)
    }

    // Call the scheduler
    cmd := exec.Command("/usr/local/bin/python", "scheduler.py")
    //cmd := exec.Command("/usr/bin/java", "-jar Scheduler.jar")
    err = cmd.Start()
    if err != nil {
        panic(err)
    }

    // Await finish
    err = cmd.Wait()

    // Read file
    vis, err := ioutil.ReadFile("schedule.json")

    if err != nil {
        panic(err)
    }

    var v struct {
        Schedule VisEvents
    }

    if err := json.Unmarshal(vis, &v); err != nil {
        panic(err)
    }

    return v.Schedule
}

func GetGraspScript(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
    vis, err := ioutil.ReadFile("schedule.grasp")

    if err != nil {
        w.WriteHeader(http.StatusExpectationFailed)
        return
    }

    w.Header().Add("Content-Type","text/plain")
    w.Header().Add("Content-Disposition","attachment; filename=\"schedule.grasp\"")

    fmt.Fprintf(w, "%s", vis)
}
