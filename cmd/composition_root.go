package cmd

import (
	httpadapter "delivery/internal/adapters/in/http"
	kafkain "delivery/internal/adapters/in/kafka"
	grpcgeo "delivery/internal/adapters/out/grpc/geo"
	kafkaout "delivery/internal/adapters/out/kafka"
	"delivery/internal/adapters/out/postgres"
	"delivery/internal/adapters/out/postgres/courierrepo"
	"delivery/internal/adapters/out/postgres/inboxrepo"
	"delivery/internal/adapters/out/postgres/orderrepo"
	"delivery/internal/adapters/out/postgres/outboxrepo"
	"delivery/internal/core/application/eventhandlers"
	"delivery/internal/core/application/usecases/commands"
	"delivery/internal/core/application/usecases/queries"
	"delivery/internal/core/domain/model/order"
	"delivery/internal/core/domain/services"
	"delivery/internal/core/ports"
	"delivery/internal/jobs"
	"delivery/internal/pkg/ddd"
	"delivery/internal/pkg/inbox"
	"delivery/internal/pkg/outbox"
	"fmt"
	"log"
	"reflect"

	postgresgorm "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type CompositionRoot struct {
	configs Config

	closers []Closer
}

func NewCompositionRoot(configs Config) *CompositionRoot {
	return &CompositionRoot{
		configs: configs,
	}
}

func (cr *CompositionRoot) NewOrderDispatcher() services.OrderDispatcher {
	return services.NewOrderDispatcher()
}

func (cr *CompositionRoot) NewOrderRepository() ports.OrderRepository {
	return orderrepo.NewRepository()
}

func (cr *CompositionRoot) NewCourierRepository() ports.CourierRepository {
	return courierrepo.NewRepository()
}

func (cr *CompositionRoot) NewUnitOfWork(db *gorm.DB) (ports.UnitOfWork, error) {
	outboxRepository, err := cr.NewOutboxRepository(db)
	if err != nil {
		return nil, err
	}
	return postgres.NewUnitOfWork(db, outboxRepository)
}

func (cr *CompositionRoot) NewOutboxRepository(db *gorm.DB) (ports.OutboxRepository, error) {
	return outboxrepo.NewRepository(db)
}

func (cr *CompositionRoot) NewInboxRepository() ports.InboxRepository {
	return inboxrepo.NewRepository()
}

func (cr *CompositionRoot) NewGeoClient() (ports.GeoClient, error) {
	client, err := grpcgeo.NewClient(cr.configs.GeoServiceGrpcHost)
	if err != nil {
		return nil, err
	}
	cr.RegisterCloser(client)
	return client, nil
}

func (cr *CompositionRoot) OpenPostgres() (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cr.configs.DbHost,
		cr.configs.DbPort,
		cr.configs.DbUser,
		cr.configs.DbPassword,
		cr.configs.DbName,
		cr.configs.DbSslMode,
	)
	db, err := gorm.Open(postgresgorm.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	if err = db.AutoMigrate(&orderrepo.OrderDTO{}, &courierrepo.CourierDTO{}, &courierrepo.StoragePlaceDTO{}, &outbox.Message{}, &inbox.Message{}); err != nil {
		return nil, err
	}
	return db, nil
}

func (cr *CompositionRoot) NewCreateOrderCommandHandler(db *gorm.DB) (commands.CreateOrderCommandHandler, error) {
	unitOfWork, err := cr.NewUnitOfWork(db)
	if err != nil {
		return nil, err
	}
	geoClient, err := cr.NewGeoClient()
	if err != nil {
		return nil, err
	}
	return commands.NewCreateOrderCommandHandler(unitOfWork, cr.NewOrderRepository(), geoClient)
}

func (cr *CompositionRoot) NewProcessBasketConfirmedCommandHandler(db *gorm.DB) (commands.ProcessBasketConfirmedCommandHandler, error) {
	unitOfWork, err := cr.NewUnitOfWork(db)
	if err != nil {
		return nil, err
	}
	geoClient, err := cr.NewGeoClient()
	if err != nil {
		return nil, err
	}
	return commands.NewProcessBasketConfirmedCommandHandler(
		unitOfWork,
		cr.NewInboxRepository(),
		cr.NewOrderRepository(),
		geoClient,
	)
}

