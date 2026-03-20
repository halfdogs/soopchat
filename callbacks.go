package soopchat

/*************
 * callbacks *
 *************/

// OnError 메서드는 웹 소켓 통신 이후의 발생한 에러 데이터를 전달한다.
func (c *Client) OnError(callback func(err error)) {
	c.onError = callback
}

// OnConnect 메서드는 서버와 연결에 관한 데이터를 전달한다.
func (c *Client) OnConnect(callback func(connect bool)) {
	c.onConnect = callback
}

// OnJoinChannel 메서드는 채널 접속에 관한 데이터를 전달한다.
func (c *Client) OnJoinChannel(callback func(join bool)) {
	c.onJoinChannel = callback
}

// OnRawMessage 메서드는 메시지 원문 데이터를 전달한다.
func (c *Client) OnRawMessage(callback func(message string)) {
	c.onRawMessage = callback
}

// OnChatMessage 메서드는 채팅 메시지가 왔을 때 데이터를 전달한다.
func (c *Client) OnChatMessage(callback func(message ChatMessage)) {
	c.onChatMessage = callback
}

// OnUserLists 메서드는 유저 입장/퇴장 데이터를 전달한다.
func (c *Client) OnUserLists(callback func(userlist []UserList)) {
	c.onUserLists = callback
}

// OnBalloon 메서드는 별풍선 데이터를 전달한다.
func (c *Client) OnBalloon(callback func(balloon Balloon)) {
	c.onBalloon = callback
}

// OnAdballoon 메서드는 애드벌룬 데이터를 전달한다.
func (c *Client) OnAdballoon(callback func(adballoon Adballoon)) {
	c.onAdballoon = callback
}

// OnSubscription 메서드는 구독 데이터를 전달한다.
func (c *Client) OnSubscription(callback func(subscription Subscription)) {
	c.onSubscription = callback
}

// OnAdminNotice 메서드는 운영자 알림 데이터를 전달한다.
func (c *Client) OnAdminNotice(callback func(message string)) {
	c.onAdminNotice = callback
}

// OnMission 메서드는 도전미션 데이터를 전달한다.
func (c *Client) OnMission(callback func(mission Mission)) {
	c.onMission = callback
}

// OnStreamerNotice 메서드는 방송 공지 메시지 데이터를 전달한다.
func (c *Client) OnStreamerNotice(callback func(message string)) {
	c.onStreamerNotice = callback
}

/*****************
 * API callbacks *
 *****************/

// OnLogin 메서드는 로그인 성공/실패 데이터를 전달한다
func (c *Client) OnLogin(callback func(isLoginSuccess bool)) {
	c.onLogin = callback
}
