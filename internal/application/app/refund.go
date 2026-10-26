package app

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"dawn-errand/internal/application"
	"dawn-errand/internal/domain/notification"
	"dawn-errand/internal/domain/order"
	"dawn-errand/internal/domain/refund"
	"dawn-errand/internal/pkg/errcode"
	"dawn-errand/internal/pkg/metrics"
	"dawn-errand/internal/pkg/ulidgen"
)

// RefundService 退款/售后。
type RefundService struct {
	Repos *application.RepoUnit
	Tx    application.TxRunner
}

// Apply 用户申请售后（整单退款；M4 不支持部分退款）。
func (s *RefundService) Apply(ctx context.Context, userID, orderID, reason string, amount float64) (*refund.Refund, error) {
	o, err := s.Repos.Orders.FindByID(ctx, orderID)
	if err != nil {
		if errors.Is(err, application.ErrNotFound) {
			return nil, errcode.NewStatus(404, errcode.CodeOrderNotFound, "订单不存在")
		}
		return nil, err
	}
	if o.UserID != userID {
		return nil, errcode.NewStatus(404, errcode.CodeOrderNotFound, "订单不存在")
	}
	switch o.Status {
	case order.StatusDelivered, order.StatusPendingReview, order.OrderDone:
	default:
		return nil, errcode.New(errcode.CodeOrderState, "该订单当前不可申请售后")
	}
	if existing, err := s.Repos.Refunds.FindByOrder(ctx, o.ID); err == nil {
		if existing.Status == refund.StatusRequested || existing.Status == refund.StatusApproved || existing.Status == refund.StatusRefunded {
			return nil, errcode.New(errcode.CodeOrderDuplicate, "售后申请已提交")
		}
	} else if !errors.Is(err, application.ErrNotFound) {
		return nil, err
	}
	if amount <= 0 {
		amount = o.PayAmount
	}
	if round2(amount) != round2(o.PayAmount) {
		return nil, errcode.New(errcode.CodeAmountMismatch, "暂仅支持整单退款")
	}
	reason = strings.TrimSpace(reason)
	if reason == "" {
		reason = "用户申请售后"
	}
	now := time.Now().Unix()
	r := &refund.Refund{
		ID: ulidgen.New(), OrderID: o.ID, UserID: userID, MerchantID: o.MerchantID,
		Amount: round2(amount), Reason: reason, Status: refund.StatusRequested,
		PrevStatus: string(o.Status), CreatedAt: now,
	}
	err = s.Tx.RunInTx(ctx, func(tx *application.RepoUnit) error {
		if err := tx.Refunds.Save(ctx, r); err != nil {
			return err
		}
		ok, err := tx.Orders.UpdateStatus(ctx, o.ID, o.Status, order.StatusAfterSale, order.StatusPatch{}, now)
		if err != nil {
			return err
		}
		if !ok {
			return errcode.New(errcode.CodeOrderState, "订单状态已变化")
		}
		if err := tx.OrderEvents.Save(ctx, &order.Event{
			OrderID: o.ID, FromStatus: o.Status, ToStatus: order.StatusAfterSale,
			ActorType: order.ActorClient, ActorID: userID, Note: "申请售后：" + reason, CreatedAt: now,
		}); err != nil {
			return err
		}
		if o.MerchantID != "" {
			if m, err := tx.Merchants.FindByID(ctx, o.MerchantID); err == nil {
				_ = tx.Notifications.Save(ctx, &notification.Notification{
					UserID: m.OwnerUserID, Title: "收到售后申请",
					Content: "订单 " + o.OrderNo + " 申请退款：" + reason,
					Type:    "order", RefID: o.ID, CreatedAt: now,
				})
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return r, nil
}

// Decide 商家处理售后（approve/reject）。
func (s *RefundService) Decide(ctx context.Context, merchantID, orderID string, approve bool, reason string) (*order.Order, error) {
	o, err := s.Repos.Orders.FindByID(ctx, orderID)
	if err != nil {
		if errors.Is(err, application.ErrNotFound) {
			return nil, errcode.NewStatus(404, errcode.CodeOrderNotFound, "订单不存在")
		}
		return nil, err
	}
	if o.MerchantID != merchantID {
		return nil, errcode.NewStatus(404, errcode.CodeOrderNotOwned, "订单不存在")
	}
	r, err := s.Repos.Refunds.FindByOrder(ctx, o.ID)
	if err != nil {
		return nil, errcode.NewStatus(404, errcode.CodeNotFound, "售后申请不存在")
	}
	if r.Status != refund.StatusRequested {
		return nil, errcode.New(errcode.CodeOrderState, "该售后申请已处理")
	}
	now := time.Now().Unix()
	reason = strings.TrimSpace(reason)

	err = s.Tx.RunInTx(ctx, func(tx *application.RepoUnit) error {
		if !approve {
			if err := tx.Refunds.Update(ctx, r.ID, refund.StatusRejected, "", reason, now); err != nil {
				return err
			}
			prev := order.OrderStatus(r.PrevStatus)
			if prev == "" {
				prev = order.StatusPendingReview
			}
			ok, err := tx.Orders.UpdateStatus(ctx, o.ID, order.StatusAfterSale, prev, order.StatusPatch{}, now)
			if err != nil {
				return err
			}
			if !ok {
				return errcode.New(errcode.CodeOrderState, "订单状态已变化")
			}
			if err := tx.OrderEvents.Save(ctx, &order.Event{
				OrderID: o.ID, FromStatus: order.StatusAfterSale, ToStatus: prev,
				ActorType: order.ActorStore, ActorID: merchantID, Note: "售后驳回：" + reason, CreatedAt: now,
			}); err != nil {
				return err
			}
			return s.notifyUser(ctx, tx, o.UserID, "售后申请被驳回", "订单 "+o.OrderNo+" 售后被驳回："+reason, o.ID, now)
		}

		// 退款：退回用户余额 + 支付单 + 商家/骑手入账冲回。
		if err := creditBalance(ctx, tx, o.UserID, r.Amount, now); err != nil {
			return err
		}
		if p, err := tx.Payments.FindByOrder(ctx, o.ID); err == nil && p.Status == order.PaymentPaid {
			if err := tx.Payments.UpdateStatus(ctx, p.ID, order.PaymentRefunded, "", now); err != nil {
				return err
			}
		}
		if o.CouponID != "" {
			_ = tx.Coupons.ReleaseByOrder(ctx, o.ID, now)
		}
		if r.PrevStatus == string(order.OrderDone) && o.MerchantID != "" {
			if m, err := tx.Merchants.FindByID(ctx, o.MerchantID); err == nil {
				commission := round2(o.Amount * m.Commission / 100)
				income := round2(o.Amount + o.PackingFee - commission)
				if income > 0 {
					if _, err := debitWallet(ctx, tx, "merchant", m.ID, income, true,
						"refund_reversal", o.ID, "售后退款冲回", now); err != nil {
						return err
					}
				}
			}
		}
		if o.RiderID != "" && o.RiderEarning > 0 {
			if _, err := debitWallet(ctx, tx, "rider", o.RiderID, o.RiderEarning, true,
				"refund_reversal", o.ID, "售后退款冲回", now); err != nil {
				return err
			}
		}
		ok, err := tx.Orders.UpdateStatus(ctx, o.ID, order.StatusAfterSale, order.StatusRefunded,
			order.StatusPatch{PaymentStatus: "refunded"}, now)
		if err != nil {
			return err
		}
		if !ok {
			return errcode.New(errcode.CodeOrderState, "订单状态已变化")
		}
		if err := tx.Refunds.Update(ctx, r.ID, refund.StatusRefunded, merchantID, "", now); err != nil {
			return err
		}
		if err := tx.OrderEvents.Save(ctx, &order.Event{
			OrderID: o.ID, FromStatus: order.StatusAfterSale, ToStatus: order.StatusRefunded,
			ActorType: order.ActorStore, ActorID: merchantID, Note: "售后退款完成", CreatedAt: now,
		}); err != nil {
			return err
		}
		return s.notifyUser(ctx, tx, o.UserID, "退款已完成", "订单 "+o.OrderNo+" 退款 ¥"+formatMoney(r.Amount)+" 已退回余额", o.ID, now)
	})
	if err != nil {
		return nil, err
	}
	metrics.RefundsTotal.Inc()
	return s.Repos.Orders.FindByID(ctx, o.ID)
}

func (s *RefundService) notifyUser(ctx context.Context, tx *application.RepoUnit, userID, title, content, orderID string, now int64) error {
	if userID == "" {
		return nil
	}
	return tx.Notifications.Save(ctx, &notification.Notification{
		UserID: userID, Title: title, Content: content, Type: "order", RefID: orderID, CreatedAt: now,
	})
}

func formatMoney(v float64) string {
	return fmt.Sprintf("%.2f", v)
}
