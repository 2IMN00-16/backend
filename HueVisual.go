package main

import "fmt"

type HueVisual struct {
    eventChannel (chan Event)
    eventBuffer []Event

    //timeStep
}

func (v *HueVisual) EventChannel (e (chan Event)) {
    oldChannel := v.eventChannel
    v.eventChannel = e

    if oldChannel == nil {
        go func() {
            for {
                select {
                case ev := <- v.eventChannel:
                    v.eventBuffer = append(v.eventBuffer, ev)
                    break

                    // Other stuff here, like signalling lamps etc
                }
            }
        }()
    }
}

func (v *HueVisual) Restart() {
    fmt.Println("Restart HueVisual")
}
