package main

type Visualizer struct {
    taskSet TaskSet
    visuals []Visual
}

func (v *Visualizer) Init() (chan TaskSet, chan bool) {
    c := make(chan TaskSet)
    i := make(chan bool)

    go func() {
        for {
            select {
            case newValue := <-c:

                var signalReset = &v.taskSet == nil

                //if c.taskSet == nil {
                //    signalReset = true
                //}

                //signalReset = (v.taskSet == nil)
                v.taskSet = newValue

                for _, vs := range v.visuals {
                    //vs.Visualize(newValue)

                    if signalReset {
                        vs.Restart()
                    }
                }
            }
        }
    }()

    go func() {
        for {
            select {
            case <-i:
                for _, vs := range v.visuals {
                    vs.Restart()
                }
            }
        }
    }()

    return c, i
}

func (v *Visualizer) Visual(vs Visual) {
    v.visuals = append(v.visuals, vs)
}

type Visual interface {
    Restart()
    EventChannel (chan Event)
}
