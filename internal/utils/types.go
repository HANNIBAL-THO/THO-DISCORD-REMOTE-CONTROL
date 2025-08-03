package utils

type NetworkInfo struct {
	PublicIP    string   `json:"public_ip"`
	LocalIP     string   `json:"local_ip"`
	DNSServers  []string `json:"dns_servers"`
	Connections []string `json:"connections"`
}
