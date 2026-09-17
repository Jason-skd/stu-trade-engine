package engine

import "testing"

func TestPriceCompatible(t *testing.T) {
	t.Parallel()

	// === 定义测试用例 ===
	tests := []struct {
		name      string
		sellerAsk Amount
		buyerBid  Amount
		want      bool
	}{
		{name: "seller ask below buyer bid", sellerAsk: mustAmount(t, 1000), buyerBid: mustAmount(t, 1001), want: true},
		{name: "equal boundary", sellerAsk: mustAmount(t, 1000), buyerBid: mustAmount(t, 1000), want: true},
		{name: "seller ask above buyer bid", sellerAsk: mustAmount(t, 1001), buyerBid: mustAmount(t, 1000), want: false},
		{name: "invalid seller ask", buyerBid: mustAmount(t, 1000), want: false},
		{name: "invalid buyer bid", sellerAsk: mustAmount(t, 1000), want: false},
	}

	// === 执行测试用例 ===
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := PriceCompatible(tt.sellerAsk, tt.buyerBid); got != tt.want {
				t.Fatalf("PriceCompatible() = %t, want %t", got, tt.want)
			}
		})
	}
}
