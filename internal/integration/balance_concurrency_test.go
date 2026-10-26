package integration

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"dawn-errand/internal/infrastructure/ent/user"
	entuserprofile "dawn-errand/internal/infrastructure/ent/userprofile"
)

// R1-1 回归：余额支付必须原子扣减，并发支付不得双花。
// 场景：余额只够一笔，两笔订单并发支付 → 恰好一笔成功、余额非负且账实相符。

func profileBalance(t *testing.T, userID string) float64 {
	t.Helper()
	p, err := client.UserProfile.Query().
		Where(entuserprofile.UserID(userID)).Only(context.Background())
	if err != nil {
		t.Fatalf("profile %s: %v", userID, err)
	}
	return p.Balance
}

func setProfileBalance(t *testing.T, userID string, v float64) {
	t.Helper()
	_, err := client.UserProfile.Update().
		Where(entuserprofile.UserID(userID)).
		SetBalance(v).Save(context.Background())
	if err != nil {
		t.Fatalf("set balance: %v", err)
	}
}

func TestBalancePaymentConcurrency(t *testing.T) {
	const email = "client.test@dawnerrand.local"
	tok, _ := loginAs(t, email, "Client123456")
	u, err := client.User.Query().Where(user.Email(email)).Only(context.Background())
	if err != nil {
		t.Fatalf("查询测试用户: %v", err)
	}
	userID := u.ID
	if userID == "" {
		t.Fatal("未取到测试用户 ID")
	}

	orig := profileBalance(t, userID)
	defer setProfileBalance(t, userID, orig)

	// 建两笔订单，读出应付金额，构造只够支付一笔的余额。
	suffix := time.Now().UnixNano()
	idA := createOrderHTTP(t, tok, fmt.Sprintf("r11-a-%d", suffix), []map[string]any{{"dish_id": "D001", "count": 1}})
	idB := createOrderHTTP(t, tok, fmt.Sprintf("r11-b-%d", suffix), []map[string]any{{"dish_id": "D001", "count": 1}})
	defer func() {
		for _, id := range []string{idA, idB} {
			_, _ = do(t, http.MethodPost, "/api/app/client/orders/"+id+"/cancel", tok, nil)
			cleanupOrder(t, id)
		}
		// 扣除测试期间可能产生的退款入账，还原到原始余额。
		delta := profileBalance(t, userID) - orig
		if delta > 0.001 {
			setProfileBalance(t, userID, orig)
		}
	}()

	ctx := context.Background()
	oA, err := client.Order.Get(ctx, idA)
	if err != nil {
		t.Fatal(err)
	}
	// 余额 = 单笔应付：并发两笔只可能成功一笔。
	setProfileBalance(t, userID, oA.PayAmount)

	var okCount int64
	var wg sync.WaitGroup
	for _, id := range []string{idA, idB} {
		wg.Add(1)
		go func(orderID string) {
			defer wg.Done()
			r, _ := do(t, http.MethodPost, "/api/app/client/orders/"+orderID+"/pay", tok,
				map[string]string{"channel": "balance"})
			if r.Code == 0 {
				atomic.AddInt64(&okCount, 1)
			} else if r.Code != 1902 {
				t.Errorf("并发支付非预期错误码: %d %s", r.Code, r.Message)
			}
		}(id)
	}
	wg.Wait()

	if got := atomic.LoadInt64(&okCount); got != 1 {
		t.Fatalf("余额只够一笔，应恰好 1 笔成功, got %d", got)
	}
	if bal := profileBalance(t, userID); bal < -0.001 {
		t.Fatalf("余额为负（双花）: %f", bal)
	}
	// 成功支付的那笔应扣掉全额：剩余余额 = 初始设定 - 一笔应付 = 0。
	if bal := profileBalance(t, userID); bal > 0.001 {
		t.Fatalf("余额未按实扣减: %f", bal)
	}
}
