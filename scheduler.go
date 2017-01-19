package main

type Scheduler struct {
    EventChan (chan Event)
}

func (s *Scheduler) startSchedule(){



    go func(){
        for{
            select {
            case <-ResetsC:
                // Reset

                break;

            //case
            }
        }
    }()
}



type Event struct {
    taskName string
}
