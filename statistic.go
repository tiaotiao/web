package web

import (
	"fmt"
	"net/http"
	"sort"
	"time"
	"wps.cn/lib/go/sync2"
)

type Stat struct {
	Total    sync2.AtomicInt32 // Total requests count
	Handling sync2.AtomicInt32 // Handling requests count
	Handlers []*HandlerStat    // Stat of individual handler, which refers to WebHandler.stat
}

func newStat() *Stat {
	s := new(Stat)
	s.Handlers = make([]*HandlerStat, 0, 1024)
	return s
}

func (s *Stat) onServe(req *http.Request) {
	s.Total.Add(1)
	s.Handling.Add(1)
}

func (s *Stat) onDone(req *http.Request) {
	s.Handling.Add(-1)
}

func (s *Stat) TopCount() []*HandlerStat {
	return s.sortBy(func(a *HandlerStat, b *HandlerStat) bool {
		return a.Count.Get() < b.Count.Get()
	})
}

func (s *Stat) TopCountOK() []*HandlerStat {
	return s.sortBy(func(a *HandlerStat, b *HandlerStat) bool {
		return a.CountOK.Get() < b.CountOK.Get()
	})
}

func (s *Stat) TopCountErrs() []*HandlerStat {
	return s.sortBy(func(a *HandlerStat, b *HandlerStat) bool {
		aCount := a.Count4XX.Get() + a.Count5XX.Get()
		bCount := b.Count4XX.Get() + b.Count5XX.Get()
		return aCount < bCount
	})
}

func (s *Stat) TopCount4XXErrs() []*HandlerStat {
	return s.sortBy(func(a *HandlerStat, b *HandlerStat) bool {
		return a.Count4XX.Get() < b.Count4XX.Get()
	})
}

func (s *Stat) TopCount5XXErrs() []*HandlerStat {
	return s.sortBy(func(a *HandlerStat, b *HandlerStat) bool {
		return a.Count5XX.Get() < b.Count5XX.Get()
	})
}

func (s *Stat) TopAvgTime() []*HandlerStat {
	return s.sortBy(func(a *HandlerStat, b *HandlerStat) bool {
		return a.AverageTime.Get() < b.AverageTime.Get()
	})
}

func (s *Stat) TopMaxTime() []*HandlerStat {
	return s.sortBy(func(a *HandlerStat, b *HandlerStat) bool {
		return a.MaxTime.Get() < b.MaxTime.Get()
	})
}

func (s *Stat) TopUsedTime() []*HandlerStat {
	return s.sortBy(func(a *HandlerStat, b *HandlerStat) bool {
		return a.UsedTime.Get() < b.UsedTime.Get()
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
	Count       sync2.AtomicInt32
	CountOK     sync2.AtomicInt32
	Count4XX    sync2.AtomicInt32
	Count5XX    sync2.AtomicInt32
	AverageTime sync2.AtomicDuration
	MaxTime     sync2.AtomicDuration
	UsedTime    sync2.AtomicDuration
}

func (s *HandlerStat) onServe(code int, usedTime time.Duration) {

	s.Count.Add(1)
	if code < http.StatusBadRequest {
		s.CountOK.Add(1)
	} else if code < http.StatusInternalServerError {
		s.Count4XX.Add(1)
	} else {
		s.Count5XX.Add(1)
	}

	s.UsedTime.Add(usedTime)
	if s.MaxTime.Get() < usedTime {
		s.MaxTime.Set(usedTime)
	}
	s.AverageTime.Set(s.UsedTime.Get() / time.Duration(s.Count.Get()))
}

func (s *HandlerStat) Format() string {
	return fmt.Sprintf("%-40v\t count=%v,\t countok=%v,\t count4xx=%v,\t count5xx=%v,\t averagetime=%v,\t maxtime=%v,\t usedtime=%v",
		"["+s.Path+"],", s.Count.Get(), s.CountOK.Get(), s.Count4XX.Get(), s.Count5XX.Get(), s.AverageTime.Get(), s.MaxTime.Get(), s.UsedTime.Get())
}
