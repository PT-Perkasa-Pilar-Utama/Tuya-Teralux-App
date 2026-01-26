package repositories

import (
	"teralux_app/domain/common/infrastructure"
)

// Deprecated: use infrastructure.VectorService instead. This adapter is kept
// for compatibility with older callsites.
type VectorRepository struct {
	svc *infrastructure.VectorService
}

func NewVectorRepository() *VectorRepository {
	return &VectorRepository{svc: infrastructure.NewVectorService()}
}

func (r *VectorRepository) Search(query string) ([]string, error) {
	return r.svc.Search(query)
}
