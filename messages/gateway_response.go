package messages

type GatewayResponse struct {
	Err        bool `json:"err"`
	SystemPort int  `json:"system_port"`
}
