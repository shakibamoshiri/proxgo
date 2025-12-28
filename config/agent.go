package config

import "fmt"

type Agents struct {
	Agent   Agent   `yaml:"agent"`
	Account Account `yaml:"account"`
	Pools   []Pool  `yaml:"pools"`
	Cost    []Cost  `yaml:"cost"`
}

type Agent struct {
	ID        int    `yaml:"id"`
	RealName  string `yaml:"realname"`
	GroupName string `yaml:"groupname"`
	Phone     string `yaml:"phone"`
	PoolID    int    `yaml:"pool_id"`
	BotToken  string `yaml:"bot_token"`
	BotChatID int64  `yaml:"bot_chat_id"`
}

type Account struct {
	Wallet int64 `yaml:"wallet"`
	Bill   int64 `yaml:"bill"`
	Create int64 `yaml:"create"`
	Renew  int64 `yaml:"renew"`
}

type Pool struct {
	ID       int   `yaml:"id"`
	Period   int64 `yaml:"period"`
	Traffic  int64 `yaml:"traffic"`
	Sessions int   `yaml:"sessions"`
	Capacity int   `yaml:"capacity"`
	Servers  int   `yaml:"servers"`
}

type Cost struct {
	ID      int `yaml:"id"`
	Period  int `yaml:"period"`
	Traffic int `yaml:"traffic"`
	Other   int `yaml:"other"`
}

func (ag *Agents) Save() error {
	agentFile := fmt.Sprintf("%s/%d.yaml", AgentPath, AgentID)
	Log.Info("agentFile", "=", agentFile)
	return SaveYaml(agentFile, ag)
}

var agents *Agents

func (ag *Agents) Load() (*Agents, error) {
	agentFile := fmt.Sprintf("%s/%d.yaml", AgentPath, AgentID)

	if agents != nil {
		Log.Warn("Load agent file ignored (already loaded)", "=", agentFile)
		return agents, nil
	}

	err := LoadYaml(agentFile, ag)
	if err != nil {
		return nil, err
	}

	Log.Info("Load agent file", "=", agentFile)
	agents = ag
	return ag, nil
}
