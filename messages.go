package soopchat

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

func (c *Client) parseJoinChannel(message []byte) (bool, error) {
	// TODO - 수정 필요
	msg := strings.Split(string(message), "\f")
	if len(msg) == 0 {
		return false, fmt.Errorf("invalid JoinChannel: expected at 1 field, got %d", len(msg))
	}

	return msg[1] != "비밀번호가 틀렸습니다.", nil
}

// parseUserJoin 메서드는 전달된 데이터의
// 서비스 코드가 4일 때 이 데이터를 이용해
// User 구조체로 초기화하고 반환한다.
func (c *Client) parseUserJoin(message []byte) ([]UserList, error) {
	msg := strings.Split(string(message), "\f")

	// 최소 길이 (단일 유저 기준)
	if len(msg) < 5 {
		return nil, fmt.Errorf("invalid UserJoin: expected at least 5 fields, got %d", len(msg))
	}

	if len(msg) > 10 {
		// 여러 명의 유저일 경우
		// 최초 실행이기 때문에 무조건 입장했다고 표시함.
		return parseMultiUserList(msg), nil
	}

	// 단일 유저일 경우
	return []UserList{parseSingleUserList(msg)}, nil
}

// parseChatMessage 메서드는 전달된 데이터의
// 서비스 코드가 5일 때 이 데이터를 이용해
// ChatMessage 구조체로 초기화하고 반환한다.
func (c *Client) parseChatMessage(message []byte) (ChatMessage, error) {
	msg := strings.Split(string(message), "\f")
	if len(msg) < 9 {
		return ChatMessage{}, fmt.Errorf("invalid ChatMessage: expected at least 9 fields, got %d", len(msg))
	}

	flags := strings.Split(msg[7], "|")
	userFlag := setFlag(flags)

	// 숫자로 변환 중 에러가 발생하거나 구독하지 않은 상태(-1)라면
	// 가독성을 위해 0으로 변경한다.
	subMonth, err := strconv.Atoi(msg[8])
	if err != nil || subMonth == -1 {
		// return ChatMessage{}, fmt.Errorf("invalid ChatMessage count(%q): %w", msg[8], err)
		subMonth = 0
	}

	cm := ChatMessage{
		User: User{
			ID:             removeParentheses(msg[2]),
			Name:           strings.TrimSpace(msg[6]),
			SubscribeMonth: subMonth,
			Flag:           userFlag,
		},
		Message: strings.TrimSpace(msg[1]),
	}

	return cm, nil
}

// parseBallon 메서드는 전달된 데이터의
// 서비스 코드가 18일 때 이 데이터를 이용해
// Ballon 구조체로 초기화하고 반환한다.
func (c *Client) parseBalloon(message []byte) (Balloon, error) {
	msg := strings.Split(string(message), "\f")
	if len(msg) < 5 {
		return Balloon{}, fmt.Errorf("invalid Balloon: expected at least 5 fields, got %d", len(msg))
	}

	balloonCount, err := strconv.Atoi(msg[4])
	if err != nil {
		return Balloon{}, fmt.Errorf("invalid Balloon count(%q): %w", msg[4], err)
	}

	balloon := Balloon{
		User: User{
			ID:   msg[2],
			Name: msg[3],
		},
		Count: balloonCount,
	}

	return balloon, nil
}

// parseAdballoon 메서드는 전달된 데이터의
// 서비스 코드가 87일 때 이 데이터를 이용해
// Adballoon 구조체로 초기화하고 반환한다.
func (c *Client) parseAdballoon(message []byte) (Adballoon, error) {
	msg := strings.Split(string(message), "\f")
	if len(msg) < 11 {
		return Adballoon{}, fmt.Errorf("invalid Adballoon: expected at least 11 fields, got %d", len(msg))
	}

	adballoonCount, err := strconv.Atoi(msg[10])
	if err != nil {
		return Adballoon{}, fmt.Errorf("invalid Adballoon count(%q): %w", msg[10], err)
	}

	adballoon := Adballoon{
		User: User{
			ID:   msg[3],
			Name: msg[4],
		},
		Count: adballoonCount,
	}

	return adballoon, nil
}

