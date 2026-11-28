package store

import "testing"

func TestCreateSale_FacebookReceived(t *testing.T) {
	st := newTestStore(t)

	accountID, err := st.InsertCashAccount("PIX principal", "pix", 0)
	if err != nil {
		t.Fatal(err)
	}
	lotID, err := st.CreateLotPurchase(CreateLotInput{
		Name:        "Monitores — Facebook",
		PurchasedAt: "2026-11-20",
		ItemTitle:   "Monitor Dell P2016t",
		ItemQty:     1,
		Costs: []CostInput{
			{Label: "Arremate", AmountCents: 3121, AlreadyPaid: true},
		},
		CashAccountID: accountID,
		PaidAt:        "2026-11-20T12:00:00Z",
	})
	if err != nil {
		t.Fatal(err)
	}
	items, err := st.ListItemsByLot(lotID)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("items=%d want 1", len(items))
	}

	saleID, err := st.CreateSale(CreateSaleInput{
		ItemID:        items[0].ID,
		SoldAt:        "2026-11-28T12:00:00Z",
		Channel:       "facebook",
		GrossCents:    15000,
		PaymentStatus: "received",
		CashAccountID: accountID,
	})
	if err != nil {
		t.Fatalf("facebook sale rejected: %v", err)
	}

	sale, err := st.FindSaleByID(saleID)
	if err != nil {
		t.Fatal(err)
	}
	if sale.Channel != "facebook" {
		t.Errorf("channel = %q, want facebook", sale.Channel)
	}
	if sale.NetCents != 15000 {
		t.Errorf("net = %d, want 15000", sale.NetCents)
	}

	sold, err := st.FindItemByID(items[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	if sold.Status != "sold" {
		t.Errorf("item status = %q, want sold", sold.Status)
	}

	balance, err := st.CashBalance(accountID)
	if err != nil {
		t.Fatal(err)
	}
	if balance != -3121+15000 {
		t.Errorf("balance = %d, want %d", balance, -3121+15000)
	}
}

func TestCreateSale_RejectsUnknownChannel(t *testing.T) {
	st := newTestStore(t)
	accountID, err := st.InsertCashAccount("PIX principal", "pix", 0)
	if err != nil {
		t.Fatal(err)
	}
	lotID, err := st.CreateLotPurchase(CreateLotInput{
		Name:          "Lote canal inválido",
		PurchasedAt:   "2026-11-20",
		ItemTitle:     "Monitor",
		ItemQty:       1,
		Costs:         []CostInput{{Label: "Arremate", AmountCents: 1000, AlreadyPaid: true}},
		CashAccountID: accountID,
		PaidAt:        "2026-11-20T12:00:00Z",
	})
	if err != nil {
		t.Fatal(err)
	}
	items, err := st.ListItemsByLot(lotID)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := st.CreateSale(CreateSaleInput{
		ItemID:        items[0].ID,
		SoldAt:        "2026-11-28T12:00:00Z",
		Channel:       "instagram",
		GrossCents:    15000,
		PaymentStatus: "received",
		CashAccountID: accountID,
	}); err == nil {
		t.Fatal("expected CHECK violation for unknown channel, got nil")
	}
	if it, err := st.FindItemByID(items[0].ID); err != nil || it.Status != "in_stock" {
		t.Fatalf("item should stay in_stock after rejected sale: %+v %v", it, err)
	}
}