func (cr *CompositionRoot) NewCreateCourierCommandHandler(db *gorm.DB) (commands.CreateCourierCommandHandler, error) {
	unitOfWork, err := cr.NewUnitOfWork(db)
	if err != nil {
		return nil, err
	}
	return commands.NewCreateCourierCommandHandler(unitOfWork, cr.NewCourierRepository())
}

func (cr *CompositionRoot) NewAddStoragePlaceCommandHandler(db *gorm.DB) (commands.AddStoragePlaceCommandHandler, error) {
	unitOfWork, err := cr.NewUnitOfWork(db)
	if err != nil {
		return nil, err
	}
	return commands.NewAddStoragePlaceCommandHandler(unitOfWork, cr.NewCourierRepository())
}

func (cr *CompositionRoot) NewMoveCouriersCommandHandler(db *gorm.DB) (commands.MoveCouriersCommandHandler, error) {
	unitOfWork, err := cr.NewUnitOfWork(db)
	if err != nil {
		return nil, err
	}
	return commands.NewMoveCouriersCommandHandler(unitOfWork, cr.NewOrderRepository(), cr.NewCourierRepository())
}

func (cr *CompositionRoot) NewMoveCourierCommandHandler(db *gorm.DB) (commands.MoveCourierCommandHandler, error) {
	unitOfWork, err := cr.NewUnitOfWork(db)
	if err != nil {
		return nil, err
	}
	return commands.NewMoveCourierCommandHandler(unitOfWork, cr.NewCourierRepository())
}

func (cr *CompositionRoot) NewCompleteOrderCommandHandler(db *gorm.DB) (commands.CompleteOrderCommandHandler, error) {
	unitOfWork, err := cr.NewUnitOfWork(db)
	if err != nil {
		return nil, err
	}
	return commands.NewCompleteOrderCommandHandler(unitOfWork, cr.NewOrderRepository(), cr.NewCourierRepository())
}

func (cr *CompositionRoot) NewAssignOrderCommandHandler(db *gorm.DB) (commands.AssignOrderCommandHandler, error) {
	unitOfWork, err := cr.NewUnitOfWork(db)
	if err != nil {
		return nil, err
	}
	return commands.NewAssignOrderCommandHandler(unitOfWork, cr.NewOrderRepository(), cr.NewCourierRepository(), cr.NewOrderDispatcher())
}

func (cr *CompositionRoot) NewGetAllCouriersQueryHandler(db *gorm.DB) (queries.GetAllCouriersQueryHandler, error) {
	unitOfWork, err := cr.NewUnitOfWork(db)
	if err != nil {
		return nil, err
	}
	return queries.NewGetAllCouriersQueryHandler(unitOfWork, cr.NewCourierRepository())
}

func (cr *CompositionRoot) NewGetBusyCouriersQueryHandler(db *gorm.DB) (queries.GetBusyCouriersQueryHandler, error) {
	return cr.NewGetAllCouriersQueryHandler(db)
}

func (cr *CompositionRoot) NewGetNotCompletedOrdersQueryHandler(db *gorm.DB) (queries.GetNotCompletedOrdersQueryHandler, error) {
	unitOfWork, err := cr.NewUnitOfWork(db)
	if err != nil {
		return nil, err
	}
	return queries.NewGetNotCompletedOrdersQueryHandler(unitOfWork, cr.NewOrderRepository())
}

