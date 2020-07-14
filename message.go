package main

// Message is ...
type Message struct {
	room    string
	message []byte
}

func broadcast(h *Hub, ch chan *Message) { // process data from redis pub sub
	for {
		msg := <-ch

		// send it out to every client in the room that is currently connected
		h.mtx.Lock()
		for c := range h.rooms[msg.room] {
			/*
				panic: send on closed channel
			*/
			c.send <- msg.message
		}
		h.mtx.Unlock()
	}
}
