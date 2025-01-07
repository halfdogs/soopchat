package soopchat

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/gorilla/websocket"
)

// NewClient 함수는 Client 구조체를
// 초기화하여 생성한다.
func NewClient(token Token) (*Client, error) {
	// BJ ID가 있어야 SocketAddress 및 ChatRoom 설정하므로
	// 필수 토큰이다. 없을 경우 에러를 반환한다.
	if token.StreamerID == "" {
		return &Client{}, errors.New("need bj id value")
	}

	// resty client 생성
	httpClient := resty.New()
	httpClient.SetTimeout(time.Duration(2 * time.Second))

	return &Client{
		Token:           token,
		read:            make(chan []byte, 1024),
		handshake:       make([][]byte, 2),
		channelPassword: "",
		httpClient:      httpClient,
	}, nil
}

// MustConnect 메서드는 채팅 서버 연결에 필요한
// 과정을 수행한다.
// Panic이 발생하는 원인을 해결하기 전까지 Must prefix
func (c *Client) MustConnect(password ...string) error {
	// 패스워드가 있다면 필드에 값을 대입한다.
	if len(password) > 0 {
		c.channelPassword = password[0]
	}

	// Identifier 값이 있다면 로그인 과정을 수행한다.
	if c.Token.Identifier.ID != "" && c.Token.Identifier.Password != "" {
		err := c.login()
		if err != nil {
			if c.onError != nil {
				c.onLogin(false)
			}

			return err
		}

		if c.onLogin != nil {
			c.onLogin(true)
		}
	}

	// 자동으로 Socket Address 및 Chat Room를 가져옵니다.
	err := c.setSocketData()
	if err != nil {
		return err
	}

	// websocket 생성/연결 작업을 수행한다.
	err = c.createWebsocket()
	if err != nil {
		return err
	}

	// 웹소켓으로 들어오는 데이터를 처리한다.
	// 처리 중 에러가 발생하면 에러를 반환한다.
	return c.processSocket()
}

// executeHandshake 메서드는 핸드쉐이크 과정을 실행합니다.
func (c *Client) executeHandshake(svc int) error {
	var err error

	// 서비스코드 값에 따라 핸드쉐이크 준비
	switch svc {
	case SVC_LOGIN:
		err = c.setLoginHandshke()
		if err != nil {
			return err
		}
	case SVC_JOINCH:
		err = c.setJoinHandshake()
		if err != nil {
			return err
		}
	}

	// 핸드쉐이크 수행
	err = c.setHandshake(svc)
	if err != nil {
		return err
	}

	return nil
}

// setHandshake 메서드는 채팅 서버 연결에 필요한
// 핸드쉐이크 과정을 수행한다.
// 이 때 2번째 단계를 수행했을 경우
// onConnect 콜백으로 값을 전달한다.
func (c *Client) setHandshake(svc int) error {
	// 핸드쉐이크를 전송하고 에러가 있을 경우
	// onConnect 콜백에 false를 전달하고 에러를 반환한다.
	err := c.socket.WriteMessage(websocket.BinaryMessage, c.handshake[svc-1])
	if err != nil {
		if c.onConnect != nil {
			c.onConnect(false)
		}
		return err
	}

	// 채널 접속에 성공할 경우
	// onConnect 콜백에 true를 전달한다.
	if svc == SVC_JOINCH {
		if c.onConnect != nil {
			c.onConnect(true)
		}
	}

	return nil
}

// processSocket 메서드는 웹소켓으로
// 들어오는 데이터를 처리한다.
func (c *Client) processSocket() error {
	// 함수가 종료되기 전에 소켓을 닫는다.
	defer c.socket.Close()

	// WaitGroup을 생성해 작업 완료까지 대기한다.
	wg := sync.WaitGroup{}
	wg.Add(1)

	// 웹소켓으로 넘어오는 데이터를 비동기 처리한다.
	// 이 때 에러가 발생하면 작업이 완료된다.
	go c.reader(&wg)

	// 아빠 안잔다.
	c.pingpong()
	defer c.pingpongTimer.Stop()

	// 로그인 핸드쉐이크
	// 이 때 에러가 발생하면 작업이 완료된다.
	err := c.executeHandshake(SVC_LOGIN)
	if err != nil {
		wg.Done()
		return err
	}

	// 웹소켓으로 넘어오는 데이터를 분석/가공한다.
	err = c.startParser()
	if err != nil {
		wg.Done()
		return err
	}

	// 모든 작업이 완료될 때까지 대기한다.
	wg.Wait()
	return nil
}

