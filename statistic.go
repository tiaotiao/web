package web

import (
	"fmt"
	"net/http"
	"sort"
	"sync/atomic"
	"time"
)

type Stat struct {
	Total    int64          // Total requests count
	Handling int64          // Handling requests count
	Handlers []*HandlerStat // Stat of individual handler, which refers to WebHandler.stat
}

func newStat() *Stat {
	s := new(Stat)
	s.Handlers = make([]*HandlerStat, 0, 1024)
	return s
}

func (s *Stat) onServe(req *http.Request) {
	atomic.AddInt64(&s.Total, 1)
	atomic.AddInt64(&s.Handling, 1)
}

func (s *Stat) onDone(req *http.Request) {
	atomic.AddInt64(&s.Handling, -1)
}

func (s *Stat) TopCount() []*HandlerStat {
	return s.sortBy(func(a *HandlerStat, b *HandlerStat) bool {
		return a.Count < b.Count
	})
}

func (s *Stat) TopCountOK() []*HandlerStat {
	return s.sortBy(func(a *HandlerStat, b *HandlerStat) bool {
		return a.CountOK < b.CountOK
	})
}

func (s *Stat) TopCountErrs() []*HandlerStat {
	return s.sortBy(func(a *HandlerStat, b *HandlerStat) bool {
		aCount := a.Count4XX + a.Count5XX
		bCount := b.Count4XX + b.Count5XX
		return aCount < bCount
	})
}

func (s *Stat) TopCount4XXErrs() []*HandlerStat {
	return s.sortBy(func(a *HandlerStat, b *HandlerStat) bool {
		return a.Count4XX < b.Count4XX
	})
}

func (s *Stat) TopCount5XXErrs() []*HandlerStat {
	return s.sortBy(func(a *HandlerStat, b *HandlerStat) bool {
		return a.Count5XX < b.Count5XX
	})
}

func (s *Stat) TopAvgTime() []*HandlerStat {
	return s.sortBy(func(a *HandlerStat, b *HandlerStat) bool {
		return a.AverageTime < b.AverageTime
	})
}

func (s *Stat) TopMaxTime() []*HandlerStat {
	return s.sortBy(func(a *HandlerStat, b *HandlerStat) bool {
		return a.MaxTime < b.MaxTime
	})
}

func (s *Stat) TopUsedTime() []*HandlerStat {
	return s.sortBy(func(a *HandlerStat, b *HandlerStat) bool {
		return a.UsedTime < b.UsedTime
	})
}

func (s *Stat) sortBy(less func(*HandlerStat, *HandlerStat) bool) []*HandlerStat {
	ss := new(sortable)
	ss.less = less
	ss.stats = make([]*HandlerStat, len(s.Handlers))
	copy(ss.stats, s.Handlers)

	sort.Sort(ss)

	return ss.stats
}

type sortable struct {
	stats []*HandlerStat
	less  func(*HandlerStat, *HandlerStat) bool
}

func (s *sortable) Len() int {
	return len(s.stats)
}

func (s *sortable) Swap(i, j int) {
	s.stats[i], s.stats[j] = s.stats[j], s.stats[i]
}

func (s *sortable) Less(i, j int) bool {
	return s.less(s.stats[j], s.stats[i])
}

type HandlerStat struct {
	Path        string
	Count       int64
	CountOK     int64
	Count4XX    int64
	Count5XX    int64
	AverageTime int64
	MaxTime     int64
	UsedTime    int64
}

func (s *HandlerStat) onServe(code int, usedTime time.Duration) {

	atomic.AddInt64(&s.Count, 1)
	if code < http.StatusBadRequest {
		atomic.AddInt64(&s.CountOK, 1)
	} else if code < http.StatusInternalServerError {
		atomic.AddInt64(&s.Count4XX, 1)
	} else {
		atomic.AddInt64(&s.Count5XX, 1)
	}

	atomic.AddInt64(&s.UsedTime, int64(usedTime))
	if s.MaxTime < int64(usedTime) {
		atomic.StoreInt64(&s.MaxTime, int64(usedTime))
	}
	atomic.StoreInt64(&s.AverageTime, s.UsedTime/s.Count)
}

func (s *HandlerStat) Format() string {
	return fmt.Sprintf("%-40v\t count=%v,\t countok=%v,\t count4xx=%v,\t count5xx=%v,\t averagetime=%v,\t maxtime=%v,\t usedtime=%v",
		"["+s.Path+"],", s.Count, s.CountOK, s.Count4XX, s.Count5XX, s.AverageTime, s.MaxTime, s.UsedTime)
}
