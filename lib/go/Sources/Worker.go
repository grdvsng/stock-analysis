package Sources

import (
	"sync"

	"github.com/grdvsng/stock-analysis/Logger"
)

type worker struct {
	wg      *sync.WaitGroup
	sources []Source
}

func CreateWorker(sources []Source) *worker {
	wg := &sync.WaitGroup{}

	for i, source := range sources {
		Logger.Info("Init process %d", i)
		wg.Add(1)
		go func(wg *sync.WaitGroup, source Source) {
			defer wg.Done()
			source.Run()
		}(wg, source)
	}

	return &worker{
		sources: sources,
		wg:      wg,
	}
}

func (w *worker) Sources() []Source {
	return w.sources
}

func (w *worker) Close() {
	for i, source := range w.sources {
		Logger.Info("Stop process %d", i)

		source.Close()
	}

	w.wg.Wait()
}