// parseSubscription 메서드는 전달된 데이터의
// 서비스 코드가 91, 93일 때 이 데이터를 이용해
// Subscription 구조체로 초기화하고 반환한다.
func (c *Client) parseSubscription(message []byte, svc int) (Subscription, error) {
	msg := strings.Split(string(message), "\f")

	var user User
	var err error
	count := 1

	switch svc {
	case svc_FOLLOW_ITEM: // 구독
		if len(msg) < 6 {
			return Subscription{}, fmt.Errorf("invalid Subscription(FOLLOW_ITEM): expected at least 6 fields, got %d", len(msg))
		}
		user.ID = removeParentheses(msg[3])
		user.Name = msg[4]
		count, err = strconv.Atoi(msg[5])
		if err != nil {
			return Subscription{}, fmt.Errorf("invalid Subscription(FOLLOW_ITEM) count(%q): %w", msg[5], err)
		}
	case svc_FOLLOW_ITEM_EFFECT: // 구독 이펙트 (언제 실행되는 지 연구 필요.)
		if len(msg) < 5 {
			return Subscription{}, fmt.Errorf("invalid Subscription(FOLLOW_ITEM_EFFECT): expected at least 5 fields, got %d", len(msg))
		}
		user.ID = removeParentheses(msg[2])
		user.Name = msg[3]
		count, err = strconv.Atoi(msg[4])
		if err != nil {
			return Subscription{}, fmt.Errorf("invalid Subscription(FOLLOW_ITEM_EFFECT) count(%q): %w", msg[4], err)
		}
	default:
		return Subscription{}, fmt.Errorf("invalid Subscription service code: %d", svc)
	}

	subscription := Subscription{
		User:  user,
		Count: count,
	}

	return subscription, nil
}

// parseAdminNotice 메서드는 전달된 데이터의
// 서비스 코드가 58일 때 이 데이터를 이용해 문자열을 반환한다.
func (c *Client) parseAdminNotice(message []byte) (string, error) {
	msg := strings.Split(string(message), "\f")

	if len(msg) > 0 {
		return msg[1], nil
	}

	return "", fmt.Errorf("invalid AdminNotice: expected at 2 fields, got %d", len(msg))
}

// parseMission 메서드는 전달된 데이터의
// 서비스 코드가 121일 때 이 데이터를 이용해
// Mission 구조체로 초기화하고 반환한다.
func (c *Client) parseMission(message []byte) (Mission, error) {
	msg := strings.Split(string(message), "\f")
	if len(msg) < 2 {
		return Mission{}, fmt.Errorf("invalid Mission: expected at 2 fields, got %d", len(msg))
	}

	var payload missionPayload
	if err := json.Unmarshal([]byte(msg[1]), &payload); err != nil {
		return Mission{}, fmt.Errorf("invalid Mission: %w", err)
	}

	mission := Mission{
		User: User{
			ID:   payload.UserID,
			Name: payload.UserNick,
		},
		Title: payload.Title,
		Count: payload.BallonCount,
	}

	return mission, nil
}

// parseStreamerNotice 메서드는
// 서비스 코드가 104일 때 이 데이터를 이용해 공지 메시지를 반환한다.
func (c *Client) parseStreamerNotice(message []byte) (string, error) {
	msg := strings.Split(string(message), "\f")
	if len(msg) < 5 {
		return "", fmt.Errorf("invalid StreamerNotice: expected at least 5 field, got %d", len(msg))
	}

	return msg[4], nil
}

// parseOGQMessage 메서드는
// 서비스 코드가 109일 때 이 데이터를 이용해 OGQ 메시지를 반환한다.
func (c *Client) parseOGQMessage(message []byte) (OGQMessage, error) {
	const ogqImageUrl = "https://ogq-sticker-global-cdn-z01.sooplive.co.kr/sticker/%s/%s_80.%s?ver=1"

	msg := strings.Split(string(message), "\f")
	if len(msg) < 20 {
		return OGQMessage{}, fmt.Errorf("invalid OGQMessage: expected at least 20 field, got %d", len(msg))
	}

	// 숫자로 변환 중 에러가 발생하거나 구독하지 않은 상태(-1)라면
	// 가독성을 위해 0으로 변경한다.
	subMonth, err := strconv.Atoi(msg[16])
	if err != nil || subMonth == -1 {
		subMonth = 0
	}

	// 유저 플래그
	flags := strings.Split(msg[8], "|")
	userFlag := setFlag(flags)

	user := User{
		ID:             removeParentheses(msg[6]),
		Name:           msg[7],
		SubscribeMonth: subMonth,
		Flag:           userFlag,
	}
	m := msg[2]
	imageUrl := fmt.Sprintf(ogqImageUrl, msg[3], msg[4], msg[12])

	ogqMessage := OGQMessage{
		User:     user,
		Message:  m,
		ImageURL: imageUrl,
	}

	return ogqMessage, nil
}
