package engine

// PriceCompatible 判断卖方报价是否不高于买方出价；任一金额无效时返回 false。
func PriceCompatible(sellerAsk, buyerBid Amount) bool {
	return sellerAsk.valid() && buyerBid.valid() && sellerAsk.cents <= buyerBid.cents
}
