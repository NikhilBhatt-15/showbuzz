package trans

import (
	"slices"
	"testing"
)

func TestMoviePricing(t *testing.T) {
	tx := &Txn{Kind: "MOVIE", Qty: 2, Snapshot: Snapshot{Lines: []Line{{AreaCode: "PR", Price: 340, Seats: []string{"J10", "J11"}}}},
		Fnb: []FnbLine{{Code: "COMB", Price: 620, Qty: 1}}, Donation: true}
	p := tx.Pricing()
	if p.Ticket != 680 || p.ConvBase != 50 || p.GST != 9 || p.ConvFee != 59 || p.SubTotal != 739 {
		t.Fatalf("unexpected ticket pricing: %+v", p)
	}
	if p.Fnb != 620 || p.Donation != 2 || p.Total != 1361 {
		t.Fatalf("unexpected totals: %+v", p)
	}
}

func TestEventPricing(t *testing.T) {
	tx := &Txn{Kind: "EVENT", Qty: 3, Snapshot: Snapshot{Lines: []Line{{Price: 799}}}}
	p := tx.Pricing()
	if p.Ticket != 2397 || p.ConvBase != 143.82 || p.GST != 25.89 || p.Total != 2566.71 {
		t.Fatalf("unexpected event pricing: %+v", p)
	}
}

func TestSeatOrder(t *testing.T) {
	seats := []string{"J10", "J9", "A1", "J11"}
	slices.SortFunc(seats, seatOrder)
	if want := []string{"A1", "J9", "J10", "J11"}; !slices.Equal(seats, want) {
		t.Fatalf("got %v, want %v", seats, want)
	}
}

func TestGatewaySimulation(t *testing.T) {
	ok := payRequest{Method: "CARD"}
	ok.Card.Number, ok.Card.CVV, ok.Card.Expiry = "4111 1111 1111 1111", "123", "12/30"
	if _, err := simulateGateway(ok); err != nil {
		t.Fatalf("valid card declined: %v", err)
	}
	bad := ok
	bad.Card.Number = "4000 0000 0000 0002"
	if _, err := simulateGateway(bad); err == nil || err.Code != "PAYMENT_DECLINED" {
		t.Fatalf("decline card accepted: %v", err)
	}
	if _, err := simulateGateway(payRequest{Method: "UPI", UPI: "fail@upi"}); err == nil {
		t.Fatal("failing UPI accepted")
	}
}
