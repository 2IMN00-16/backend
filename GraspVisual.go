package main

import "fmt"

type GraspVisual struct {
    runningSet TaskSet
    newSet TaskSet
}

func (v *GraspVisual) Restart() {
    v.runningSet = v.newSet
    fmt.Println("Restart GrashVisual")
}

func (v *GraspVisual) Visualize(ts TaskSet) {
    v.newSet = ts
}
