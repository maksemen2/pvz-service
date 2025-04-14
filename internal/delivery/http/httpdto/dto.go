package httpdto

type ReceptionWithProductsResponse struct {
	Reception *Reception `json:"reception"`
	Products  []*Product `json:"products"`
}
type PVZWithReceptionsResponse struct {
	PVZ        *PVZ                             `json:"pvz"`
	Receptions []*ReceptionWithProductsResponse `json:"receptions"`
}
