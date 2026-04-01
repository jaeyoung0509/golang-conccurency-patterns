package ordoneteebridge

import (
	"context"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestOrDoneStopsOnCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	in := make(chan int)
	out := OrDone(ctx, in)

	cancel()

	select {
	case _, ok := <-out:
		if ok {
			t.Fatal("expected OrDone output to close after cancellation")
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("OrDone did not stop after cancellation")
	}
}

func TestTeeCopiesValuesToBothOutputs(t *testing.T) {
	ctx := context.Background()
	in := make(chan int, 3)
	in <- 10
	in <- 20
	in <- 30
	close(in)

	left, right := Tee(ctx, in)

	var wg sync.WaitGroup
	var leftValues []int
	var rightValues []int

	wg.Add(2)
	go func() {
		defer wg.Done()
		for value := range left {
			leftValues = append(leftValues, value)
		}
	}()
	go func() {
		defer wg.Done()
		for value := range right {
			rightValues = append(rightValues, value)
		}
	}()
	wg.Wait()

	want := []int{10, 20, 30}
	if !reflect.DeepEqual(leftValues, want) {
		t.Fatalf("left values = %v, want %v", leftValues, want)
	}
	if !reflect.DeepEqual(rightValues, want) {
		t.Fatalf("right values = %v, want %v", rightValues, want)
	}
}

func TestBridgeFlattensStreamOfStreams(t *testing.T) {
	out := MergeTenantFeeds(
		context.Background(),
		[]FeedItem{
			{TenantID: "tenant-a", Offset: 1, Kind: "created"},
			{TenantID: "tenant-a", Offset: 2, Kind: "paid"},
		},
		[]FeedItem{
			{TenantID: "tenant-b", Offset: 1, Kind: "created"},
		},
	)

	var got []FeedItem
	for item := range out {
		got = append(got, item)
	}

	want := []FeedItem{
		{TenantID: "tenant-a", Offset: 1, Kind: "created"},
		{TenantID: "tenant-a", Offset: 2, Kind: "paid"},
		{TenantID: "tenant-b", Offset: 1, Kind: "created"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("merged feed = %#v, want %#v", got, want)
	}
}
