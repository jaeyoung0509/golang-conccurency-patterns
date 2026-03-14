---
title: Go 함정 부록
description: 컴파일은 되지만 프로덕션에서 문제를 만드는 Go 함정 50가지를 한 페이지에 모은 부록입니다.
---

# Go 함정 부록

이 부록은 다음 같은 코드를 다룹니다.

- 컴파일은 되고
- 대충 보면 리뷰도 통과하고
- 운영에 들어가면 크게 문제를 일으키는 코드

목표는 문법 입문이 아닙니다. 숙련된 Go 엔지니어도 일정 압박 아래에서 자주 만드는 실수를 더 날카롭게 구분하는 데 있습니다.

## 무엇을 다루나

| 범주 | 초점 |
| --- | --- |
| [Control Flow와 Evaluation](#control-flow-and-evaluation) | `select`, `break`, loop variable, `defer`, 평가 시점처럼 겉보기엔 단순하지만 자주 틀리는 지점 |
| [Interfaces와 Types](#interfaces-and-types) | typed nil, type assertion, method set, wrapped error, `any`가 만드는 모호함 |
| [Collections와 Memory](#collections-and-memory) | map/slice 함정, iteration 가정, backing array retention, 메모리 형태 관련 놀라움 |
| [Concurrency, Context, Time](#concurrency-context-and-time) | channel ownership, goroutine lifetime, cancellation, timer, `WaitGroup` 순서 |
| [Stdlib과 API Boundary](#stdlib-and-api-boundaries) | `net/http`, `database/sql`, `io`, `json`, `time`, `os/exec`, `sync.Map` 경계에서 생기는 footgun |

## 이 페이지를 읽는 법

모든 함정은 같은 형식을 따릅니다.

1. 나쁜 코드
2. 왜 깨지거나 오해를 부르는지
3. 더 나은 코드
4. 어떻게 빨리 잡는지
5. 한 줄 규칙

언어 튜토리얼처럼 읽기보다, 코드 리뷰 체크리스트처럼 읽는 편이 맞습니다.

## 더 깊게 보고 싶다면

이 부록은 의도적으로 짧고 날카롭게 썼습니다. 더 큰 런타임/운영 주제로 이어지는 항목은 이미 사이트의 다른 문서에서 깊게 다룹니다.

- [Race Detector](/ko/testing/race-detector)
- [synctest로 결정적 테스트](/ko/testing/synctest)
- [context 패키지 내부](/ko/stdlib/context-internals)
- [time, Timers, Tickers](/ko/stdlib/time-timers-tickers)
- [net/http 서버와 Transport 내부](/ko/stdlib/net-http-server-transport)
- [sync와 atomic 프리미티브](/ko/stdlib/sync-and-atomic)
- [Channels, Select, 그리고 Memory Model](/ko/fundamentals/channels-memory-model)

## 현대 Go 기준으로 한 가지 주의점

Go 함정을 다루는 오래된 글과 발표 중 일부는 이제 완전히 맞지는 않습니다.

대표적으로 loop variable capture는 Go 1.22에서 흔한 `for ... := range ...` 패턴이 개선됐습니다. 이 부록은 현대 Go 동작을 기준으로 쓰되, 오래된 코드나 predeclared loop variable에서 여전히 남는 함정은 따로 짚습니다.

## 범주별로 바로 이동하기

- [Control Flow와 Evaluation](#control-flow-and-evaluation)
- [Interfaces와 Types](#interfaces-and-types)
- [Collections와 Memory](#collections-and-memory)
- [Concurrency, Context, Time](#concurrency-context-and-time)
- [Stdlib과 API Boundary](#stdlib-and-api-boundaries)

<a id="control-flow-and-evaluation"></a>

## Control Flow와 Evaluation

#### 1. `select`에 `default`를 넣어 의도치 않은 busy-spin이 생김

**Bad**

```go
for {
	select {
	case job := <-jobs:
		handle(job)
	default:
		continue
	}
}
```

왜 문제인가: 이 루프는 절대 block되지 않습니다. `jobs`가 비면 일을 기다리는 대신 CPU만 태웁니다.

**Better**

```go
for {
	select {
	case job, ok := <-jobs:
		if !ok {
			return
		}
		handle(job)
	case <-ctx.Done():
		return
	}
}
```

어떻게 빨리 잡나: CPU profile, execution trace, review.

Rule: 오래 사는 `select`는 정말 polling이 필요할 때가 아니라면 block되어야 합니다.

#### 2. nil channel branch가 `select` case를 조용히 비활성화함

**Bad**

```go
var updates <-chan Event

if cfg.LiveReload {
	updates = watcher.Events()
}

select {
case ev := <-updates:
	apply(ev)
case <-ctx.Done():
	return ctx.Err()
}
```

왜 문제인가: nil channel case는 절대 실행되지 않습니다. 의도적으로 쓸 수도 있지만, 한 branch가 영구적으로 꺼졌다는 사실을 놓치기 쉽습니다.

**Better**

```go
if !cfg.LiveReload {
	<-ctx.Done()
	return ctx.Err()
}

updates := watcher.Events()
select {
case ev := <-updates:
	apply(ev)
case <-ctx.Done():
	return ctx.Err()
}
```

어떻게 빨리 잡나: focused test, review.

Rule: nil channel로 case를 끄는 패턴은 코드에서 그 의도가 명확히 보여야 합니다.

#### 3. 닫힌 channel에서 받은 zero value를 실제 값으로 오해함

**Bad**

```go
for {
	id := <-ids
	if id == "" {
		return nil
	}
	process(id)
}
```

왜 문제인가: 닫힌 channel에서 receive하면 zero value가 즉시 반환됩니다. 여기서는 `""`가 channel 종료인지 실제 빈 값인지 구분되지 않습니다.

**Better**

```go
for {
	id, ok := <-ids
	if !ok {
		return nil
	}
	process(id)
}
```

어떻게 빨리 잡나: test, review.

Rule: channel close가 의미를 가지면 항상 `ok` 플래그를 같이 받으십시오.

#### 4. 닫힌 channel로 send함

**Bad**

```go
func consume(work <-chan Job) {
	for job := range work {
		handle(job)
	}
}

func stop(work chan Job) {
	close(work)
}
```

왜 문제인가: `stop` 이후에도 `work`로 보내는 쪽이 있으면 panic이 납니다.

**Better**

```go
func produce(done <-chan struct{}, work chan<- Job, job Job) error {
	select {
	case work <- job:
		return nil
	case <-done:
		return context.Canceled
	}
}
```

어떻게 빨리 잡나: test, review.

Rule: 데이터 channel을 닫는 책임은 보통 send하는 쪽이 가져야 합니다.

#### 5. `select`가 공정할 거라고 가정함

**Bad**

```go
for {
	select {
	case job := <-hotQueue:
		handle(job)
	case <-ticker.C:
		flushMetrics()
	}
}
```

왜 문제인가: `select`는 사람들이 기대하는 식의 공정성을 보장하지 않습니다. 뜨거운 case가 덜 바쁜 case를 계속 밀어낼 수 있습니다.

**Better**

```go
go func() {
	for range ticker.C {
		flushMetrics()
	}
}()

for job := range hotQueue {
	handle(job)
}
```

어떻게 빨리 잡나: load test, trace, review.

Rule: 어떤 branch가 독립적으로 반드시 전진해야 한다면 별도 owner를 두십시오.

#### 6. `select`나 `switch` 안의 `break`가 바깥 loop를 끝내지 못함

**Bad**

```go
for {
	select {
	case <-ctx.Done():
		break
	case job := <-jobs:
		handle(job)
	}
}
cleanup()
```

왜 문제인가: `break`는 `select`만 빠져나오고 바깥 `for`는 계속 돕니다.

**Better**

```go
loop:
for {
	select {
	case <-ctx.Done():
		break loop
	case job := <-jobs:
		handle(job)
	}
}
cleanup()
```

어떻게 빨리 잡나: test, review.

Rule: 바깥 loop를 끝낼 생각이면 `return`이나 labeled `break`를 쓰십시오.

#### 7. goroutine에서 loop variable을 capture함

**Bad**

```go
var id int

for _, id = range ids {
	go func() {
		log.Println(id)
	}()
}
```

왜 문제인가: goroutine이 모두 같은 `id` 변수를 캡처합니다. 현대 Go에서는 흔한 `for _, id := range ids` 패턴이 개선됐지만, predeclared variable이나 오래된 코드에서는 여전히 함정입니다.

**Better**

```go
var id int

for _, id = range ids {
	id := id
	go func() {
		log.Println(id)
	}()
}
```

어떻게 빨리 잡나: test, review.

Rule: goroutine이 loop variable을 닫아버리면 iteration별 바인딩을 명시적으로 만드십시오.

#### 8. `defer`에서 loop variable을 capture함

**Bad**

```go
var name string

for _, name = range files {
	defer func() {
		log.Println("closing", name)
	}()
}
```

왜 문제인가: 모든 deferred closure가 같은 변수를 보기 때문에 마지막 값만 찍습니다.

**Better**

```go
var name string

for _, name = range files {
	name := name
	defer func() {
		log.Println("closing", name)
	}()
}
```

어떻게 빨리 잡나: test, review.

Rule: `defer`는 값을 snapshot하지 않고 변수를 캡처합니다. 필요하면 직접 snapshot을 만드십시오.

#### 9. `:=`로 `err`를 shadowing함

**Bad**

```go
func run() error {
	err := start()
	if err != nil {
		return err
	}

	if data, err := fetch(); err != nil {
		log.Println("fetch failed:", err)
	} else {
		use(data)
	}

	return err
}
```

왜 문제인가: `if` 안의 `err`는 바깥 `err`와 다른 변수입니다. 함수는 여전히 `nil`인 바깥 `err`를 반환합니다.

**Better**

```go
func run() error {
	if err := start(); err != nil {
		return err
	}

	data, err := fetch()
	if err != nil {
		return err
	}
	use(data)
	return nil
}
```

어떻게 빨리 잡나: test, review.

Rule: block 밖에서도 의미 있는 에러라면 short declaration으로 숨기지 마십시오.

#### 10. loop 안의 `defer`가 리소스를 너무 오래 붙잡음

**Bad**

```go
for _, name := range names {
	f, err := os.Open(name)
	if err != nil {
		return err
	}
	defer f.Close()
	scan(f)
}
```

왜 문제인가: 모든 파일이 함수가 끝날 때까지 열린 채로 남습니다. loop가 길면 file descriptor를 다 써버릴 수 있습니다.

**Better**

```go
for _, name := range names {
	f, err := os.Open(name)
	if err != nil {
		return err
	}
	scan(f)
	if err := f.Close(); err != nil {
		return err
	}
}
```

어떻게 빨리 잡나: load가 있는 test, review.

Rule: `defer`는 loop iteration scope가 아니라 function scope입니다.

<a id="interfaces-and-types"></a>

## Interfaces와 Types

#### 11. interface 안의 typed nil이 `nil`과 같지 않음

**Bad**

```go
func maybeReader() io.Reader {
	var f *os.File
	return f
}

r := maybeReader()
if r == nil {
	return
}
```

왜 문제인가: interface 값 안에는 concrete type `*os.File`이 들어 있으므로, 내부 포인터가 nil이어도 interface 자체는 nil이 아닙니다.

**Better**

```go
func maybeReader() io.Reader {
	var f *os.File
	if f == nil {
		return nil
	}
	return f
}
```

어떻게 빨리 잡나: test, review.

Rule: interface가 nil이려면 type과 value가 둘 다 nil이어야 합니다.

#### 12. concrete value가 comparable하지 않은 interface끼리 비교함

**Bad**

```go
var left any = []int{1, 2}
var right any = []int{1, 2}

if left == right {
	log.Println("equal")
}
```

왜 문제인가: slice는 comparable하지 않으므로 비교 시 panic이 납니다. interface 비교는 concrete type의 comparability 규칙을 그대로 따릅니다.

**Better**

```go
left := []int{1, 2}
right := []int{1, 2}

if slices.Equal(left, right) {
	log.Println("equal")
}
```

어떻게 빨리 잡나: test, review.

Rule: 박싱된 interface끼리 비교하지 말고 concrete type에 맞는 helper로 비교하십시오.

#### 13. pointer receiver와 value receiver를 헷갈려 interface 구현을 잘못 이해함

**Bad**

```go
type Writer struct{}

func (*Writer) Write(p []byte) (int, error) {
	return len(p), nil
}

var _ io.Writer = Writer{}
```

왜 문제인가: `Writer{}`는 `io.Writer`를 구현하지 못하고 `*Writer`만 구현합니다. 이건 컴파일 타임에 바로 깨집니다.

**Better**

```go
type Writer struct{}

func (*Writer) Write(p []byte) (int, error) {
	return len(p), nil
}

var _ io.Writer = (*Writer)(nil)
```

어떻게 빨리 잡나: compiler, compile-time interface assertion.

Rule: method가 pointer receiver라면 interface를 만족하는 것도 pointer type입니다.

#### 14. nil receiver method를 허용해 놓고 내부에서는 바로 panic남

**Bad**

```go
type Config struct {
	Timeout time.Duration
}

func (c *Config) EffectiveTimeout() time.Duration {
	return c.Timeout
}
```

왜 문제인가: nil receiver에 대한 method dispatch 자체는 합법이지만, 내부에서 field를 바로 dereference하면 panic이 납니다.

**Better**

```go
func (c *Config) EffectiveTimeout() time.Duration {
	if c == nil {
		return 5 * time.Second
	}
	return c.Timeout
}
```

어떻게 빨리 잡나: test, review.

Rule: nil receiver를 허용할 생각이라면 method 안에서 그 상태를 명시적으로 처리하십시오.

#### 15. implicit interface satisfaction을 확인하지 않고 가정만 함

**Bad**

```go
type Saver interface {
	Save(context.Context, string) error
}

type Store struct{}

func (Store) Save(context.Context, string) error { return nil }
```

왜 문제인가: 지금은 맞아 보이지만, 나중에 refactor로 method set이 미묘하게 바뀌면 먼 wiring 지점에서야 깨질 수 있습니다.

**Better**

```go
type Saver interface {
	Save(context.Context, string) error
}

type Store struct{}

func (Store) Save(context.Context, string) error { return nil }

var _ Saver = Store{}
```

어떻게 빨리 잡나: compiler, compile-time interface assertion.

Rule: interface 만족이 설계의 일부라면 타입 옆에서 바로 assert하십시오.

#### 16. type assertion을 검사 없이 씀

**Bad**

```go
func handle(msg any) {
	order := msg.(OrderPlaced)
	process(order)
}
```

왜 문제인가: 예상 밖의 입력 하나가 panic으로 바로 이어집니다.

**Better**

```go
func handle(msg any) error {
	order, ok := msg.(OrderPlaced)
	if !ok {
		return fmt.Errorf("unexpected message type %T", msg)
	}
	process(order)
	return nil
}
```

어떻게 빨리 잡나: test, review.

Rule: panic이 진짜 정책이 아닌 이상 one-value assertion을 쓰지 마십시오.

#### 17. wrapped error를 `==`로 비교함

**Bad**

```go
if err := load(); err != nil {
	if err == sql.ErrNoRows {
		return nil
	}
	return err
}
```

왜 문제인가: 에러가 wrap되면 direct equality는 기대대로 동작하지 않습니다.

**Better**

```go
if err := load(); err != nil {
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	return err
}
```

어떻게 빨리 잡나: test, review.

Rule: wrap될 수 있는 에러는 값이 아니라 의미로 비교하십시오.

#### 18. `any`가 도메인 의미까지 보존해 줄 거라고 생각함

**Bad**

```go
func setLimit(v any) int {
	limit := v.(int)
	if limit < 0 {
		panic("negative")
	}
	return limit
}
```

왜 문제인가: `any`는 도메인 의도를 보존하지 않습니다. `int64(10)`이나 `json.Number("10")`이 들어오면 런타임 문제로 바뀝니다.

**Better**

```go
func setLimit(limit int) (int, error) {
	if limit < 0 {
		return 0, errors.New("negative limit")
	}
	return limit, nil
}
```

어떻게 빨리 잡나: compiler, test, review.

Rule: 도메인이 이미 정해져 있다면 `any`보다 typed input을 쓰십시오.

#### 19. `json.Unmarshal`을 `interface{}`에 넣고 숫자가 `float64`가 되는 걸 놓침

**Bad**

```go
var payload map[string]any
_ = json.Unmarshal([]byte(`{"retries":3}`), &payload)

retries := payload["retries"].(int)
fmt.Println(retries)
```

왜 문제인가: generic JSON decode는 기본적으로 숫자를 `float64`로 넣기 때문에 `int` assertion이 panic납니다.

**Better**

```go
type Payload struct {
	Retries int `json:"retries"`
}

var payload Payload
_ = json.Unmarshal([]byte(`{"retries":3}`), &payload)
fmt.Println(payload.Retries)
```

어떻게 빨리 잡나: test, review.

Rule: shape를 알고 있다면 typed struct로 decode하십시오.

#### 20. interface field가 zero value와 absent state를 흐림

**Bad**

```go
type Query struct {
	Limit any `json:"limit"`
}

var q Query
_ = json.Unmarshal(body, &q)
```

왜 문제인가: `nil`, `0`, `0.0`, 잘못된 concrete type이 모두 `any` 하나 안으로 뭉개집니다. validation이 늦고 지저분해집니다.

**Better**

```go
type Query struct {
	Limit *int `json:"limit"`
}

var q Query
_ = json.Unmarshal(body, &q)
```

어떻게 빨리 잡나: test, review.

Rule: typed field를 쓰고, "없음"과 0이 다르면 pointer로 표현하십시오.

<a id="collections-and-memory"></a>

## Collections와 Memory

#### 21. nil map에 write함

**Bad**

```go
func record(hit map[string]int, key string) {
	hit[key]++
}

func main() {
	var hit map[string]int
	record(hit, "cache")
}
```

왜 문제인가: nil map은 읽기는 되지만 쓰기는 panic입니다.

**Better**

```go
func record(hit map[string]int, key string) {
	if hit == nil {
		panic("record called with nil map")
	}
	hit[key]++
}
```

어떻게 빨리 잡나: test, review.

Rule: writable map이라면 넘겨주기 전에 반드시 초기화하십시오.

#### 22. nil slice와 empty slice를 API나 JSON에서 같다고 가정함

**Bad**

```go
type Response struct {
	Items []string `json:"items"`
}

func emptyResponse() Response {
	return Response{}
}
```

왜 문제인가: Go 코드에서는 비슷하게 보이지만 JSON에서는 `null`과 `[]`로 다르게 encode됩니다.

**Better**

```go
type Response struct {
	Items []string `json:"items"`
}

func emptyResponse() Response {
	return Response{Items: []string{}}
}
```

어떻게 빨리 잡나: JSON test, API contract review.

Rule: 클라이언트가 빈 리스트를 기대하면 nil 대신 empty slice를 반환하십시오.

#### 23. nil map과 empty map을 API나 JSON에서 같다고 가정함

**Bad**

```go
type Response struct {
	Labels map[string]string `json:"labels"`
}

func emptyResponse() Response {
	return Response{}
}
```

왜 문제인가: nil map은 `null`, 빈 map은 `{}`로 marshal됩니다.

**Better**

```go
type Response struct {
	Labels map[string]string `json:"labels"`
}

func emptyResponse() Response {
	return Response{Labels: map[string]string{}}
}
```

어떻게 빨리 잡나: JSON test, API contract review.

Rule: caller가 object를 기대하면 object를 초기화해서 넘기십시오.

#### 24. shared backing array 때문에 `append`가 서로 엉킴

**Bad**

```go
base := make([]int, 0, 4)

left := append(base, 1, 2)
right := append(base, 9)

fmt.Println(left, right)
```

왜 문제인가: `left`와 `right`가 같은 backing array를 공유할 수 있습니다. 한쪽 append가 다른 쪽 내용을 덮어쓸 수 있습니다.

**Better**

```go
base := []int{1, 2}

left := append([]int(nil), base...)
right := append([]int(nil), base...)

left = append(left, 3)
right = append(right, 9)
```

어떻게 빨리 잡나: test, review.

Rule: 두 slice가 독립적으로 자라야 한다면 분기 전에 복사하십시오.

#### 25. subslice가 거대한 backing array를 붙잡아 둠

**Bad**

```go
func readHeader() []byte {
	buf := make([]byte, 10<<20)
	fill(buf)
	return buf[:32]
}
```

왜 문제인가: 반환하는 건 32바이트지만 10MB backing array 전체가 살아남습니다.

**Better**

```go
func readHeader() []byte {
	buf := make([]byte, 10<<20)
	fill(buf)

	header := make([]byte, 32)
	copy(header, buf[:32])
	return header
}
```

어떻게 빨리 잡나: heap profile, review.

Rule: 큰 backing array를 버리고 싶다면 필요한 작은 조각을 복사해서 분리하십시오.

#### 26. `range` variable의 주소를 잡음

**Bad**

```go
var user User
var ptrs []*User

for _, user = range users {
	ptrs = append(ptrs, &user)
}
```

왜 문제인가: 모든 pointer가 같은 loop variable을 가리킵니다. 현대 Go에서는 흔한 `:= range` 패턴이 개선됐지만, predeclared variable과 오래된 코드는 여전히 함정입니다.

**Better**

```go
var ptrs []*User

for i := range users {
	ptrs = append(ptrs, &users[i])
}
```

어떻게 빨리 잡나: test, review.

Rule: element 주소가 필요하면 loop variable이 아니라 slice element의 주소를 잡으십시오.

#### 27. slice를 `range`하면서 동시에 수정함

**Bad**

```go
for i, id := range ids {
	if shouldDrop(id) {
		ids = append(ids[:i], ids[i+1:]...)
	}
}
```

왜 문제인가: `range`는 원래 slice header를 기준으로 돕니다. 중간 삭제를 하면 항목을 건너뛰거나 stale position을 읽을 수 있습니다.

**Better**

```go
filtered := ids[:0]

for _, id := range ids {
	if !shouldDrop(id) {
		filtered = append(filtered, id)
	}
}
ids = filtered
```

어떻게 빨리 잡나: test, review.

Rule: 순회 중인 slice를 직접 바꾸지 말고 다음 slice를 명시적으로 만드십시오.

#### 28. map iteration order가 안정적이라고 믿음

**Bad**

```go
func pickOne(m map[string]Route) Route {
	for _, route := range m {
		return route
	}
	panic("empty")
}
```

왜 문제인가: map iteration 순서는 의도적으로 안정적이지 않습니다. "첫 번째 원소"는 계약이 아닙니다.

**Better**

```go
func pickOne(keys []string, m map[string]Route) Route {
	sort.Strings(keys)
	return m[keys[0]]
}
```

어떻게 빨리 잡나: 여러 번 도는 test, review.

Rule: 순서가 중요하면 직접 순서를 만들어야 합니다.

#### 29. 동기화 없이 map을 concurrent access함

**Bad**

```go
cache := map[string]int{}

go func() { cache["a"] = 1 }()
go func() { _ = cache["a"] }()

time.Sleep(10 * time.Millisecond)
```

왜 문제인가: 일반 Go map은 동기화 없는 concurrent read/write에 안전하지 않습니다.

**Better**

```go
var (
	mu    sync.Mutex
	cache = map[string]int{}
)

mu.Lock()
cache["a"] = 1
mu.Unlock()
```

어떻게 빨리 잡나: `-race`, test, review.

Rule: plain map은 ownership이나 synchronization이 필요합니다.

#### 30. `string`과 `[]byte` 변환이 공짜이거나 storage를 공유한다고 착각함

**Bad**

```go
s := "prod"
b := []byte(s)

b[0] = 'P'
fmt.Println(s)
```

왜 문제인가: string에서 `[]byte`로 변환하면 복사가 일어납니다. `b`를 바꿔도 `s`는 바뀌지 않고, hot path에서 반복 변환하면 allocation도 생깁니다.

**Better**

```go
b := []byte("prod")
b[0] = 'P'

s := string(b)
fmt.Println(s)
```

어떻게 빨리 잡나: benchmark, alloc profile, review.

Rule: storage를 공유할 거라고 기대하지 말고, 표현을 바꿔야 할 때만 변환하십시오.

<a id="concurrency-context-and-time"></a>

## Concurrency, Context, Time

#### 31. receiver 쪽에서 channel을 닫음

**Bad**

```go
func consume(work chan Job) {
	job := <-work
	handle(job)
	close(work)
}

func produce(work chan<- Job, job Job) {
	work <- job
}
```

왜 문제인가: 다음 sender가 panic납니다. receiver는 보통 send lifecycle의 owner가 아닙니다.

**Better**

```go
func produce(done <-chan struct{}, work chan<- Job, job Job) error {
	select {
	case work <- job:
		return nil
	case <-done:
		return context.Canceled
	}
}
```

어떻게 빨리 잡나: test, review.

Rule: 데이터 channel을 닫는 책임은 send하는 쪽에 두십시오.

#### 32. 여러 sender가 같은 channel을 닫으려 경쟁함

**Bad**

```go
func stop(done chan struct{}) {
	close(done)
}

go stop(done)
go stop(done)
```

왜 문제인가: close는 한 번만 성공합니다. 두 번째 close는 panic입니다.

**Better**

```go
var once sync.Once

func stop(done chan struct{}) {
	once.Do(func() {
		close(done)
	})
}
```

어떻게 빨리 잡나: test, review.

Rule: 여러 goroutine이 같은 signal channel을 끌 수 있다면 close를 직렬화하십시오.

#### 33. `len(ch)`를 coordination signal로 사용함

**Bad**

```go
if len(jobs) == 0 {
	return ErrIdle
}

job := <-jobs
handle(job)
```

왜 문제인가: `len(ch)`는 그 순간의 snapshot일 뿐입니다. 바로 다음 순간 다른 goroutine이 channel 상태를 바꿀 수 있습니다.

**Better**

```go
select {
case job := <-jobs:
	handle(job)
default:
	return ErrIdle
}
```

어떻게 빨리 잡나: test, review.

Rule: channel coordination은 `len`이 아니라 실제 channel operation으로 하십시오.

#### 34. caller timeout 이후 reply send가 막혀 goroutine leak이 남음

**Bad**

```go
func request(ctx context.Context) error {
	reply := make(chan Result)

	go func() {
		reply <- slowQuery()
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-reply:
		return nil
	}
}
```

왜 문제인가: caller가 먼저 timeout되면 goroutine이 `reply <-`에서 영원히 block될 수 있습니다.

**Better**

```go
func request(ctx context.Context) error {
	reply := make(chan Result, 1)

	go func() {
		reply <- slowQuery()
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-reply:
		return nil
	}
}
```

어떻게 빨리 잡나: leak test, shutdown test, review.

Rule: caller가 먼저 떠나도 reply path는 반드시 종료할 수 있어야 합니다.

#### 35. owner나 stop condition 없이 goroutine을 시작함

**Bad**

```go
func StartMetricsLoop() {
	go func() {
		for range time.Tick(time.Second) {
			publishMetrics()
		}
	}()
}
```

왜 문제인가: 이 goroutine에는 owner도 없고 shutdown path도 없고 언제 끝나는지 계약도 없습니다.

**Better**

```go
func StartMetricsLoop(ctx context.Context) {
	ticker := time.NewTicker(time.Second)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				publishMetrics()
			case <-ctx.Done():
				return
			}
		}
	}()
}
```

어떻게 빨리 잡나: lifecycle test, review.

Rule: 모든 goroutine에는 owner, stop condition, wait strategy가 있어야 합니다.

#### 36. request-scoped work에서 `context.Background()`를 사용함

**Bad**

```go
func handleRequest(ctx context.Context, id string) error {
	go auditLog(context.Background(), id)
	return fetch(ctx, id)
}
```

왜 문제인가: background 작업이 request lifetime에서 이탈합니다. timeout과 cancellation이 더 이상 통제하지 못합니다.

**Better**

```go
func handleRequest(ctx context.Context, id string) error {
	go auditLog(ctx, id)
	return fetch(ctx, id)
}
```

어떻게 빨리 잡나: test, trace, review.

Rule: request 범위의 작업은 의도적으로 분리하는 게 아니라면 request context를 상속해야 합니다.

#### 37. `cancel` 호출을 잊음

**Bad**

```go
func fetchAll(parent context.Context) error {
	ctx, _ := context.WithTimeout(parent, 2*time.Second)
	return callBackend(ctx)
}
```

왜 문제인가: 작업이 빨리 끝나도 deadline이 만료될 때까지 timer와 child context 자원이 남아 있습니다.

**Better**

```go
func fetchAll(parent context.Context) error {
	ctx, cancel := context.WithTimeout(parent, 2*time.Second)
	defer cancel()
	return callBackend(ctx)
}
```

어떻게 빨리 잡나: review, test.

Rule: cancelable context를 만들었다면 cancel function도 호출하십시오.

#### 38. loop 안의 `time.After`가 timer churn을 만듦

**Bad**

```go
for {
	select {
	case msg := <-msgs:
		handle(msg)
	case <-time.After(time.Second):
		reportIdle()
	}
}
```

왜 문제인가: loop를 돌 때마다 새 timer가 생깁니다. load가 걸리면 불필요한 churn이 커집니다.

**Better**

```go
timer := time.NewTimer(time.Second)
defer timer.Stop()

for {
	timer.Reset(time.Second)
	select {
	case msg := <-msgs:
		handle(msg)
	case <-timer.C:
		reportIdle()
	}
}
```

어떻게 빨리 잡나: alloc profile, review.

Rule: 오래 사는 loop는 보통 timer를 재사용해야 합니다.

#### 39. `Ticker`를 만들고 `Stop`하지 않음

**Bad**

```go
func start() {
	ticker := time.NewTicker(time.Minute)
	go func() {
		for range ticker.C {
			refreshCache()
		}
	}()
}
```

왜 문제인가: ticker는 stop하기 전까지 계속 tick합니다. owner가 죽거나 loop가 끝나도 cleanup 책임은 여전히 남습니다.

**Better**

```go
func start(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				refreshCache()
			case <-ctx.Done():
				return
			}
		}
	}()
}
```

어떻게 빨리 잡나: test, review.

Rule: ticker를 만든 쪽이 누가 언제 멈추는지도 같이 정의해야 합니다.

#### 40. goroutine 시작 후에 `WaitGroup.Add`를 호출함

**Bad**

```go
var wg sync.WaitGroup

go func() {
	defer wg.Done()
	work()
}()
wg.Add(1)
wg.Wait()
```

왜 문제인가: `Wait`와 `Add`가 경쟁할 수 있고, counter가 올라가기 전에 `Done`이 실행될 수도 있습니다.

**Better**

```go
var wg sync.WaitGroup

wg.Add(1)
go func() {
	defer wg.Done()
	work()
}()
wg.Wait()
```

어떻게 빨리 잡나: test, review.

Rule: goroutine이 runnable해지기 전에 wait-group count를 먼저 올리십시오.

<a id="stdlib-and-api-boundaries"></a>

## Stdlib과 API Boundary

#### 41. 요청마다 새로운 `http.Client`나 `Transport`를 만듦

**Bad**

```go
func fetch(ctx context.Context, url string) error {
	client := &http.Client{
		Transport: &http.Transport{},
	}
	_, err := client.Get(url)
	return err
}
```

왜 문제인가: connection pool과 transport reuse가 사라집니다. handshake와 socket 비용을 매번 다시 냅니다.

**Better**

```go
var client = &http.Client{
	Timeout: 5 * time.Second,
	Transport: &http.Transport{
		MaxIdleConnsPerHost: 16,
	},
}
```

어떻게 빨리 잡나: load test, metrics, review.

Rule: `http.Client`, 특히 `Transport`는 일회용 helper가 아니라 오래 사는 policy object입니다.

#### 42. `http.Response.Body`를 drain하지도 close하지도 않음

**Bad**

```go
resp, err := client.Do(req)
if err != nil {
	return err
}
if resp.StatusCode != http.StatusOK {
	return fmt.Errorf("bad status: %s", resp.Status)
}
```

왜 문제인가: body가 열린 채로 남고, transport가 connection을 재사용하지 못할 수도 있습니다.

**Better**

```go
resp, err := client.Do(req)
if err != nil {
	return err
}
defer resp.Body.Close()
_, _ = io.Copy(io.Discard, resp.Body)
```

어떻게 빨리 잡나: test, transport metrics, review.

Rule: response body를 받았으면 닫고, reuse가 중요하면 drain도 하십시오.

#### 43. `io.Reader`가 한 번에 버퍼를 다 채워 줄 거라고 생각함

**Bad**

```go
buf := make([]byte, 32)
n, err := r.Read(buf)
if err != nil {
	return err
}
return decode(buf[:n])
```

왜 문제인가: `Read`는 요청한 길이보다 적게 읽고도 완전히 정상일 수 있습니다.

**Better**

```go
buf := make([]byte, 32)
if _, err := io.ReadFull(r, buf); err != nil {
	return err
}
return decode(buf)
```

어떻게 빨리 잡나: short reader를 쓰는 test, review.

Rule: `io.Reader`는 full-buffer delivery가 아니라 byte stream만 약속합니다.

#### 44. `database/sql`에서 `Rows.Close`와 `Rows.Err`를 빼먹음

**Bad**

```go
rows, err := db.QueryContext(ctx, q)
if err != nil {
	return err
}
for rows.Next() {
	if err := rows.Scan(&item.ID); err != nil {
		return err
	}
}
```

왜 문제인가: rows를 닫지 않으면 리소스를 누수할 수 있고, `Rows.Err`를 확인하지 않으면 iteration 중 생긴 에러를 놓칠 수 있습니다.

**Better**

```go
rows, err := db.QueryContext(ctx, q)
if err != nil {
	return err
}
defer rows.Close()

for rows.Next() {
	if err := rows.Scan(&item.ID); err != nil {
		return err
	}
}
return rows.Err()
```

어떻게 빨리 잡나: integration test, review.

Rule: `Rows`를 다룰 때는 항상 `Query`, `Close`, `Err`를 같이 생각하십시오.

#### 45. `QueryRow`나 `Exec`가 맞는 자리에 `Query`를 씀

**Bad**

```go
rows, err := db.QueryContext(ctx, `
	INSERT INTO audit_log(actor, action) VALUES (?, ?)
`, actor, action)
if err != nil {
	return err
}
defer rows.Close()
```

왜 문제인가: `Query`는 result-set handling을 전제로 합니다. row를 반환하지 않는 statement에서 contract가 불필요하게 시끄럽고 오용되기 쉽습니다.

**Better**

```go
_, err := db.ExecContext(ctx, `
	INSERT INTO audit_log(actor, action) VALUES (?, ?)
`, actor, action)
return err
```

어떻게 빨리 잡나: review.

Rule: `database/sql`에서는 결과 형태에 가장 좁게 맞는 API를 고르십시오.

#### 46. strict input을 원했는데 `json.Decoder`가 unknown field를 조용히 무시함

**Bad**

```go
var req CreateUserRequest

dec := json.NewDecoder(r.Body)
if err := dec.Decode(&req); err != nil {
	return err
}
```

왜 문제인가: 오타가 있는 field나 예상하지 못한 field가 기본적으로 무시됩니다. 클라이언트는 값을 보냈다고 생각하지만 서버는 쓰지 않았을 수 있습니다.

**Better**

```go
var req CreateUserRequest

dec := json.NewDecoder(r.Body)
dec.DisallowUnknownFields()
if err := dec.Decode(&req); err != nil {
	return err
}
```

어떻게 빨리 잡나: request validation test, review.

Rule: 입력 계약이 엄격해야 한다면 decoder도 엄격하게 만드십시오.

#### 47. reference time을 쓰지 않아 `time.Parse` layout을 틀림

**Bad**

```go
ts, err := time.Parse("YYYY-MM-DD", rawDate)
if err != nil {
	return err
}
use(ts)
```

왜 문제인가: Go layout은 기호형 format string이 아닙니다. reference time으로 써야 합니다.

**Better**

```go
ts, err := time.Parse("2006-01-02", rawDate)
if err != nil {
	return err
}
use(ts)
```

어떻게 빨리 잡나: test, review.

Rule: Go 시간 layout에서 마법의 날짜가 곧 format language입니다.

#### 48. `os/exec.CommandContext`가 `Wait` 책임까지 대신한다고 착각함

**Bad**

```go
cmd := exec.CommandContext(ctx, "worker", "--once")
if err := cmd.Start(); err != nil {
	return err
}

<-ctx.Done()
return ctx.Err()
```

왜 문제인가: `CommandContext`는 cancellation policy만 연결합니다. `Start`를 썼다면 리소스 해제와 I/O copier 종료를 위해 여전히 `Wait`가 필요합니다.

**Better**

```go
cmd := exec.CommandContext(ctx, "worker", "--once")
if err := cmd.Start(); err != nil {
	return err
}

if err := cmd.Wait(); err != nil {
	return err
}
return nil
```

어떻게 빨리 잡나: test, review.

Rule: `CommandContext`는 cancel 정책을 다루지 `Wait` ownership을 없애지 않습니다.

#### 49. `context.Value`를 옵션 파라미터나 config 전달용으로 남용함

**Bad**

```go
func ListUsers(ctx context.Context) error {
	limit, _ := ctx.Value("limit").(int)
	return query(limit)
}
```

왜 문제인가: API contract가 숨겨지고, weakly typed해지고, grep이나 validation도 어려워집니다.

**Better**

```go
type ListUsersOptions struct {
	Limit int
}

func ListUsers(ctx context.Context, opt ListUsersOptions) error {
	return query(opt.Limit)
}
```

어떻게 빨리 잡나: review.

Rule: `context.Value`는 request-scoped metadata에만 쓰고, 일반 파라미터는 함수 인자로 드러내십시오.

#### 50. `sync.Map`을 기본 map 선택지처럼 사용함

**Bad**

```go
var users sync.Map

func SetUser(id string, u User) {
	users.Store(id, u)
}
```

왜 문제인가: `sync.Map`은 특수화된 도구입니다. 기본값처럼 쓰면 타입 안정성과 invariant가 흐려집니다.

**Better**

```go
type UserStore struct {
	mu    sync.Mutex
	users map[string]User
}

func (s *UserStore) SetUser(id string, u User) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.users[id] = u
}
```

어떻게 빨리 잡나: review, contention 때문에 쓴다면 benchmark.

Rule: 기본값은 `map + mutex`이고, workload가 진짜 맞을 때만 `sync.Map`으로 가십시오.

## Practical takeaway

아주 아픈 Go 버그는 대개 희귀한 마법에서 나오지 않습니다.

대부분은 API 경계, ownership 경계, evaluation 경계에 대한 작은 오해에서 나옵니다. 이 부록은 바로 그 지점을 조여 주기 위해 존재합니다.
