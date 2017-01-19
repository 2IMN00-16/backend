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
    Tasks     []Task
    Name      string
}

type Settings struct {
    Speed int
    Type  string
}

type LampAction struct {
    Lamp  string
    Color string
}

type Visualize struct {
    Scheduler string

}
