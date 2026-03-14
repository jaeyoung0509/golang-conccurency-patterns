---
title: Stdlib and API Boundary Pitfalls
description: Ten standard-library and API-boundary mistakes that compile cleanly and still cause production trouble.
---

# Stdlib and API Boundary Pitfalls

## 41. Creating a new `http.Client` or `Transport` per request

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

Why it bites: connection pooling and transport reuse disappear. You pay for handshakes and sockets over and over again.

**Better**

```go
var client = &http.Client{
	Timeout: 5 * time.Second,
	Transport: &http.Transport{
		MaxIdleConnsPerHost: 16,
	},
}
```

Catch it early: load tests, metrics, review.

Rule: `http.Client` and especially `Transport` are long-lived policy objects, not throwaway helpers.

## 42. Not draining and closing `http.Response.Body`

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

Why it bites: the body stays open, and the transport may not be able to reuse the connection.

**Better**

```go
resp, err := client.Do(req)
if err != nil {
	return err
}
defer resp.Body.Close()
_, _ = io.Copy(io.Discard, resp.Body)
```

Catch it early: tests, transport metrics, review.

Rule: if you got a response body, close it, and drain it when connection reuse matters.

## 43. Assuming `io.Reader` fills the whole buffer in one read

**Bad**

```go
buf := make([]byte, 32)
n, err := r.Read(buf)
if err != nil {
	return err
}
return decode(buf[:n])
```

Why it bites: `Read` can return fewer bytes than you asked for and still be perfectly correct.

**Better**

```go
buf := make([]byte, 32)
if _, err := io.ReadFull(r, buf); err != nil {
	return err
}
return decode(buf)
```

Catch it early: tests with short readers, review.

Rule: `io.Reader` promises a stream, not full-buffer delivery.

## 44. Forgetting `Rows.Close` and `Rows.Err` in `database/sql`

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

Why it bites: you leak resources if you never close rows, and you can miss iteration errors if you never check `Rows.Err`.

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

Catch it early: integration tests, review.

Rule: with `Rows`, always think in the trio `Query`, `Close`, and `Err`.

## 45. Using `Query` when `QueryRow` or `Exec` matches the contract better

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

Why it bites: `Query` implies result-set handling. For statements that do not return rows, it makes the contract noisier and easier to misuse.

**Better**

```go
_, err := db.ExecContext(ctx, `
	INSERT INTO audit_log(actor, action) VALUES (?, ?)
`, actor, action)
return err
```

Catch it early: review.

Rule: choose the narrowest `database/sql` API that matches the result shape.

## 46. `json.Decoder` silently accepting unknown fields when strictness was intended

**Bad**

```go
var req CreateUserRequest

dec := json.NewDecoder(r.Body)
if err := dec.Decode(&req); err != nil {
	return err
}
```

Why it bites: typoed or unexpected fields are ignored by default. Clients think they sent data the server never used.

**Better**

```go
var req CreateUserRequest

dec := json.NewDecoder(r.Body)
dec.DisallowUnknownFields()
if err := dec.Decode(&req); err != nil {
	return err
}
```

Catch it early: request-validation tests, review.

Rule: if input contracts are strict, make the decoder strict too.

## 47. `time.Parse` layout mistakes from not using the reference time correctly

**Bad**

```go
ts, err := time.Parse("YYYY-MM-DD", rawDate)
if err != nil {
	return err
}
use(ts)
```

Why it bites: Go layouts are not symbolic format strings. They must be written using the reference time.

**Better**

```go
ts, err := time.Parse("2006-01-02", rawDate)
if err != nil {
	return err
}
use(ts)
```

Catch it early: tests, review.

Rule: in Go time layouts, the magic date is the format language.

## 48. Assuming `os/exec.CommandContext` replaces explicit `Wait`

**Bad**

```go
cmd := exec.CommandContext(ctx, "worker", "--once")
if err := cmd.Start(); err != nil {
	return err
}

<-ctx.Done()
return ctx.Err()
```

Why it bites: `CommandContext` ties cancellation to the process, but if you call `Start`, you still need `Wait` to release resources and wait for I/O copying to finish.

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

Catch it early: tests, review.

Rule: `CommandContext` handles cancellation policy, not `Wait` ownership.

## 49. Using `context.Value` for optional parameters or config

**Bad**

```go
func ListUsers(ctx context.Context) error {
	limit, _ := ctx.Value("limit").(int)
	return query(limit)
}
```

Why it bites: the API contract becomes hidden, weakly typed, and hard to grep or validate.

**Better**

```go
type ListUsersOptions struct {
	Limit int
}

func ListUsers(ctx context.Context, opt ListUsersOptions) error {
	return query(opt.Limit)
}
```

Catch it early: review.

Rule: use `context.Value` for request-scoped metadata, not ordinary function parameters.

## 50. Using `sync.Map` as the default map choice instead of a specialized tool

**Bad**

```go
var users sync.Map

func SetUser(id string, u User) {
	users.Store(id, u)
}
```

Why it bites: `sync.Map` is specialized. Most code loses type safety and invariant clarity when it reaches for it by default.

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

Catch it early: review, benchmarks when contention is the reason.

Rule: start with `map + mutex`; reach for `sync.Map` only when its specialization actually matches the workload.
