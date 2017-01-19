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
    "log"
    "github.com/lucasb-eyer/go-colorful"
    "strconv"
    "regexp"
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

    // Make sure we always have a custom type set
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

func restart() {
    fmt.Println("Restarting visualizers")
    ResetsC <- true
    fmt.Println("Restarted visualizers")
}

func ForceRestart(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
    fmt.Fprint(w, "Will do!")
    restart()
}

func LampAmount(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {



    c, err := coap.Dial("t", "seminar.tuupke.nl:5683")
    if err != nil {
        w.WriteHeader(418)
        return
    }

    req := coap.Message{
        Type:      coap.Confirmable,
        Code:      coap.GET,
        MessageID: <-COaPId,
        //Payload:   []byte(""),
    }

    req.SetOption(coap.ETag, "weetag")
    req.SetOption(coap.MaxAge, 3)
    req.SetPathString(".well-known/core")

    rv, err := c.Send(req)
    if err != nil {
        log.Fatalf("Error sending request: %v", err)
    }

    var message string = ""
    for err == nil {
        if rv != nil {
            if err != nil {
                log.Fatalf("Error receiving: %v", err)
            }
            message = message + string(rv.Payload)
        }

        rv, err = c.Receive()

    }

    message = strings.Replace(message, ">", "\"", -1)
    message = strings.Replace(message, "</", "\"", -1)

    message = "[" + message + "]"

    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    fmt.Fprint(w, message)
    log.Printf("Response payload: %s", message)
}

func LampsIdentify(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

    c, err := coap.Dial("udp", "seminar.tuupke.nl:5683")
    if err != nil {
        log.Fatalf("Error dialing: %v", err)
    }

    req := coap.Message{
        Type:      coap.Confirmable,
        Code:      coap.GET,
        MessageID: <-COaPId,
        //Payload:   []byte(""),
    }

    req.SetOption(coap.ETag, "weetag")
    req.SetOption(coap.MaxAge, 3)
    req.SetPathString(".well-known/core")

    rv, err := c.Send(req)
    if err != nil {
        log.Fatalf("Error sending request: %v", err)
    }

    var message string = ""
    for err == nil {
        if rv != nil {
            if err != nil {
                log.Fatalf("Error receiving: %v", err)
            }
            message = message + string(rv.Payload)
        }

        rv, err = c.Receive()
    }

    re := regexp.MustCompile("lamp[0-9]+")

    matches := re.FindAllString(message, -1)
    if matches == nil{
        fmt.Fprint(w, "Something went wrong")
    }

    pal1, err1 := colorful.WarmPalette(len(matches))
    if err1 != nil{
        fmt.Fprint(w, "Something went wrong")
    }


    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    fmt.Fprint(w, "{")

    for i := 0; i < len(matches); i++ {
        color := pal1[i].Hex()
        fmt.Fprintf(w, "\"%s\" : \"%s\"", matches[i], color)

        if i < len(matches) - 1 {
            fmt.Fprint(w, ",")
        }

        lampC <- LampAction{
            matches[i],
            color,
        }
    }

    fmt.Fprint(w, "}")

    duration, _ := strconv.Atoi(p.ByName("dur"))


    timer2 := time.NewTimer(time.Duration(duration) * time.Second)

    go func() {
        <-timer2.C
        for i := 0; i < len(matches); i++ {
            color := pal1[i].Hex()
            fmt.Fprintf(w, "\"%s\" : \"%s\"", matches[i], color)

            if i < len(matches) - 1 {
                fmt.Fprint(w, ",")
            }

            lampC <- LampAction{
                matches[i],
                "#000000",
            }
        }
    }()

}

func Schedulers(w http.ResponseWriter, r *http.Request, _ httprouter.Params){
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    fmt.Fprintln(w, "[\"FPPS\",\"FPTS\",\"FPNS\"]")
}
func SetSchedulers(w http.ResponseWriter, r *http.Request, _ httprouter.Params){
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    fmt.Fprintln(w, "[\"FPPS\",\"FPTS\",\"FPNS\"]")
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
    lChan := make(chan LampAction)

    go func() {
        co, err := coap.Dial("udp", "seminar.tuupke.nl:5683")
        if err != nil {
            return
        }

        for {
            select {
            case a := <-lChan:

                if a.Color == "#000000" {
                    co.Send(buildMessage(a.Lamp, "on=False"))
                } else {
                    c, err := colorful.Hex(a.Color)
                    if err != nil {
                        log.Fatal(err)
                    }

                    h, s, v := c.Hsv()

                    co.Send(buildMessage(a.Lamp, "on=True"))
                    co.Send(buildMessage(a.Lamp, "hue="+strconv.Itoa(int(h / 360 * 65535))))
                    co.Send(buildMessage(a.Lamp, "sat="+strconv.Itoa(int(s * 254))))
                    co.Send(buildMessage(a.Lamp, "bri="+strconv.Itoa(int(1 + 243 * v))))
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
        Type:      coap.Confirmable,
        Code:      coap.PUT,
        MessageID: <-COaPId,
        Payload:   []byte(command),
    }

    req.SetOption(coap.ETag, "weetag")
    req.SetOption(coap.MaxAge, 3)
    req.SetPathString(lamp)

    return req
}
