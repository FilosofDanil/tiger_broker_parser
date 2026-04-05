package model

type NonZeroBalance struct {
	Asset  string
	Free   float64
	Locked float64
	Total  float64
}

type Position struct {
	Symbol     string
	Side       string
	Amt        float64
	EntryPrice float64
	MarkPrice  float64
	LiqPrice   float64
	UnPnL      float64
	Leverage   string
	MarginType string
}
