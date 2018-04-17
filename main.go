package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

const (
	StatusStop = iota
	StatusStart
)

type Message struct {
	Operation string `json:"op,omitempty"`
	Number    int    `json:"number,omitempty"`
}

type Pool struct {
	Clients []*Client
	sync.RWMutex
}

type Client struct {
	Conn   net.Conn
	Msg    chan *Message
	Close  chan struct{}
	Status int
}

func NewPool() *Pool {
	p := &Pool{
		Clients: []*Client{},
	}
	go p.Broadcast()
	return p
}

func NewClient(conn net.Conn) *Client {
	c := &Client{
		Conn:   conn,
		Msg:    make(chan *Message),
		Close:  make(chan struct{}),
		Status: StatusStop,
	}
	go c.Writing()
	go c.Reading()
	return c
}

func (c *Client) Reading() {
	for {
		m, _, err := wsutil.ReadClientData(c.Conn)
		if err != nil {
			close(c.Close)
			c.Conn.Close()
			return
		}
		msg := &Message{}
		err = json.Unmarshal(m, msg)
		if err != nil {
			log.Println(err)
		}
		switch msg.Operation {
		case "start":
			c.Status = StatusStart
		case "stop":
			c.Status = StatusStop
		}
	}
}

func (c *Client) Writing() {
	for {
		select {
		case <-c.Close:
			return
		case msg := <-c.Msg:
			m, _ := json.Marshal(msg)
			wsutil.WriteServerMessage(c.Conn, ws.OpText, m)
		}
	}
}

func (p *Pool) Add(c *Client) {
	p.Lock()
	p.Clients = append(p.Clients, c)
	log.Printf("Client #%d connected\n", len(p.Clients)-1)
	p.Unlock()
}

func (p *Pool) Remove(i int) {
	p.Lock()
	p.Clients[i] = nil
	p.Clients = append(p.Clients[:i], p.Clients[i+1:]...)
	log.Printf("Client #%d diconnected\n", i)
	p.Unlock()
}

func (p *Pool) Broadcast() {
	rand.Seed(time.Now().UnixNano())
	for {
		<-time.After(1 * time.Second)
		n := rand.Intn(1000)
		for i, c := range p.Clients {
			select {
			case <-c.Close:
				p.Remove(i)
			default:
				if c.Status == StatusStart {
					c.Msg <- &Message{
						Number: n,
					}
				}
			}
		}
	}
}

func IndexPage(t *template.Template) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		t.ExecuteTemplate(w, "main", nil)
	})
}

func WS(p *Pool) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, _, _, err := ws.UpgradeHTTP(r, w, nil)
		if err != nil {
			log.Println(err)
			return
		}
		p.Add(NewClient(conn))
	})
}

func main() {
	p := NewPool()
	t := template.Must(template.ParseGlob("template/*.tpl"))
	http.Handle("/", IndexPage(t))
	http.Handle("/ws", WS(p))
	fmt.Println("Start listening on port 8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
