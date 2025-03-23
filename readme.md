# Soop chat, Go library
공부용으로 제작하였습니다.

[숲](https://sooplive.co.kr) 채팅을 읽을 수 있는 [Go](https://go.dev) 라이브러리입니다.

공식 API가 없기 때문에 임시로 제작하였으므로 언제든지 막히거나 수정될 가능성이 매우 높습니다.

아직 실력이 미숙하여 코드에 대한 부분들은 언제든지 이슈 및 PR로 남겨주시면 감사하겠습니다.

## 사용 방법
**라이브러리 가져오기**

`go get -u https://github.com/you0me0/soopchat`

**예제**
```go
...

func main() {
    token := soopchat.Token{
        StreamerID: {Streamer ID}, // 필수

        // 선택사항
        Identifier: soopchat.Identifier{
          ID:       {Your ID},
          Password: {Your PW},
        }
    }
    client, err := soopchat.NewClient(token)
    if err != nil {
        // client 생성 중 에러가 발생할 경우
        // panic() 을 호출합니다.
        panic(err)
    }

    client.OnError(func(err error) {
        // 에러가 발생할 경우 에러를 출력합니다.
        fmt.Println(err)
    })

    client.OnChatMessage(func(message soopchat.ChatMessage) {
        // 채팅 메시지가 있을 경우 ID, NAME, MESSAGE를 출력합니다.
        fmt.Printf("ID: %s, NAME: %s, MESSAGE: %s\n", message.User.ID, message.User.Name, message.Message)
    })

    // 채팅방 접속을 시도합니다.
    // 에러가 발생할 경우 panic()을 호출합니다.
    err := client.MustConnect()
    if err != nil {
        panic(err)
    }
}
```

### Token
- `StreamerID` **(필수)**
  - 원하는 채팅방에 입장하기 위해 반드시 필요한 BJ 아이디입니다.
- `Flag`
  - 채팅 채널 연결할 때 사용되는 유저 플래그 값
  - example: `524304`, `16`
- `Identifier`
  - 채팅 채널 연결을 할 때 로그인 데이터
  - `Identifier.ID` 와 `Identifier.Password`의 값이 있을 경우 자동으로 로그인을 진행합니다.
  - 입력하지 않을 경우 비로그인으로 채팅 채널에 연결합니다.

### Callback
- `OnError(error)`
  - 채팅 데이터를 읽는 중 에러가 발생할 경우 에러를 반환합니다.
- `OnConnect(bool)`
  - 채널 입장 Handshake가 성공하면 `true`를 반환합니다.
- `OnRawMessage(string)`
  - 원본 데이터 문자열을 반환합니다.
- `OnChatMessage(ChatMessage)`
  - 채팅 메시지가 있을 때마다 `ChatMessage` 구조체를 반환합니다.
- `OnUserLists([]UserList)`
  - 유저 입장/퇴장 메시지가 있을 때마다 `[]UserList` 구조체를 반환합니다.
- `OnBalloon(Balloon)`
  - 별풍선 메시지가 있을 때마다 `Balloon` 구조체를 반환합니다.
- `OnAdballoon(Adballoon)`
  - 애드벌룬 메시지가 있을 때마다 `Adballoon` 구조체를 반환합니다.
- `OnSubscription(Subscription)`
  - 구독 메시지가 있을 때마다 `Subscription` 구조체를 반환합니다.
- `OnAdminNotice(string)`
  - 운영자 알림 메시지가 있을 때마다 문자열을 반환합니다.
  - example: `"{BJ NAME}님의 방송이 별별랭킹의 '웃음이 끊이지 않는 방송' 1위에 등극!"`
- `OnMisson(Misson)`
  - 도전미션 별풍선 메시지가 있을 때마다 `Misson` 구조체를 반환합니다.
- `OnLogin(bool)`
  - 로그인 과정에서 성공/실패 여부를 반환합니다.

### 예제
- [Warudo](https://warudo.app)
  - [별풍선을 받을 때마다 OSC 통신 예제](https://github.com/you0me0/afreeca-warudo)

다른 프로그램(VSeeFace, Unity) 등 예제는 작성중입니다.


## TODO
아래에 작성된 순서는 개발 순서가 아닌 생각난 대로 작성하였습니다.

- 기능
  - [ ] 스티커를 받았을 때 콜백
  - [ ] 현재 채팅방에 있는 구독자 가져오기
  - [ ] 테스트 파일 작성
  - [ ] 코드 최적화
  - [ ] 에러 처리
    - [ ] 조금 더 자세한 에러 메시지 반환
    - [ ] `"\x1b\t000100005807\f시스템 에러가 발생 했습니다. (중복 세션)\f"`

## 버전 히스토리
- v0.2.0
  - 기존 `afreecachat`에서 `soopchat`으로 패키지 이름 및 깃허브 이름 변경

## 참고하면 좋은 라이브러리
- [https://github.com/wakscord/afreeca](https://github.com/wakscord/afreeca)
