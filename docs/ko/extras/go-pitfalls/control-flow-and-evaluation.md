---
title: Control Flow와 Evaluation 함정
description: 숙련된 Go 엔지니어도 자주 걸리는 control-flow와 evaluation 관련 실수 10가지입니다.
---

# Control Flow와 Evaluation 함정

## 1. `select`에 `default`를 넣어 의도치 않은 busy-spin이 생김

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

## 2. nil channel branch가 `select` case를 조용히 비활성화함

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

## 3. 닫힌 channel에서 받은 zero value를 실제 값으로 오해함

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

## 4. 닫힌 channel로 send함

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

## 5. `select`가 공정할 거라고 가정함

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

## 6. `select`나 `switch` 안의 `break`가 바깥 loop를 끝내지 못함

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

## 7. goroutine에서 loop variable을 capture함

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

## 8. `defer`에서 loop variable을 capture함

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

## 9. `:=`로 `err`를 shadowing함

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

## 10. loop 안의 `defer`가 리소스를 너무 오래 붙잡음

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
