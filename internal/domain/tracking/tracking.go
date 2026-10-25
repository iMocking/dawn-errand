// Package tracking 骑手实时位置与配送追踪。
package tracking

// Location 骑手最新位置（GCJ-02）。
type Location struct {
	ID        string
	RiderID   string
	RegionID  string
	Lat       float64
	Lng       float64
	Heading   float64
	Speed     float64
	Accuracy  float64
	UpdatedAt int64
}

// Repository 位置仓储端口。
type Repository interface {
	Upsert(l *Location) error
	FindByRider(riderID string) (*Location, error)
	FindByRiders(riderIDs []string) (map[string]*Location, error)
	// DeleteStale 清理长期未更新的位置。
	DeleteStale(before int64) error
}
