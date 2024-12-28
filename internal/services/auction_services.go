package services

import (
	"context"
	"errors"
	"log/slog"
	"sync"

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
)

type Message struct {
	Message string
	Kind    MessageKind
	UserId  uuid.UUID
	Amount  float64
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
					client.Send <- Message{Message: ErrBidIsToLow.Error(), Kind: FailedToPlaceBid}
				}
			}
			return
		}

		if client, ok := ar.Clients[m.UserId]; ok {
			client.Send <- Message{Message: "Your bid was ssuccessfully placed", Kind: SuccessfullyPlaceBid}
		}

		for id, client := range ar.Clients {
			newBidPalced := Message{Kind: NewBidPlaced, Message: "A new bid was placed", Amount: bid.BidAmount}
			if id == m.UserId {
				continue
			}
			client.Send <- newBidPalced
		}
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
			for _, client := range ar.Clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(ar.Clients, client.UserId)
				}
			}
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
