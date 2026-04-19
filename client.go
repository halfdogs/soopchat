package soopchat

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/gorilla/websocket"
)

// NewClient 함수는 Client 구조체를
// 초기화하여 생성한다.
func NewClient(token Token) (*Client, error) {
	// StreamerID가 있어야 SocketAddress 및 ChatRoom 설정하므로
	// 필수 토큰이다. 없을 경우 에러를 반환한다.
	if token.StreamerID == "" {
		return &Client{}, errors.New("StreamerID is missing from token")
	}

	// resty client 생성
	httpClient := resty.New()
	httpClient.SetTimeout(time.Duration(10 * time.Second))

	return &Client{
		Token:      token,
		read:       make(chan []byte, 1024),
		handshake:  make([][]byte, 2),
		apiService: apiService{http: httpClient},
	}, nil
}

func (c *Client) SetOption(option Option) *Client {
	c.option = option
	return c
}

// Connect 메서드는 채팅 서버 연결에 필요한
// 과정을 수행한다.
func (c *Client) Connect(password ...string) error {
	// 패스워드가 있다면 필드에 값을 대입한다.
	if len(password) > 0 {
		c.Token.streamPassword = password[0]
	}

	// Identifier 값이 있다면 로그인 과정을 자동으로 수행한다.
	if c.Token.Identifier.ID != "" && c.Token.Identifier.Password != "" {
		err := c.apiService.login(*c)
		if err != nil {
			// 에러가 발생하였다면 로그인 실패
			if c.onLogin != nil {
				c.onLogin(false)
			}

			// 로그인 실패 시 에러를 반환하는 대신
			// 콜백으로 전달
			c.reportError(err)
			return err
		}

		if c.onLogin != nil {
			c.onLogin(true)
		}
	}

	// 자동으로 Socket Address 및 Chat Room를 가져옵니다.
	err := c.apiService.setSocketData(c)
	if err != nil {
		c.reportError(err)
		return err
	}

	// websocket 생성/연결 작업을 수행한다.
	err = c.createWebsocket()
	if err != nil {
		c.reportError(err)
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
	case svc_LOGIN:
		err = c.setLoginHandshke()
		if err != nil {
			c.reportError(err)
			return err
		}
	case svc_JOINCH:
		err = c.setJoinHandshake()
		if err != nil {
			c.reportError(err)
			return err
		}
	}

	// 핸드쉐이크 수행
	err = c.setHandshake(svc)
	if err != nil {
		c.reportError(err)
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

		c.reportError(err)
		return err
	}

	// 채널 접속에 성공할 경우
	// onConnect 콜백에 true를 전달한다.
	if svc == svc_JOINCH {
		if c.onConnect != nil {
			c.onConnect(true)
		}
	}

	return nil
}

// processSocket 메서드는 웹소켓으로
// 들어오는 데이터를 처리한다.
func (c *Client) processSocket() error {
	// WaitGroup을 생성해 작업 완료까지 대기한다.
	var wg sync.WaitGroup
	wg.Add(1)

	// 함수가 종료될 떄 타이머와 소켓을 종료한다.
	defer func() {
		if c.pingpongTimer != nil {
			c.pingpongTimer.Stop()
		}
		if c.socket != nil {
			_ = c.socket.Close()
		}
		wg.Wait()
	}()

	// 웹소켓으로 넘어오는 데이터를 비동기 처리한다.
	// 이 때 에러가 발생하면 작업이 완료된다.
	go c.reader(&wg)

	// 아빠 안잔다.
	c.pingpong()

	// 로그인 핸드쉐이크
	// 이 때 에러가 발생하면 작업이 완료된다.
	if err := c.executeHandshake(svc_LOGIN); err != nil {
		return err
	}

	// 웹소켓으로 넘어오는 데이터를 분석/가공한다.
	if err := c.startParser(); err != nil {
		return err
	}

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
	// 에러 발생 시 웹소켓 연결 실패이므로 루프를 탈출한다.
	for {
		_, msg, err := c.socket.ReadMessage()
		if err != nil {
			if c.onConnect != nil {
				c.onConnect(false)
			}
			c.reportError(err)
			return
		}

		c.read <- msg
	}
}

