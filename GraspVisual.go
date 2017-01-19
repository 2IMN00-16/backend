package main

import "fmt"

type GraspVisual struct {
    runningSet   TaskSet
    newSet       TaskSet
    eventChannel (chan Event)
}

func (v *GraspVisual) EventChannel(e (chan Event)) {
    oldChannel := v.eventChannel
    v.eventChannel = e

    if oldChannel == nil {
        go func() {
            for {
                select {
                case  <-v.eventChannel:
                    // Write to file.
                    break
                }
            }
        }()
    }
}

func (v *GraspVisual) Restart() {
    v.runningSet = v.newSet
    fmt.Println("Restart GrashVisual")
}
