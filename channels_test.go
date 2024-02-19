package channels

import (
	"bufio"
	"context"
	"os"
	"slices"
	"strings"
	"testing"
	"time"
)

var TheWords = func() []string {
	f, err := os.Open("/usr/share/dict/words")
	if err != nil {
		panic("error opening words file: " + err.Error())
	}
	defer f.Close()
	results := make([]string, 0)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if line := strings.TrimSpace(scanner.Text()); len(line) > 0 {
			results = append(results, line)
		}
	}
	slices.Sort(results)
	return results
}()

func TestContextDoneAndContextNotDone(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if ContextDone(ctx) {
		t.Fatalf("error: brand new context is done...\n")
	}
	cancel()
	time.Sleep(time.Millisecond * 5)
	if ContextNotDone(ctx) {
		t.Fatalf("error: canceled context is not done...\n")
	}
}

func TestAddToChannelWithContext(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*25)
	defer cancel()
	c := make(chan int, 1)
	if err := AddToChannelWithContext(ctx, c, 1); err != nil {
		t.Fatalf("error adding to channel: %v\n", err)
	} else if err := AddToChannelWithContext(ctx, c, 2); err == nil {
		t.Fatalf("error: should have timed out adding to full channel...\n")
	}
}

func TestSliceToChannelAndChannelToSlice(t *testing.T) {
	sample := TheWords
	c := SliceToChannel(len(sample), sample)
	contents := ChannelToSlice(c)
	if want, got := len(sample), len(contents); want != got {
		t.Fatalf("error: wanted %d in results; got %d\n", len(sample), len(contents))
	} else if want, got := sample, contents; !slices.Equal(want, got) {
		t.Fatalf("error: wanted %v; got %v\n", want, got)
	}
}

func TestCopyChannelWithContext(t *testing.T) {
	sample := TheWords
	a, b := CopyChannel(SliceToChannel(len(sample), sample))
	gotA, gotB := ChannelToSlice(a), ChannelToSlice(b)
	if want, got := len(sample), len(gotA); want != got {
		t.Fatalf("error: wanted %d; got %d\n", want, got)
	} else if want, got := sample, gotA; !slices.Equal(want, got) {
		t.Fatalf("error: wanted %v; got %v\n", want, got)
	} else if want, got := len(sample), len(gotB); want != got {
		t.Fatalf("error: wanted %d; got %d\n", want, got)
	} else if want, got := sample, gotB; !slices.Equal(want, got) {
		t.Fatalf("error: wanted %v; got %v\n", want, got)
	}
}

func TestMultiplexWithContext(t *testing.T) {
	f := func(i int) int { return i * i }
	workers := 100
	work, want := func() (<-chan int, []int) {
		c := make(chan int, 50)
		defer close(c)
		s := make([]int, 0, 50)
		for i := 0; i < 50; i++ {
			c <- i
			s = append(s, f(i))
		}
		return c, s
	}()
	results := ChannelToSlice(MultiplexWithContext(context.Background(), work, workers, f))
	slices.Sort(results)
	if want, got := want, results; !slices.Equal(want, got) {
		t.Fatalf("error: wanted %v; got %v\n", want, got)
	}

}