func (cr *CompositionRoot) NewHTTPHandler(db *gorm.DB) (*httpadapter.Handler, error) {
	createOrder, err := cr.NewCreateOrderCommandHandler(db)
	if err != nil {
		return nil, err
	}
	createCourier, err := cr.NewCreateCourierCommandHandler(db)
	if err != nil {
		return nil, err
	}
	moveCourier, err := cr.NewMoveCourierCommandHandler(db)
	if err != nil {
		return nil, err
	}
	completeOrder, err := cr.NewCompleteOrderCommandHandler(db)
	if err != nil {
		return nil, err
	}
	allCouriers, err := cr.NewGetAllCouriersQueryHandler(db)
	if err != nil {
		return nil, err
	}
	activeOrders, err := cr.NewGetNotCompletedOrdersQueryHandler(db)
	if err != nil {
		return nil, err
	}
	return httpadapter.NewHandler(createOrder, createCourier, moveCourier, completeOrder, allCouriers, activeOrders)
}

func (cr *CompositionRoot) NewDeliveryJobs(db *gorm.DB) (jobs.DeliveryJobs, error) {
	moveCouriers, err := cr.NewMoveCouriersCommandHandler(db)
	if err != nil {
		return nil, err
	}
	assignOrder, err := cr.NewAssignOrderCommandHandler(db)
	if err != nil {
		return nil, err
	}
	outboxJob, err := cr.NewOutboxJob(db)
	if err != nil {
		return nil, err
	}
	deliveryJobs, err := jobs.NewDeliveryJobs(moveCouriers, assignOrder, outboxJob)
	if err != nil {
		return nil, err
	}
	cr.RegisterCloser(deliveryJobs)
	return deliveryJobs, nil
}

func (cr *CompositionRoot) NewBasketConfirmedConsumer(db *gorm.DB) (kafkain.BasketConfirmedConsumer, error) {
	processBasketConfirmed, err := cr.NewProcessBasketConfirmedCommandHandler(db)
	if err != nil {
		return nil, err
	}
	consumer, err := kafkain.NewBasketConfirmedConsumer(
		[]string{cr.configs.KafkaHost},
		cr.configs.KafkaConsumerGroup,
		cr.configs.KafkaBasketEventsTopic,
		processBasketConfirmed,
	)
	if err != nil {
		return nil, err
	}
	cr.RegisterCloser(consumer)
	return consumer, nil
}

func (cr *CompositionRoot) NewOrderProducer() (ports.OrderProducer, error) {
	producer, err := kafkaout.NewOrderProducer(
		[]string{cr.configs.KafkaHost},
		cr.configs.KafkaOrderEventsTopic,
	)
	if err != nil {
		return nil, err
	}
	cr.RegisterCloser(producer)
	return producer, nil
}

func (cr *CompositionRoot) NewMediatr() (ddd.Mediatr, error) {
	producer, err := cr.NewOrderProducer()
	if err != nil {
		return nil, err
	}

	publishCompleted, err := eventhandlers.NewPublishOrderCompleted(producer)
	if err != nil {
		return nil, err
	}

	mediatr := ddd.NewMediatr()
	mediatr.Subscribe(publishCompleted, order.NewEmptyCompletedDomainEvent())
	return mediatr, nil
}

func (cr *CompositionRoot) NewOutboxJob(db *gorm.DB) (jobs.OutboxJob, error) {
	unitOfWork, err := cr.NewUnitOfWork(db)
	if err != nil {
		return nil, err
	}
	repository, err := cr.NewOutboxRepository(db)
	if err != nil {
		return nil, err
	}
	registry, err := outbox.NewEventRegistry()
	if err != nil {
		return nil, err
	}
	if err = registry.RegisterDomainEvent(reflect.TypeOf(order.CompletedDomainEvent{})); err != nil {
		return nil, err
	}
	mediatr, err := cr.NewMediatr()
	if err != nil {
		return nil, err
	}

	return jobs.NewOutboxJob(unitOfWork, repository, registry, mediatr)
}

///////////////////////////////////////////////////////////
//////////////////// LIFECYCLE ////////////////////////////
///////////////////////////////////////////////////////////

func (cr *CompositionRoot) RegisterCloser(c Closer) {
	cr.closers = append(cr.closers, c)
}

func (cr *CompositionRoot) CloseAll() {
	for _, c := range cr.closers {
		if err := c.Close(); err != nil {
			log.Printf("close error: %v", err)
		}
	}
}
