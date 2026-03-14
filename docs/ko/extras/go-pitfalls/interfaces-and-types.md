---
title: Interfaces와 Types 함정
description: 컴파일은 되지만 인터페이스와 타입 시스템 경계에서 놀라운 버그를 만드는 실수 10가지입니다.
---

# Interfaces와 Types 함정

## 11. interface 안의 typed nil이 `nil`과 같지 않음

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

## 12. concrete value가 comparable하지 않은 interface끼리 비교함

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

## 13. pointer receiver와 value receiver를 헷갈려 interface 구현을 잘못 이해함

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

## 14. nil receiver method를 허용해 놓고 내부에서는 바로 panic남

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

## 15. implicit interface satisfaction을 확인하지 않고 가정만 함

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

## 16. type assertion을 검사 없이 씀

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

## 17. wrapped error를 `==`로 비교함

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

## 18. `any`가 도메인 의미까지 보존해 줄 거라고 생각함

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

## 19. `json.Unmarshal`을 `interface{}`에 넣고 숫자가 `float64`가 되는 걸 놓침

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

## 20. interface field가 zero value와 absent state를 흐림

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
