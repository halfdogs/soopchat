package soopchat

import (
	"fmt"
	"strconv"
	"strings"
)

// defaultLog 함수는 채팅 서버 연결에 필요한
// Handshake 데이터 중 미리 초기화된 Log 데이터를
// 반환한다.
func defaultLog() log {
	return log{
		SetBps:          "undefined",
		ViewBps:         "NaN",
		Quality:         "ori",
		GeoContryCode:   "undefined",
		GeoRegionCode:   "undefined",
		AcceptLanguage:  "undefined",
		ServiceLanguage: "undefined",
		JoinContryCode:  "410",
		Subscribe:       "1",
	}
}

// defaultInfo 함수는 채팅 서버 연결에 필요한
// Handshake 데이터 중 미리 초기화된 Info 데이터를
// 반환한다.
func defaultInfo(password string) info {
	return info{
		Password: password,
		AuthInfo: "undefined",
	}
}

// makeBuffer 함수는 전달된 데이터를
// 웹소켓으로 전달할 수 있는 []byte 데이터로 변환한다.
func makeBuffer(p []string) []byte {
	var buf []byte
	for _, e := range p {
		buf = append(buf, []byte(e)...)
	}

	return buf
}

// makeHeader 함수는 전달된 데이터를
// 웹소켓으로 전달할 수 있는 []byte 데이터로 변환한다.
func makeHeader(svc int, plen int, option int) []byte {
	var header []byte
	header = append(header, 27, 9)
	header = append(header, []byte(fmt.Sprintf("%04d", svc))...)
	header = append(header, []byte(fmt.Sprintf("%06d", plen))...)
	header = append(header, []byte(fmt.Sprintf("%02d", option))...)

	return header
}

// readInt 함수는 전달된 데이터를
// int 로 변환한다.
func readInt(message []byte) (int, error) {
	result := ""
	for _, b := range message {
		result += string(b)
	}

	return strconv.Atoi(result)
}

// getServiceCode 메서드는 전달된 데이터의
// 일부를 검사하여 서비스 코드를 반환한다.
func getServiceCode(message []byte) (int, error) {
	return readInt(message[2:6])
}

func parseMultiUserList(msg []string) []UserList {
	users := make([]UserList, 0)

	for i := 2; i < len(msg); i += 3 {
		if i+2 >= len(msg) || msg[i] == "-1" {
			continue
		}
		flags := strings.Split(msg[i+2], "|")
		userFlag := setFlag(flags)

		user := UserList{
			User: User{
				ID:   msg[i],
				Name: msg[i+1],
				Flag: userFlag,
			},
			Status: true,
		}

		users = append(users, user)
	}

	return users
}

func parseSingleUserList(msg []string) UserList {
	var status bool
	var userFlag UserFlag

	switch msg[1] {
	case "1":
		status = true
	case "-1":
		status = false
	}

	if status {
		flags := strings.Split(msg[4], "|")
		userFlag = setFlag(flags)
	}

	result := UserList{
		User: User{
			ID:   removeParentheses(msg[2]),
			Name: msg[3],
			Flag: userFlag,
		},
		Status: status,
	}

	return result
}

// removeParentheses 함수는 문자열에 포함되어 있는
// () 와 그 안의 있는 내용을 제거하여 반환합니다.
func removeParentheses(str string) string {
	idx := strings.Index(str, "(")
	if idx != -1 {
		return str[:idx]
	}

	return str
}

// setFlag 함수는 입력된 flag를 UserFlag로 변환하여
// 반환합니다.
func setFlag(flags []string) UserFlag {
	flag1, _ := strconv.Atoi(flags[0])
	flag2, _ := strconv.Atoi(flags[1])
	return UserFlag{
		Flag1: getFlag1(flag1),
		Flag2: getFlag2(flag2),
	}
}

// getFlag1 함수는 전달된 데이터의 값을 이용하여
// Flag1 구조체를 초기화하고 반환한다.
func getFlag1(flag int) Flag1 {
	return Flag1{
		Admin:          flag&(1<<0) != 0,
		Hidden:         flag&(1<<1) != 0,
		BJ:             flag&(1<<2) != 0,
		Dumb:           flag&(1<<3) != 0,
		Guest:          flag&(1<<4) != 0,
		Fanclub:        flag&(1<<5) != 0,
		AutoManager:    flag&(1<<6) != 0,
		ManagerList:    flag&(1<<7) != 0,
		Manager:        flag&(1<<8) != 0,
		Female:         flag&(1<<9) != 0,
		AutoDumb:       flag&(1<<10) != 0,
		DumbBlind:      flag&(1<<11) != 0,
		DobaeBlind:     flag&(1<<12) != 0,
		DobaeBlind2:    flag&(1<<24) != 0,
		ExitUser:       flag&(1<<13) != 0,
		Mobile:         flag&(1<<14) != 0,
		TopFan:         flag&(1<<15) != 0,
		Realname:       flag&(1<<16) != 0,
		NoDirect:       flag&(1<<17) != 0,
		GlobalApp:      flag&(1<<18) != 0,
		QuickView:      flag&(1<<19) != 0,
		SptrSticker:    flag&(1<<20) != 0,
		Chromecast:     flag&(1<<21) != 0,
		Subscriber:     flag&(1<<28) != 0,
		NotiVodBalloon: flag&(1<<30) != 0,
		NotiTopFan:     flag&(1<<31) != 0,
	}
}

// getFlag2 함수는 전달된 데이터의 값을 이용하여
// Flag2 구조체를 초기화하고 반환한다.
func getFlag2(flag int) Flag2 {
	return Flag2{
		GlobalPC:    flag&(1<<0) != 0,
		Clan:        flag&(1<<1) != 0,
		TopClan:     flag&(1<<2) != 0,
		Top20:       flag&(1<<3) != 0,
		GameGod:     flag&(1<<4) != 0,
		ATagAllow:   flag&(1<<5) != 0,
		NoSuperChat: flag&(1<<6) != 0,
		NoRecvChat:  flag&(1<<7) != 0,
		Flash:       flag&(1<<8) != 0,
		LGGame:      flag&(1<<9) != 0,
		Employee:    flag&(1<<10) != 0,
		CleanAti:    flag&(1<<11) != 0,
		Police:      flag&(1<<12) != 0,
		AdminChat:   flag&(1<<13) != 0,
		PC:          flag&(1<<14) != 0,
		Specify:     flag&(1<<15) != 0,
	}
}