// startParser 메서드는 read 필드로 전달된 데이터를
// 처리하여 콜백 함수로 전달한다.
func (c *Client) startParser() error {
	for msg := range c.read {
		if c.onRawMessage != nil {
			c.onRawMessage(strconv.Quote(string(msg)))
		}

		svc, err := getServiceCode(msg)
		if err != nil {
			c.reportError(err)
		}

		switch svc {
		case svc_LOGIN: // Login, need JOIN handshake
			// 로그인 단계에서 실패할 경우엔 에러를 반환한다.
			if err := c.executeHandshake(svc_JOINCH); err != nil {
				c.reportError(err)
				return err
			}

		case svc_JOINCH: // 채널 입장
			dispatch(c, c.parseJoinChannel, c.onJoinChannel, msg)

		case svc_CHUSER: // 입장/퇴장
			dispatch(c, c.parseUserJoin, c.onUserLists, msg)

		case svc_CHATMESG: // Chat
			dispatch(c, c.parseChatMessage, c.onChatMessage, msg)

		case svc_SENDBALLOON: // 별풍선
			dispatch(c, c.parseBalloon, c.onBalloon, msg)

		case svc_ADCON_EFFECT: // 애드벌룬
			dispatch(c, c.parseAdballoon, c.onAdballoon, msg)

		case svc_FOLLOW_ITEM, svc_FOLLOW_ITEM_EFFECT: // 신규 구독 / 연속 구독
			parseFunc := func(b []byte) (Subscription, error) {
				return c.parseSubscription(b, svc)
			}
			dispatch(c, parseFunc, c.onSubscription, msg)

		case svc_SENDADMINNOTICE: // 어드민 메시지
			dispatch(c, c.parseAdminNotice, c.onAdminNotice, msg)

		case svc_MISSION: // 도전미션
			dispatch(c, c.parseMission, c.onMission, msg)

		case svc_BJ_NOTICE: // 스트리머 채팅 공지
			dispatch(c, c.parseStreamerNotice, c.onStreamerNotice, msg)

		case svc_OGQ_EMOTICON: // OGQ 메시지
			dispatch(c, c.parseOGQMessage, c.onOGQMessage, msg)

		case svc_CLOSE_BROAD: // 방송 종료
			if c.onClose != nil {
				c.onClose(true)
			}

			if c.option.AutoClose {
				c.socket.Close()
				return nil
			}
		}
	}

	return nil
}

// reportError 메서드는 유효한 에러가 있을 때 onError 콜백을 호출한다.
func (c *Client) reportError(err error) {
	if err != nil && c.onError != nil {
		c.onError(err)
	}
}

// dispatch 메서드는 메시지를 parse로 해석하고
// 결과를 callback에 전달한다.
func dispatch[T any](
	c *Client, parse func([]byte) (T, error), callback func(T), msg []byte,
) {
	// 콜백이 없다면 반환
	if callback == nil {
		return
	}

	// 전달받은 파서 함수에 msg를 넣는다
	v, err := parse(msg)
	if err != nil {
		c.reportError(err)
		return
	}

	// 전달받은 콜백 호출
	callback(v)
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
			headerbuf := makeHeader(svc_KEEPALIVE, len(bodyBuf), 0)
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
	packet = append(packet, "\f", c.Token.authTicket, "\f", "\f", "", "\f")

	return c.setHandshakeData(1, packet)
}

// setJoinHandshake 메서드는 채팅 서버 연결에
// 필요한 Join Handshake 데이터를 준비한다.
func (c *Client) setJoinHandshake() error {
	infoPacket := append(
		c.setLogHandshake(defaultLog()),
		c.setInfoHandshake(defaultInfo(c.Token.streamPassword))...,
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

// setLogHandshake 메서드는 Handshake 과정 중
// 필요한 Log 데이터를 가공한다.
func (c *Client) setLogHandshake(log log) []byte {
	result := append([]byte("log"), 17)
	result = append(result, c.setLogValue(log)...)
	result = append(result, 18)

	return result
}

// setInfoHandshake 메서드는 Handshake 과정 중
// 필요한 Info 데이터를 가공한다.
func (c *Client) setInfoHandshake(info info) []byte {
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
func (c *Client) setLogValue(log log) []byte {
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
