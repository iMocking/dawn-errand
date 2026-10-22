package repo

import (
	"context"

	"dawn-errand/internal/application"
	"dawn-errand/internal/infrastructure/ent"
)

// AuditRepo 实现 application.AuditLogger。
type AuditRepo struct{ client *ent.Client }

// Record 记录审计日志（不阻塞业务；失败仅记 error 不向上抛）。
func (r AuditRepo) Record(ctx context.Context, e application.AuditEntry) {
	_, err := r.client.AuditLog.Create().
		SetUserID(e.UserID).SetAction(e.Action).SetObject(e.Object).
		SetResult(e.Result).SetRequestID(e.RequestID).
		SetOperator(e.Operator).SetIP(e.IP).
		Save(ctx)
	if err != nil {
		// 审计失败不影响主流程；日志由调用方负责。
		_ = err
	}
}
