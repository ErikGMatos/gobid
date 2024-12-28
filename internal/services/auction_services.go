package services

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type MessageKind int

const (
	//Request
	PlaceBid MessageKind = iota

	//Ok / Success
	SuccessfullyPlaceBid

	// Info
	NewBidPlaced
	AuctionFinished

	//Errors
	FailedToPlaceBid
	InvalidJSON
)

type Message struct {
	Message string      `json:"message,omitempty"`
	Amount  float64     `json:"amount,omitempty"`
	Kind    MessageKind `json:"kind"`
	UserId  uuid.UUID   `json:"user_id,omitempty"`
}

type AuctionLobby struct {
	sync.Mutex
	Rooms map[uuid.UUID]*AuctionRoom
}

type AuctionRoom struct {
	Id         uuid.UUID
	Context    context.Context
	Broadcast  chan Message
	Register   chan *Client
	Unregister chan *Client
	Clients    map[uuid.UUID]*Client

	BidsServices BidsService
}

func (ar *AuctionRoom) registerClient(c *Client) {
	slog.Info("New user connected", "Client", c)
	ar.Clients[c.UserId] = c
}

func (ar *AuctionRoom) unRegisterClient(c *Client) {
	slog.Info("User disconnected", "Client", c)
	delete(ar.Clients, c.UserId)
}

func (ar *AuctionRoom) broadcastMessage(m Message) {
	slog.Info("New message received", "RoomID", ar.Id, "Message", m.Message, "user_id", m.UserId)
	switch m.Kind {
	case PlaceBid:
		bid, err := ar.BidsServices.PlaceBid(ar.Context, ar.Id, m.UserId, m.Amount)
		if err != nil {
			if errors.Is(err, ErrBidIsToLow) {
				if client, ok := ar.Clients[m.UserId]; ok {
					client.Send <- Message{Message: ErrBidIsToLow.Error(), Kind: FailedToPlaceBid, UserId: m.UserId}
				}
			}
			return
		}

		if client, ok := ar.Clients[m.UserId]; ok {
			client.Send <- Message{Message: "Your bid was ssuccessfully placed", Kind: SuccessfullyPlaceBid, UserId: m.UserId}
		}

		for id, client := range ar.Clients {
			newBidPalced := Message{Kind: NewBidPlaced, Message: "A new bid was placed", Amount: bid.BidAmount, UserId: m.UserId}
			if id == m.UserId {
				continue
			}
			client.Send <- newBidPalced
		}

	case InvalidJSON:
		client, ok := ar.Clients[m.UserId]
		if !ok {
			slog.Info("Client not found in hashmap", "user_id", m.UserId)
			return
		}
		client.Send <- m
	}
}

func (ar *AuctionRoom) Run() {
	slog.Info("Auction has begun", "AuctionId", ar.Id)
	defer func() {
		close(ar.Broadcast)
		close(ar.Register)
		close(ar.Unregister)
	}()

	for {
		select {
		case client := <-ar.Register:
			ar.registerClient(client)
		case client := <-ar.Unregister:
			ar.unRegisterClient(client)
		case message := <-ar.Broadcast:
			ar.broadcastMessage(message)
		case <-ar.Context.Done():
			slog.Info("Auction has ended.", "AuctionID", ar.Id)
			for _, client := range ar.Clients {
				client.Send <- Message{Message: "Auction has been finished", Kind: AuctionFinished}
			}
			return
		}
	}
}

func NewAuctionRoom(ctx context.Context, id uuid.UUID, bidsServices BidsService) *AuctionRoom {
	return &AuctionRoom{
		Id:           id,
		Broadcast:    make(chan Message),
		Register:     make(chan *Client),
		Unregister:   make(chan *Client),
		Clients:      make(map[uuid.UUID]*Client),
		Context:      ctx,
		BidsServices: bidsServices,
	}
}

type Client struct {
	Room   *AuctionRoom
	Conn   *websocket.Conn
	Send   chan Message
	UserId uuid.UUID
}

func NewClient(room *AuctionRoom, conn *websocket.Conn, userId uuid.UUID) *Client {
	return &Client{
		Room:   room,
		Conn:   conn,
		Send:   make(chan Message, 512),
		UserId: userId,
	}
}

const (
	maxMessageSize = 512
	readDeadline   = 60 * time.Second
	writeWait      = 10 * time.Second
	pingPeriod     = (readDeadline * 9) / 10
)

func (c *Client) ReadEventLoop() {
	defer func() {
		c.Room.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(readDeadline))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(readDeadline))
		return nil
	})

	for {
		var m Message
		m.UserId = c.UserId
		err := c.Conn.ReadJSON(&m)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				slog.Error("Unexpected close error", "Error", err)
				return
			}
			c.Room.Broadcast <- Message{Message: "this message should be a valid JSON", Kind: InvalidJSON, UserId: m.UserId}
			continue
		}
		c.Room.Broadcast <- m
	}
}

func (c *Client) WriteEventLoop() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				c.Conn.WriteJSON(Message{Message: "Close websocket connection", Kind: websocket.CloseMessage})
				return
			}
			if message.Kind == AuctionFinished {
				close(c.Send)
				return
			}
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			err := c.Conn.WriteJSON(message)
			if err != nil {
				c.Room.Unregister <- c
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				slog.Error("Unexpected error", "error", err)
				return
			}
		}
	}
}