// reader 메서드는 웹소켓으로 들어오는 데이터를
// read 필드로 전달한다.
func (c *Client) reader(wg *sync.WaitGroup) {
	// 에러가 발생하여 무한 루프가 끝나고 함수가 반환될 때
	// 작업을 완료시킨다.
	defer wg.Done()

	// 작업이 완료될 때까지 계속 웹소켓으로 들어오는 데이터를
	// 리시버의 read 필드로 전달한다.
	// 에러가 발생할 경우 read 필드에 error 를 전달한다.
	for {
		_, msg, err := c.socket.ReadMessage()
		if err != nil {
			c.read <- []byte(fmt.Sprintf("error: %s", err.Error()))
			continue
		}

		c.read <- msg
	}
}

// startParser 메서드는 read 필드로 전달된 데이터를
// 처리하여 콜백 함수로 전달한다.
func (c *Client) startParser() error {
	for msg := range c.read {
		if strings.HasPrefix(string(msg), "error: ") {
			return errors.New(string(msg))
		}

		if c.onRawMessage != nil {
			c.onRawMessage(fmt.Sprintf("%q", msg))
		}

		svc, err := getServiceCode(msg)
		if err != nil && c.onError != nil {
			c.onError(err)
		}

		switch svc {
		case SVC_LOGIN: // Login, need JOIN handshake
			// 로그인 단계에서 실패할 경우엔 에러를 반환한다.
			err := c.executeHandshake(SVC_JOINCH)
			if err != nil {
				if c.onError != nil {
					c.onError(err)
				}
				return err
			}
		case SVC_JOINCH: // 채널 입장
			if c.onJoinChannel != nil {
				if b := c.parseJoinChannel(msg); b {
					c.onJoinChannel(true)
				} else {
					c.onJoinChannel(false)
				}
			}
		case SVC_CHUSER: // 입장/퇴장
			if c.onUserLists != nil {
				m := c.parseUserJoin(msg)
				c.onUserLists(m)
			}
		case SVC_CHATMESG: // Chat
			if c.onChatMessage != nil {
				m, err := c.parseChatMessage(msg)
				if err != nil {
					if c.onError != nil {
						c.onError(err)
					}
				} else {
					c.onChatMessage(m)
				}
			}
		case SVC_SENDBALLOON: // 별풍선
			if c.onBalloon != nil {
				m, err := c.parseBalloon(msg)
				if err != nil {
					if c.onError != nil {
						c.onError(err)
					}
				} else {
					c.onBalloon(m)
				}
			}
		case SVC_ADCON_EFFECT: // 애드벌룬
			if c.onAdballoon != nil {
				m, err := c.parseAdballoon(msg)
				if err != nil {
					if c.onError != nil {
						c.onError(err)
					}
				} else {
					c.onAdballoon(m)
				}
			}
		case SVC_FOLLOW_ITEM, SVC_FOLLOW_ITEM_EFFECT: // 신규 구독 / 연속 구독
			if c.onSubscription != nil {
				m, err := c.parseSubscription(msg, svc)
				if err != nil {
					if c.onError != nil {
						c.onError(err)
					}
				} else {
					c.onSubscription(m)
				}
			}
		case SVC_SENDADMINNOTICE: // 어드민 메시지
			if c.onAdminNotice != nil {
				m, err := c.parseAdminNotice(msg)
				if err != nil {
					if c.onError != nil {
						c.onError(err)
					}
				} else {
					c.onAdminNotice(m)
				}
			}
		case SVC_MISSION: // 도전미션
			if c.onMission != nil {
				m, err := c.parseMission(msg)
				if err != nil {
					if c.onError != nil {
						c.onError(err)
					}
				} else {
					c.onMission(m)
				}
			}
		}
	}

	return nil
}

// SendChatMessage 메서드는 채팅 채널에 채팅 데이터를 전송한다.
// 메시지를 보낼 때 실패한 경우 에러를 반환한다.
func (c *Client) SendChatMessage(message string) error {
	if c.Token.authTicket == "" {
		return errors.New("cannot non-member send message")
	}

	var tBuf []string
	tBuf = append(tBuf, "\f", message, "\f", "0", "\f")
	bodyBuf := makeBuffer(tBuf)
	headerBuf := makeHeader(5, len(bodyBuf), 0)

	packet := append(headerBuf, bodyBuf...)
	return c.socket.WriteMessage(websocket.BinaryMessage, packet)
}

