package Sources

import (
	"github.com/grdvsng/stock-analysis/Exceptions"
)

var StockAllreadySubscriber = Exceptions.CreateError("Stock %s allready subscribed")

var SubscriptionError = Exceptions.CreateError("Subscription %s error: %s")
