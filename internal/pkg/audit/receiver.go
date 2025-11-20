package audit

type Receiver interface {
	AddEvent(event *Event) error
}
