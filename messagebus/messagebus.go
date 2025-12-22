package messagebus

type MessageBus interface {
	Subscribe(topic string) chan []byte
	Unsubscribe(topic string, ch chan []byte)
	Publish(topic string, msg []byte)
}
