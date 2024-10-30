package warehouses_management_system

import (
	"context"
	"time"

	"github.com/moguchev/microservices_courcse/orders_management_system/internal/app/models"
	"github.com/moguchev/microservices_courcse/orders_management_system/pkg/logger"
	"github.com/opentracing/opentracing-go"
)

func (r *Client) ReserveStocks(
	ctx context.Context,
	userID models.UserID,
	items []models.Item,
) error {
	const api = "warehouses_management_system.ReserveStocks"

	span, ctx := opentracing.StartSpanFromContext(ctx, "warehouses_management_system.ReserveStocks")
	defer span.Finish()

	span.SetTag("user_id", userID)

	logger.Info(ctx, "stock reserved")

	time.Sleep(50 * time.Millisecond)

	return nil
}
