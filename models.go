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

type TaskSet []Task

type Settings struct {
    Speed int
    Type  string
}
