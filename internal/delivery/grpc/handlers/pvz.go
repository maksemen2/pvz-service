package grpchandlers

import (
	"context"
	"github.com/maksemen2/pvz-service/internal/delivery/grpc/pvz_v1"
	"github.com/maksemen2/pvz-service/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// PVZServer - gRPC сервер для работы с пунктами выдачи заказов.
type PVZServer struct {
	pvz_v1.UnimplementedPVZServiceServer
	pvzService service.PVZService
}

func NewPVZServer(pvzService service.PVZService) *PVZServer {
	return &PVZServer{
		pvzService: pvzService,
	}
}

// GetPVZList - метод для получения списка пунктов выдачи заказов.
func (h *PVZServer) GetPVZList(ctx context.Context, req *pvz_v1.GetPVZListRequest) (*pvz_v1.GetPVZListResponse, error) {
	pvzs, err := h.pvzService.GetAllPVZs(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pvz_v1.GetPVZListResponse{
		Pvzs: pvz_v1.ConvertToProtoPVZs(pvzs),
	}, nil
}
