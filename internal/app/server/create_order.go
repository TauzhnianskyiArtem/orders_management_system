package server

import (
	"context"

	"github.com/moguchev/microservices_courcse/orders_management_system/internal/app/models"
	"github.com/moguchev/microservices_courcse/orders_management_system/internal/app/usecases/orders_management_system"
	pb "github.com/moguchev/microservices_courcse/orders_management_system/pkg/api/orders_management_system"
	grpcutils "github.com/moguchev/microservices_courcse/orders_management_system/pkg/grpc_utils"
)

func (s *Server) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.CreateOrderResponse, error) {
	if err := s.validator.Validate(req); err != nil {
		return nil, grpcutils.RPCValidationError(err)
	}

	createOrderInfo := createOrderInfoFromPbCreateOrderRequest(req)

	order, err := s.OMSUsecase.CreateOrder(ctx, models.UserID(req.GetUserId()), createOrderInfo)
	if err != nil {
		return nil, err
	}

	return &pb.CreateOrderResponse{
		OrderId: order.ID.String(),
	}, nil
}

func createOrderInfoFromPbCreateOrderRequest(req *pb.CreateOrderRequest) orders_management_system.CreateOrderInfo {
	items := make([]models.Item, 0, len(req.GetItems()))
	for _, item := range req.GetItems() {
		items = append(items, models.Item{
			SKU: models.SKU{
				ID: models.SKUID(item.GetId()),
			},
			Quantity:    item.GetQuantity(),
			WarehouseID: models.WarehouseID(item.GetQuantity()),
		})
	}

	deliveryInfo := req.GetDeliveryInfo()

	return orders_management_system.CreateOrderInfo{
		DeliveryOrderInfo: models.DeliveryOrderInfo{
			DeliveryVariantID: models.DeliveryVariantID(deliveryInfo.GetDeliveryVariantId()),
			DeliveryDate:      deliveryInfo.GetDeliveryDate().AsTime(),
		},
		Items: items,
	}
}
