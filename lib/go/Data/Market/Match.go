package Market

type Match struct {
	Symbol          string `json:"symbol"`
	Name            string `json:"name"`
	PrimaryExchange string `json:"primary_exchange"`
	Currency        string `json:"currency"`
	Country         string `json:"country"`
}
