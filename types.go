package soopchat

import (
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	callback

	Token Token

	channelPassword string
	socket          *websocket.Conn
	socketAddress   string // Socket Address
	read            chan []byte
	pingpongTimer   *time.Ticker
	handshake       [][]byte

	apiService apiService
}

type callback struct {
	// callback
	onError        func(err error)
	onConnect      func(isConnect bool)
	onJoinChannel  func(isJoin bool)
	onRawMessage   func(message string)
	onChatMessage  func(message ChatMessage)
	onUserLists    func(userlist []UserList)
	onAdballoon    func(adballoon Adballoon)
	onBalloon      func(balloon Balloon)
	onSubscription func(subscrption Subscription)
	onAdminNotice  func(message string)
	onMission      func(mission Mission)

	// api callback
	onLogin func(isLoginSuccess bool)
}

type Token struct {
	StreamerID string
	Identifier Identifier
	Flag       string

	authTicket string
	fanTicket  string
	chatRoom   string
}

type Identifier struct {
	ID       string
	Password string
}

type log struct {
	SetBps          string `json:"set_bps"`
	ViewBps         string `json:"view_bps"`
	Quality         string `json:"quality"`
	GeoContryCode   string `json:"geo_cc"`
	GeoRegionCode   string `json:"geo_rc"`
	AcceptLanguage  string `json:"acpt_lang"`
	ServiceLanguage string `json:"svc_lang"`
	JoinContryCode  string `json:"join_cc"`
	Subscribe       string `json:"subscribe"`
}

type info struct {
	Password string `json:"pwd"`
	AuthInfo string `json:"auth_info"`
	// PVer     string `json:"pver"`
	// AccessSystem string `json:"access_system"`
}

type User struct {
	ID             string // 유저 아이디
	Name           string // 유저 닉네임
	SubscribeMonth int    // 구독 개월
	Flag           UserFlag
}

type UserFlag struct {
	Flag1
	Flag2
}

type Flag1 struct {
	Admin          bool // 관리자
	Hidden         bool // 아이디 숨김
	BJ             bool // 방장
	Dumb           bool // 벙어리
	Guest          bool // 비회원
	Fanclub        bool // 팬클럽
	AutoManager    bool // 고정 매니저
	ManagerList    bool // 매니저 리스트
	Manager        bool // 매니저
	Female         bool // 여자 아니면 남자
	AutoDumb       bool // 자동 벙어리
	DumbBlind      bool // 벙어리 블라인드
	DobaeBlind     bool // 도배 블라인드
	DobaeBlind2    bool // 도배 블라인드 2
	ExitUser       bool // 나간 사람
	Mobile         bool // 모바일 유저
	TopFan         bool // 열혈
	Realname       bool // 실명인증
	NoDirect       bool // 1:1 직접 채팅 금지
	GlobalApp      bool // 글로벌 앱
	QuickView      bool // 퀵뷰 유저
	SptrSticker    bool // 스티커 서포터
	Chromecast     bool // 크롬 캐스트
	Subscriber     bool // 구독자
	NotiVodBalloon bool // VOD 별풍 알림
	NotiTopFan     bool // 열혈 알림
}

type Flag2 struct {
	GlobalPC    bool
	Clan        bool
	TopClan     bool
	Top20       bool
	GameGod     bool
	ATagAllow   bool
	NoSuperChat bool
	NoRecvChat  bool
	Flash       bool
	LGGame      bool
	Employee    bool
	CleanAti    bool
	Police      bool
	AdminChat   bool
	PC          bool
	Specify     bool
}

type ChatMessage struct {
	User    User
	Message string // 채팅
}

type UserList struct {
	User   User
	Status bool // 입장 true, 퇴장 false
}

type Balloon struct {
	User  User
	Count int // 별풍선 갯수
}

type Adballoon struct {
	User  User
	Count int // 애드벌룬 갯수
}

type Subscription struct {
	User  User
	Count int // 구독 개월 수
}

type Mission struct {
	User  User   // 유저
	Title string // 미션 이름
	Count int    // 미션 별풍선 갯수
}
