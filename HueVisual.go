package main

import "fmt"

type HueVisual struct {

}

func (v *HueVisual) Restart() {
    fmt.Println("Restart HueVisual")
}

func (v *HueVisual) Visualize(ts TaskSet) {
}
