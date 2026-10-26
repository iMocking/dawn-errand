package app

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"dawn-errand/internal/application"
	"dawn-errand/internal/domain/finance"
	"dawn-errand/internal/domain/notification"
	"dawn-errand/internal/domain/order"
	"dawn-errand/internal/domain/review"
	"dawn-errand/internal/pkg/errcode"
	"dawn-errand/internal/pkg/metrics"
	"dawn-errand/internal/pkg/ulidgen"
)

// ReviewService 订单评价。
type ReviewService struct {
	Repos *application.RepoUnit
	Tx    application.TxRunner
}

// ReviewInput 评价入参。
type ReviewInput struct {
	Rating  int
	Tags    []string
	Content string
	Images  []string
}

// Submit 提交评价：订单完成 + 商家入账（商品+打包-佣金）+ 结算流水。
func (s *ReviewService) Submit(ctx context.Context, userID, orderID string, in ReviewInput) (*order.Order, *review.Review, error) {
	if in.Rating < 1 || in.Rating > 5 {
		return nil, nil, errcode.New(errcode.CodeBadRequest, "评分需为 1-5")
	}
	in.Content = strings.TrimSpace(in.Content)
	if utf8.RuneCountInString(in.Content) > 1024 {
		return nil, nil, errcode.New(errcode.CodeBadRequest, "评价内容过长")
	}
	if len(in.Tags) > 6 || len(in.Images) > 6 {
		return nil, nil, errcode.New(errcode.CodeBadRequest, "标签/图片最多 6 个")
	}
	for _, t := range in.Tags {
		if utf8.RuneCountInString(t) > 16 {
			return nil, nil, errcode.New(errcode.CodeBadRequest, "标签过长")
		}
	}
	o, err := s.Repos.Orders.FindByID(ctx, orderID)
	if err != nil {
		if errors.Is(err, application.ErrNotFound) {
			return nil, nil, errcode.NewStatus(404, errcode.CodeOrderNotFound, "订单不存在")
		}
		return nil, nil, err
	}
	if o.UserID != userID {
		return nil, nil, errcode.NewStatus(404, errcode.CodeOrderNotFound, "订单不存在")
	}
	if _, err := s.Repos.Reviews.FindByOrder(ctx, orderID); err == nil {
		return nil, nil, errcode.New(errcode.CodeOrderDuplicate, "该订单已评价")
	} else if !errors.Is(err, application.ErrNotFound) {
		return nil, nil, err
	}
	if o.Status != order.StatusPendingReview {
		return nil, nil, errcode.New(errcode.CodeOrderState, "订单当前状态不可评价")
	}

	now := time.Now().Unix()
	rv := &review.Review{
		ID: ulidgen.New(), OrderID: o.ID, UserID: userID,
		MerchantID: o.MerchantID, RiderID: o.RiderID,
		Rating: in.Rating, Tags: in.Tags, Content: in.Content, Images: in.Images,
		CreatedAt: now,
	}
	var merchantIncome float64
	err = s.Tx.RunInTx(ctx, func(tx *application.RepoUnit) error {
		if err := tx.Reviews.Save(ctx, rv); err != nil {
			return err
		}
		ok, err := tx.Orders.UpdateStatus(ctx, o.ID, order.StatusPendingReview, order.OrderDone,
			order.StatusPatch{ReviewedAt: now}, now)
		if err != nil {
			return err
		}
		if !ok {
			return errcode.New(errcode.CodeOrderState, "订单状态已变化")
		}
		if err := tx.OrderEvents.Save(ctx, &order.Event{
			OrderID: o.ID, FromStatus: order.StatusPendingReview, ToStatus: order.OrderDone,
			ActorType: order.ActorClient, ActorID: userID, Note: "用户评价", CreatedAt: now,
		}); err != nil {
			return err
		}
		if o.MerchantID != "" {
			if m, err := tx.Merchants.FindByID(ctx, o.MerchantID); err == nil {
				commission := round2(o.Amount * m.Commission / 100)
				merchantIncome = round2(o.Amount + o.PackingFee - commission)
				if merchantIncome < 0 {
					merchantIncome = 0
				}
				if err := creditWallet(ctx, tx, "merchant", m.ID, merchantIncome, "order_income", o.ID, "订单收入", now); err != nil {
					return err
				}
				_ = tx.Settlements.Save(ctx, &finance.Settlement{
					RegionID: m.RegionID, Party: m.Name, PartyType: "merchant", PartyID: m.ID,
					Type: finance.SettlementMerchant, Amount: merchantIncome,
					Status: finance.SettlementArrived,
					Period: time.Unix(now, 0).UTC().Format("2006-01-02"), SettledAt: now,
					Action: "订单完成自动入账",
				})
				total := float64(m.ReviewCount)
				m.Rating = round2((m.Rating*total + float64(in.Rating)) / (total + 1))
				m.ReviewCount++
				m.MonthlyOrders++
				m.MonthlyRevenue = round2(m.MonthlyRevenue + merchantIncome)
				m.LastUpdatedAt = now
				if err := tx.Merchants.Save(ctx, m); err != nil {
					return err
				}
				_ = tx.Notifications.Save(ctx, &notification.Notification{
					UserID: m.OwnerUserID, Title: "收到新评价",
					Content: "订单 " + o.OrderNo + " 收到 " + fmt.Sprint(in.Rating) + " 星评价",
					Type:    "order", RefID: o.ID, CreatedAt: now,
				})
			}
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	metrics.OrdersCompletedTotal.Inc()
	updated, err := s.Repos.Orders.FindByID(ctx, o.ID)
	if err != nil {
		return nil, nil, err
	}
	return updated, rv, nil
}

// ListStore 商家评价列表。
func (s *ReviewService) ListStore(ctx context.Context, merchantID string, page, size int) ([]*review.Review, int64, error) {
	page, size = normalizePageParams(page, size)
	return s.Repos.Reviews.List(ctx, review.Query{MerchantID: merchantID, Page: page, PageSize: size})
}

// Reply 商家回复评价。
func (s *ReviewService) Reply(ctx context.Context, merchantID, reviewID, content string) error {
	content = strings.TrimSpace(content)
	if content == "" || utf8.RuneCountInString(content) > 1024 {
		return errcode.New(errcode.CodeBadRequest, "回复内容需为 1-1024 字")
	}
	ok, err := s.Repos.Reviews.Reply(ctx, reviewID, merchantID, content, time.Now().Unix())
	if err != nil {
		return err
	}
	if !ok {
		return errcode.NewStatus(404, errcode.CodeNotFound, "评价不存在")
	}
	return nil
}

// ListDriver 骑手评价列表（M4 可选）。
func (s *ReviewService) ListDriver(ctx context.Context, riderID string, page, size int) ([]*review.Review, int64, error) {
	page, size = normalizePageParams(page, size)
	return s.Repos.Reviews.List(ctx, review.Query{RiderID: riderID, Page: page, PageSize: size})
}
