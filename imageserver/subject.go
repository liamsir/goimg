package imageserver

import "fmt"

type Listener struct {
	ID int
}
type ListenerInterface interface {
	execute(m string)
}

func (l *Listener) execute(m string) {
	fmt.Printf("%q message receiver for id %d \n", m, l.ID)
}

//Subject is an
type Subject struct {
	listeners []ListenerInterface
}

//AddListener is a
func (s *Subject) addListener(l ListenerInterface) {
	s.listeners = append(s.listeners, l)
}

func (s *Subject) notify(m string) {
	for _, l := range s.listeners {
		if l != nil {
			l.execute(m)
		}
	}
}
