---
title: Collections와 Memory 함정
description: subtle bug와 불필요한 allocation을 만드는 collection과 memory-shape 실수 10가지입니다.
---

# Collections와 Memory 함정

## 21. nil map에 write함

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

## 22. nil slice와 empty slice를 API나 JSON에서 같다고 가정함

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

## 23. nil map과 empty map을 API나 JSON에서 같다고 가정함

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

## 24. shared backing array 때문에 `append`가 서로 엉킴

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

## 25. subslice가 거대한 backing array를 붙잡아 둠

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

## 26. `range` variable의 주소를 잡음

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

## 27. slice를 `range`하면서 동시에 수정함

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

## 28. map iteration order가 안정적이라고 믿음

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

## 29. 동기화 없이 map을 concurrent access함

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

## 30. `string`과 `[]byte` 변환이 공짜이거나 storage를 공유한다고 착각함

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
