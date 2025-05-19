package soopchat

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
)

func (c *Client) parseJoinChannel(message []byte) bool {
	msg := strings.Split(string(message), "\f")
	return msg[1] != "비밀번호가 틀렸습니다."
}

// parseUserJoin 메서드는 전달된 데이터의
// 서비스 코드가 4일 때 이 데이터를 이용해
// User 구조체로 초기화하고 반환한다.
func (c *Client) parseUserJoin(message []byte) []UserList {
	msg := strings.Split(string(message), "\f")

	if len(msg) > 10 {
		// 여러 명의 유저일 경우
		// 최초 실행이기 때문에 무조건 입장했다고 표시함.
		return parseMultiUserList(msg)
	}

	// 단일 유저일 경우
	return []UserList{parseSingleUserList(msg)}
}

// parseChatMessage 메서드는 전달된 데이터의
// 서비스 코드가 5일 때 이 데이터를 이용해
// ChatMessage 구조체로 초기화하고 반환한다.
func (c *Client) parseChatMessage(message []byte) (ChatMessage, error) {
	msg := strings.Split(string(message), "\f")
	if !(len(msg) >= 8) {
		return ChatMessage{}, errors.New("message splitting failure [5]")
	}

	flags := strings.Split(msg[7], "|")
	userFlag := setFlag(flags)
	subMonth, err := strconv.Atoi(msg[8])
	if err != nil {
		return ChatMessage{}, err
	}

	// 구독 개월이 -1 이라면 구독을 하지 않은 사람이지만
	// 가독성을 위해 0으로 바꿈
	if subMonth == -1 {
		subMonth = 0
	}

	cm := ChatMessage{
		User: User{
			ID:             removeParentheses(strings.TrimSpace(msg[2])),
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
	if !(len(msg) >= 4) {
		return Balloon{}, errors.New("message splitting failure [18]")
	}

	balloonCount, err := strconv.Atoi(msg[4])
	if err != nil {
		return Balloon{}, err
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
	if !(len(msg) >= 10) {
		return Adballoon{}, errors.New("message splitting failure [87]")
	}

	adballoonCount, err := strconv.Atoi(msg[10])
	if err != nil {
		return Adballoon{}, err
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
	if !(len(msg) >= 8) {
		return Subscription{}, errors.New("message splitting failure [91]")
	}

	var user User
	var err error
	count := 1

	switch svc {
	case svc_FOLLOW_ITEM: // 구독
		user.ID = removeParentheses(msg[3])
		user.Name = msg[4]
		count, err = strconv.Atoi(msg[5])
		if err != nil {
			return Subscription{}, err
		}
	case svc_FOLLOW_ITEM_EFFECT: // 구독 이펙트 (언제 실행되는 지 연구 필요.)
		user.ID = removeParentheses(msg[2])
		user.Name = msg[3]
		count, err = strconv.Atoi(msg[4])
		if err != nil {
			return Subscription{}, err
		}
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

	return "", errors.New("message splitting failure [58]")
}

// parseMission 메서드는 전달된 데이터의
// 서비스 코드가 121일 때 이 데이터를 이용해
// Mission 구조체로 초기화하고 반환한다.
func (c *Client) parseMission(message []byte) (Mission, error) {
	msg := strings.Split(string(message), "\f")
	if !(len(msg) > 1) {
		return Mission{}, errors.New("message splitting failure [121]")
	}

	var jsonData map[string]interface{}
	err := json.Unmarshal([]byte(msg[1]), &jsonData)
	if err != nil {
		return Mission{}, errors.New("json unmarshal failure [121]")
	}

	mission := Mission{
		User: User{
			ID:   jsonData["user_id"].(string),
			Name: jsonData["user_nick"].(string),
		},
		Title: jsonData["title"].(string),
	}

	switch v := jsonData["gift_count"].(type) {
	case float64:
		mission.Count = int(v)
	case int:
		mission.Count = v
	}

	return mission, nil
}
