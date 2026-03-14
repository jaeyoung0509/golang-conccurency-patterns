---
title: Stdlib과 API Boundary 함정
description: 컴파일은 되지만 표준 라이브러리와 API 경계에서 실제 운영 문제를 만드는 실수 10가지입니다.
---

# Stdlib과 API Boundary 함정

## 41. 요청마다 새로운 `http.Client`나 `Transport`를 만듦

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

## 42. `http.Response.Body`를 drain하지도 close하지도 않음

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

## 43. `io.Reader`가 한 번에 버퍼를 다 채워 줄 거라고 생각함

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

## 44. `database/sql`에서 `Rows.Close`와 `Rows.Err`를 빼먹음

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

## 45. `QueryRow`나 `Exec`가 맞는 자리에 `Query`를 씀

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

## 46. strict input을 원했는데 `json.Decoder`가 unknown field를 조용히 무시함

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

## 47. reference time을 쓰지 않아 `time.Parse` layout을 틀림

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

## 48. `os/exec.CommandContext`가 `Wait` 책임까지 대신한다고 착각함

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

## 49. `context.Value`를 옵션 파라미터나 config 전달용으로 남용함

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

## 50. `sync.Map`을 기본 map 선택지처럼 사용함

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
