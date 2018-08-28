package mqpool_test

import (
	"log"
	"testing"

	"github.com/streadway/amqp"

	"github.com/go-mqpool/mqpool"
)

var p mqpool.Pool

func TestReceive(t *testing.T) {
	var err error
	p, err = mqpool.NewConnPool(10, 100, "amqp://guest:guest@127.0.0.1:5672/")
	if err != nil {
		t.Error("Pool ", err)
		return
	}
	conn, err := p.Get()
	if err != nil {
		t.Error("Connection ", err)
		return
	}
	defer conn.Close()
	ch, err := conn.Channel()
	if err != nil {
		t.Error("Channel ", err)
		return
	}
	defer ch.Close()

	err = ch.ExchangeDeclare("amq.direct", "direct", true, false, false, false, nil)
	if err != nil {
		t.Error("ExchangeDeclare ", err)
		return
	}

	q, err := ch.QueueDeclare("", false, false, true, false, nil)
	if err != nil {
		t.Error("QueueDeclare ", err)
		return
	}
	err = ch.QueueBind(q.Name, "info", "amq.direct", false, nil)

	msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		t.Error("Consume ", err)
		return
	}

	forever := make(chan bool)

	go func() {
		var i = 0
		for d := range msgs {
			log.Println("receive msg :", string(d.Body))
			i++
			if i >= 10000 {
				break
			}
		}
		forever <- true
	}()
	for j := 0; j <= 10000; j++ {
		testSend(t)
	}
	<-forever
	p.Close()
}

func testSend(t *testing.T) {

	conn, err := p.Get()
	if err != nil {
		t.Error("Connection ", err)
		return
	}
	defer conn.Close()
	ch, err := conn.Channel()
	if err != nil {
		t.Error("Channel ", err)
		return
	}
	defer ch.Close()

	err = ch.ExchangeDeclare("amq.direct", "direct", true, false, false, false, nil)
	if err != nil {
		t.Error("ExchangeDeclare ", err)
		return
	}
	msg := "test content"
	err = ch.Publish("amq.direct", "info", false, false, amqp.Publishing{ContentType: "text/plain", Body: []byte(msg)})
	if err != nil {
		t.Error("ExchangeDeclare ", err)
		return
	}
	log.Println("sent msg : ", msg)

}