// pingpong 메서드는 매 1분마다 ping 데이터를
// 전송한다.
func (c *Client) pingpong() {
	c.pingpongTimer = time.NewTicker(1 * time.Minute)

	go func() {
		for range c.pingpongTimer.C {
			bodyBuf := makeBuffer([]string{"\f"})
			headerbuf := makeHeader(SVC_KEEPALIVE, len(bodyBuf), 0)
			p := append(headerbuf, bodyBuf...)
			c.socket.WriteMessage(websocket.BinaryMessage, p)
		}
	}()
}

// createWebsocket 메서드는 아프리카TV 채팅서버에
// 소켓을 연결한다.
func (c *Client) createWebsocket() error {
	// 이미 존재하는 소켓이라면 반환한다.
	if c.socket != nil {
		return nil
	}

	// 웹소켓 설정
	dialer := websocket.Dialer{
		HandshakeTimeout: 5 * time.Second, // 설정하지 않으면 너무 오래 대기함.
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	header := http.Header{}
	header.Set("Sec-WebSocket-Protocol", "chat")

	// 웹소켓 연결
	var err error
	c.socket, _, err = dialer.Dial(c.socketAddress, header)
	return err
}

// setLoginHandshake 메서드는 채팅 서버 연결에
// 필요한 Login Handshake 데이터를 준비한다.
func (c *Client) setLoginHandshke() error {
	var packet []string
	packet = append(packet, "\f", c.Token.authTicket, "\f", "\f", c.Token.Flag, "\f")

	return c.setHandshakeData(1, packet)
}

// setJoinHandshake 메서드는 채팅 서버 연결에
// 필요한 Join Handshake 데이터를 준비한다.
func (c *Client) setJoinHandshake() error {
	infoPacket := append(
		c.SetLogHandshake(DefaultLog()),
		c.SetInfoHandshake(DefaultInfo(c.channelPassword))...,
	)
	var packet []string
	packet = append(
		packet,
		"\f",
		c.Token.chatRoom,
		"\f",
		"\f",
		c.Token.fanTicket,
		"0",
		"\f",
		"",
		"\f",
		string(infoPacket),
		"\f",
	)

	return c.setHandshakeData(2, packet)
}

// setHandshakeData 메서드는 아프리카TV 채팅 서버에 연결할 때
// 필요한 데이터를 생성하는 과정을 수행한다.
func (c *Client) setHandshakeData(svc int, packet []string) error {
	bodyBuf := makeBuffer(packet)
	headerBuf := makeHeader(svc, len(bodyBuf), 0)
	p := append(headerBuf, bodyBuf...)

	c.handshake[svc-1] = p

	return nil
}

// SetLogHandshake 메서드는 Handshake 과정 중
// 필요한 Log 데이터를 가공한다.
func (c *Client) SetLogHandshake(log Log) []byte {
	result := append([]byte("log"), 17)
	result = append(result, c.setLogValue(log)...)
	result = append(result, 18)

	return result
}

// SetInfoHandshake 메서드는 Handshake 과정 중
// 필요한 Info 데이터를 가공한다.
func (c *Client) SetInfoHandshake(info Info) []byte {
	var result []byte
	infoValue := reflect.ValueOf(info)

	for i := 0; i < infoValue.NumField(); i++ {
		field := infoValue.Field(i)
		if !field.IsZero() {
			k := strings.ToLower(infoValue.Type().Field(i).Tag.Get("json"))
			v := fmt.Sprintf("%v", field.Interface())
			kv := append([]byte(k), 17)
			kv = append(kv, []byte(v)...)
			kv = append(kv, 18)
			result = append(result, kv...)
		}
	}

	return result
}

// setLogValue 메서드는 Handshake 과정 중
// Log 구조체를 []byte 로 변환한다.
func (c *Client) setLogValue(log Log) []byte {
	var result []byte
	logValue := reflect.ValueOf(log)

	for i := 0; i < logValue.NumField(); i++ {
		field := logValue.Field(i)
		if !field.IsZero() {
			k := strings.ToLower(logValue.Type().Field(i).Tag.Get("json"))
			v := fmt.Sprintf("%v", field.Interface())
			kv := append([]byte{6}, []byte(k)...)
			kv = append(kv, 6, '=', 6)
			kv = append(kv, []byte(v)...)
			kv = append(kv, 6, '&')
			result = append(result, kv...)
		}
	}

	return append([]byte{6, 38}, result...)
}
