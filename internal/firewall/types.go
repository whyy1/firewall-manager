package firewall

// RuleDirection 规则方向
type RuleDirection string

const (
	Inbound  RuleDirection = "in"
	Outbound RuleDirection = "out"
)

// RuleAction 规则动作
type RuleAction string

const (
	Allow RuleAction = "allow"
	Block RuleAction = "block"
)

// FirewallRule 防火墙规则
type FirewallRule struct {
	Name        string       `json:"name"`
	Direction   RuleDirection `json:"direction"`
	Action      RuleAction   `json:"action"`
	Program     string       `json:"program"`
	LocalAddr   string       `json:"localAddr"`
	RemoteAddr  string       `json:"remoteAddr"`
	LocalPort   string       `json:"localPort"`
	RemotePort  string       `json:"remotePort"`
	Protocol    string       `json:"protocol"`
	Enabled     bool         `json:"enabled"`
	Profile     string       `json:"profile"`
}
