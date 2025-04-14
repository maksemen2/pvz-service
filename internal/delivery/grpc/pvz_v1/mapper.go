package pvz_v1

import (
	"github.com/maksemen2/pvz-service/internal/domain/models"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func ConvertToProtoPVZs(pvzs []*models.PVZ) []*PVZ {
	result := make([]*PVZ, 0, len(pvzs))

	for _, p := range pvzs {
		result = append(result, &PVZ{
			Id:               p.ID.String(),
			RegistrationDate: timestamppb.New(p.RegistrationDate),
			City:             p.City.String(),
		})
	}

	return result
}
