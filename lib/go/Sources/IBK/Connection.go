package IBK

import (
	"context"
	"errors"
	"time"

	"github.com/ThomasMarcelis/ibkr-go/v2"
	"github.com/grdvsng/stock-analysis/Result"
)

const HISTORY_BAR_SIZE = ibkr.Bar1Min

type Connection struct {
	ctx     *context.Context
	close   context.CancelFunc
	timeout time.Duration
	client  *ibkr.Client
}

func (conn *Connection) request() (context.Context, context.CancelFunc) {
	return context.WithTimeout(*conn.ctx, conn.timeout)
}

func (conn *Connection) Bars(ctx context.Context, contract ibkr.Contract, end time.Time, whatToShow ibkr.WhatToShow) ([]ibkr.Bar, error) {
	ctx, cancel := context.WithTimeout(ctx, conn.timeout)
	defer cancel()

	return conn.client.History().Bars(ctx, ibkr.HistoricalBarsRequest{
		Contract:   contract,
		EndTime:    end,
		Duration:   ibkr.Days(1),
		BarSize:    HISTORY_BAR_SIZE,
		WhatToShow: whatToShow,
	})
}

func (conn *Connection) Search(pattern string) ([]ibkr.MatchingSymbol, error) {
	ctx, cancel := conn.request()
	defer cancel()

	return conn.client.Contracts().Search(ctx, pattern)
}

func (conn *Connection) GetContract(id string) (ibkr.ContractDetails, bool, error) {
	ctx, cancel := conn.request()
	defer cancel()

	contract, e := conn.client.Contracts().Qualify(ctx, ibkr.Stock(id))

	if e != nil {
		if errors.Is(e, ibkr.ErrNoMatch) || errors.Is(e, ibkr.ErrAmbiguousContract) {
			return contract, false, nil
		}

		return contract, false, errors.New(IBKApiError.Render(false, e))
	}

	return contract, true, nil
}

func (conn *Connection) LastTicks(contract ibkr.Contract, whatToShow ibkr.WhatToShow) (ibkr.HistoricalTicksResult, error) {
	ctx, cancel := conn.request()
	defer cancel()

	return conn.client.History().Ticks(ctx, ibkr.HistoricalTicksRequest{
		Contract:      contract,
		EndTime:       time.Now().UTC(),
		NumberOfTicks: 1,
		WhatToShow:    whatToShow,
	})
}

func (conn *Connection) Close() {
	if conn.ctx != nil {
		conn.close()
	}

	if conn.client != nil {
		conn.client.Close()
	}
}

func CreateConnection(host string, port, timeout, clientId int) *Connection {
	return Result.Ok(Dial(host, port, timeout, clientId))
}

func Dial(host string, port, timeout, clientId int) (*Connection, error) {
	ctx, close := context.WithCancel(context.Background())
	conn := &Connection{
		ctx:     &ctx,
		close:   close,
		timeout: time.Duration(timeout) * time.Second,
	}

	dialCtx, cancel := conn.request()
	defer cancel()

	client, e := ibkr.DialContext(dialCtx, ibkr.WithHost(host), ibkr.WithPort(port), ibkr.WithClientID(ibkr.ClientID(clientId)))

	if e != nil {
		close()

		return nil, e
	}

	conn.client = client

	return conn, nil
}
