package soopchat

import (
	"errors"

	"github.com/go-resty/resty/v2"
	"github.com/gorilla/websocket"
	"github.com/tidwall/gjson"
)

type Client struct {
	channel    *channel
	connection *connection
	identifier *identifier
	apiService *apiService
}

type connection struct {
	socket        *websocket.Conn
	socketAddress string
	read          chan []byte
}

type identifier struct {
	ID       string
	Password string
}

type channel struct {
	streamer string
	chatRoom string
	password string
}

type apiService struct {
	httpClient *resty.Client
}

var (
	ErrNeedStreamerID = errors.New("need a streamer ID")
	ErrAPIResponseNil = errors.New("api response is nil")
	ErrLoginRequired  = errors.New("login required")
)

func New() *Client {
	return &Client{
		channel: &channel{
			streamer: "",
			password: "",
			chatRoom: "",
		},
		connection: &connection{
			socket:        &websocket.Conn{},
			socketAddress: "",
			read:          make(chan []byte, 1024),
		},
		identifier: &identifier{
			ID:       "",
			Password: "",
		},
		apiService: &apiService{
			httpClient: resty.New(),
		},
	}
}

func (c *Client) WithIdentifier(id, password string) *Client {
	c.identifier = &identifier{
		ID:       id,
		Password: password,
	}

	return c
}

func (c *Client) WithStreamer(streamerID string) *Client {
	c.channel.streamer = streamerID
	return c
}

func (c *Client) Connect(password ...string) error {
	if len(password) != 0 {
		c.channel.password = password[0]
	}

	// 스트리머를 설정하지 않았을 경우
	if c.channel.streamer == "" {
		return ErrNeedStreamerID
	}

	// identifier가 있다면 로그인을 시도한다.
	if c.identifier.ID != "" && c.identifier.Password != "" {
		err := c.apiService.login(*c.identifier)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) setSocketData() error {
	resp, err := c.apiService.setSocketData(c)
	if err != nil {
		return err
	}

	if resp == nil {
		return ErrAPIResponseNil
	}

	jsonResult := gjson.GetBytes(resp, "CHANNEL")
	result := jsonResult.Get("RESULT").Int()
	switch result {
	case -6:
		return ErrLoginRequired
	}

	return nil
}
