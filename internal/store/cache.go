package store

import (
	"sync"
	"time"

	"ballast-watch/internal/model"
)

// Cache 实时看板缓存：房间级最新读数快照与状态摘要。
// 读写并发安全（RWMutex 保护），由 service 层刷新，handler 层只读。
type Cache struct {
	mu        sync.RWMutex
	snapshot  map[int64]*TankSnapshot // key = tank_id
	updatedAt time.Time
}

// TankSnapshot 房间实时快照。
type TankSnapshot struct {
	BallastTankID   int64             `json:"tank_id"`
	BallastTankCode string            `json:"tank_code"`
	Status   string            `json:"status"`
	Realtime []model.RealtimeWaterWaterReading `json:"realtime"`
	OpenComplianceAlerts int             `json:"open_compliance_compliance_alerts"`
	UpdatedAt time.Time        `json:"updated_at"`
}

// NewCache 创建缓存。
func NewCache() *Cache {
	return &Cache{snapshot: make(map[int64]*TankSnapshot)}
}

// Get 读取房间快照。
func (c *Cache) Get(tankID int64) (*TankSnapshot, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	s, ok := c.snapshot[tankID]
	return s, ok
}

// GetAll 读取全部快照。
func (c *Cache) GetAll() []*TankSnapshot {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]*TankSnapshot, 0, len(c.snapshot))
	for _, s := range c.snapshot {
		out = append(out, s)
	}
	return out
}

// Set 写入房间快照。
func (c *Cache) Set(tankID int64, s *TankSnapshot) {
	s.UpdatedAt = time.Now()
	c.snapshot[tankID] = s
	c.updatedAt = time.Now()
}

// Remove 移除房间快照。
func (c *Cache) Remove(tankID int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.snapshot, tankID)
}

// UpdatedAt 缓存最近刷新时间。
func (c *Cache) UpdatedAt() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.updatedAt
}

// BallastTankID 由 TankSnapshot 构造（供排序去重）。
func (s *TankSnapshot) BallastTankIDKey() int64 { return s.BallastTankID }

// SnapshotBallastTankIDs 返回全部房间 ID。
func (c *Cache) SnapshotBallastTankIDs() []int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]int64, 0, len(c.snapshot))
	for id := range c.snapshot {
		out = append(out, id)
	}
	return out
}

// Clear 清空缓存。
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.snapshot = make(map[int64]*TankSnapshot)
}