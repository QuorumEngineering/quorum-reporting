package monitor

import (
	"quorumengineering/quorum-report/client"
	"quorumengineering/quorum-report/database"
)

// MonitorService starts all monitors. It pulls data from Quorum node and update the database.
type MonitorService struct {
	blockMonitor *BlockMonitor
}

func NewMonitorService(db database.Database, quorumClient client.Client) *MonitorService {
	return &MonitorService{
		NewBlockMonitor(db, quorumClient),
	}
}

func (m *MonitorService) Start() error {
	// BlockMonitor will sync all new blocks and historical blocks.
	// It will invoke TransactionMonitor internally.
	return m.blockMonitor.Start()
}
