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
    Scheduler  string
    Lights     []LampDef
}

type LampDef struct {
    Name  string
    Value string
}
