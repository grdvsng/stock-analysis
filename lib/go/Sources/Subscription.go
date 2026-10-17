package Sources

import (
	"sync"

	"github.com/grdvsng/stock-analysis/Data/Market"
	"github.com/grdvsng/stock-analysis/Logger"
)

type OnPostHanlder func(metric Market.TradeMetric)

type OnErrorHandler func(err error)

type Subscription struct {
	Market.Contract

	OnPost  OnPostHanlder
	OnError OnErrorHandler
}

func (s *Subscription) Post(metric Market.TradeMetric) {
	if s.OnPost != nil {
		s.OnPost(metric)
	}
}

func (s *Subscription) Error(err error) {
	if s.OnError != nil {
		s.OnError(err)
	} else {
		Logger.Error(SubscriptionError, s.Stock.Id, err)
	}
}

type Subscrubtions struct {
	m       *sync.RWMutex
	storage map[string]*Subscription
}

func CreateSubscriptions() *Subscrubtions {
	return &Subscrubtions{m: &sync.RWMutex{}, storage: map[string]*Subscription{}}
}

func (s *Subscrubtions) UnSubscribe(contract *Market.Contract) {
	s.m.Lock()
	defer s.m.Unlock()

	delete(s.storage, contract.Stock.Id)
}

func (s *Subscrubtions) Subscribed(contract *Market.Contract) bool {
	_, exists := s.storage[contract.Stock.Id]

	return exists
}

func (s *Subscrubtions) Subscribe(contract *Market.Contract, onPost OnPostHanlder, onError OnErrorHandler) *Subscription {
	s.m.Lock()
	defer s.m.Unlock()

	if s.Subscribed(contract) {
		Logger.Throw(StockAllreadySubscriber, contract.Stock.Id)
	} else {
		s.storage[contract.Stock.Id] = &Subscription{Contract: *contract, OnPost: onPost, OnError: onError}
	}

	return s.storage[contract.Stock.Id]
}
