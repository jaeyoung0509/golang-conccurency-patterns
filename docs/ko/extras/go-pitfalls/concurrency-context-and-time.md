---
title: Concurrency, Context, Time 함정
description: leak, wedge, 시끄러운 shutdown을 만드는 동시성/lifecycle 실수 10가지입니다.
---

# Concurrency, Context, Time 함정

## 31. receiver 쪽에서 channel을 닫음

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

## 32. 여러 sender가 같은 channel을 닫으려 경쟁함

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

## 33. `len(ch)`를 coordination signal로 사용함

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

## 34. caller timeout 이후 reply send가 막혀 goroutine leak이 남음

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

## 35. owner나 stop condition 없이 goroutine을 시작함

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

## 36. request-scoped work에서 `context.Background()`를 사용함

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

## 37. `cancel` 호출을 잊음

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

## 38. loop 안의 `time.After`가 timer churn을 만듦

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

## 39. `Ticker`를 만들고 `Stop`하지 않음

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

## 40. goroutine 시작 후에 `WaitGroup.Add`를 호출함

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
