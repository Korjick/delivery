package jobs

import (
	"context"
	"delivery/internal/core/application/usecases/commands"
	"delivery/internal/pkg/errs"
	"log"
	"time"

	"github.com/go-co-op/gocron/v2"
)

type DeliveryJobs interface {
	Close() error
}

type deliveryJobs struct {
	scheduler gocron.Scheduler
}

func NewDeliveryJobs(
	moveCouriers commands.MoveCouriersCommandHandler,
	assignOrder commands.AssignOrderCommandHandler,
	outbox OutboxJob,
) (DeliveryJobs, error) {
	if moveCouriers == nil {
		return nil, errs.NewValueIsRequired("moveCouriers")
	}
	if assignOrder == nil {
		return nil, errs.NewValueIsRequired("assignOrder")
	}
	if outbox == nil {
		return nil, errs.NewValueIsRequired("outbox")
	}

	scheduler, err := gocron.NewScheduler()
	if err != nil {
		return nil, err
	}
	if _, err = scheduler.NewJob(
		gocron.DurationJob(time.Second),
		gocron.NewTask(runMoveCouriers, moveCouriers),
	); err != nil {
		return nil, err
	}
	if _, err = scheduler.NewJob(
		gocron.DurationJob(time.Second),
		gocron.NewTask(runAssignOrder, assignOrder),
	); err != nil {
		return nil, err
	}
	if _, err = scheduler.NewJob(
		gocron.DurationJob(time.Second),
		gocron.NewTask(runOutbox, outbox),
		gocron.WithSingletonMode(gocron.LimitModeReschedule),
	); err != nil {
		return nil, err
	}
	scheduler.Start()
	return &deliveryJobs{scheduler: scheduler}, nil
}

func (j *deliveryJobs) Close() error {
	return j.scheduler.Shutdown()
}

func runMoveCouriers(handler commands.MoveCouriersCommandHandler) {
	if err := handler.Handle(context.Background(), commands.MoveCouriersCommand{}); err != nil {
		log.Printf("move couriers job failed: %v", err)
	}
}

func runAssignOrder(handler commands.AssignOrderCommandHandler) {
	if err := handler.Handle(context.Background(), commands.AssignOrderCommand{}); err != nil {
		log.Printf("assign order job failed: %v", err)
	}
}

func runOutbox(job OutboxJob) {
	if err := job.Run(); err != nil {
		log.Printf("outbox job failed: %v", err)
	}
}
