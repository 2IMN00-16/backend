package main

type Task struct {
    Name        string
    Priority    int
    Computation int
    Period      int
    Deadline    int
    Threshold   int
    Color       string
}

type TaskSet struct {
    Tasks []Task
    Name  string
}

type LampAction struct {
    Lamp  string
    Color string
}

type Visualize struct {
    CycleRate  int
    TimeFactor int
    Duration   int
    Scheduler  string
    Lights     []LampDef
}

type LampDef struct {
    Name  string
    Value string
}

type VisEvent struct {
    Time  int
    Event string
    Task  string
    Job   string
}

type VisEvents []VisEvent

func (slice VisEvents) Len() int {
    return len(slice)
}

func (slice VisEvents) Less(i, j int) bool {
    return slice[i].Time < slice[j].Time || (slice[i].Time == slice[j].Time && slice[i].Event < slice[j].Event)
}

func (slice VisEvents) Swap(i, j int) {
    slice[i], slice[j] = slice[j], slice[i]
}

type CyclerSettings struct {
    Duration int
    Lamps    []string
}

type ScheduleTransimitter struct {
    Tasks    []Task
    Duration int
}
