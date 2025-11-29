package service

import (
	"context"

	"github.com/aleffnull/shortener/internal/pkg/audit"
	"github.com/aleffnull/shortener/internal/pkg/logger"
)

type AuditService interface {
	Init()
	Shutdown()
	AuditEvent(event *audit.Event)
}

type auditServiceImpl struct {
	receivers   []audit.Receiver
	logger      logger.Logger
	auditCh     chan *audit.Event
	stopWorkers context.CancelFunc
}

// Будем считать, что сервис не нагружен, а воркеры аудита работают быстро,
// поэтому канал переполняться не будет. В идеале, здесь нужно использовать outbox,
// но для одного из четырех инкрементов спринта это перебор.
const (
	channelSize  = 1000
	workersCount = 10
)

var _ AuditService = (*auditServiceImpl)(nil)

func NewAuditService(receivers []audit.Receiver, logger logger.Logger) AuditService {
	return &auditServiceImpl{
		receivers: receivers,
		logger:    logger,
		auditCh:   make(chan *audit.Event, channelSize),
	}
}

func (i *auditServiceImpl) Init() {
	ctx, cancel := context.WithCancel(context.Background())
	i.stopWorkers = cancel
	for j := range workersCount {
		go i.auditWorker(ctx, j)
	}
}

func (i *auditServiceImpl) Shutdown() {
	i.stopWorkers()
}

func (i *auditServiceImpl) AuditEvent(event *audit.Event) {
	i.auditCh <- event
}

func (i *auditServiceImpl) auditWorker(ctx context.Context, j int) {
	i.logger.Infof("Start audit worker %v", j)

	for {
		select {
		case <-ctx.Done():
			i.logger.Infof("Stop audit worker %v", j)
			return
		case event := <-i.auditCh:
			i.logger.Infof("Process audit event by worker %v", j)
			for _, receiver := range i.receivers {
				if err := receiver.AddEvent(event); err != nil {
					i.logger.Errorf("Audit error in worker %v: %v", j, err)
				}
			}
		}
	}
}
