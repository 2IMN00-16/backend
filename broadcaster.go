package main

type Broadcaster struct {
    channels  [](chan Event)
    broadcast (chan Event)
    newChannel (chan (chan Event))
}

func (b *Broadcaster) GetBroadcast() (chan Event, (chan (chan Event))) {
    if b.broadcast == nil {
        b.newChannel = make(chan (chan Event))
        b.broadcast = make(chan Event)

        go func(){
            for {
                select {
                case v := <- b.broadcast:
                    for _, c := range b.channels {
                        c <- v
                    }

                case c := <- b.newChannel:
                    b.channels = append(b.channels, c)
                }
            }

        }()
    }

    return b.broadcast, b.newChannel
}
