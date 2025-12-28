package config

import "fmt"

type Pools struct {
	Servers []Server `yaml:"servers"`
	DB      DB       `yaml:"db"`
}

type Server struct {
	ID       int    `yaml:"id"`
	Location string `yaml:"location"`
	APIAddr  string `yaml:"api_addr"`
	APIPort  int    `yaml:"api_port"`
	DownTime int    `yaml:"down_time"`
	Active   bool   `yaml:"active"`
}

type DB struct {
	Root string `yaml:"root"`
	Info []Info `yaml:"info"`
}

type Info struct {
	Name    string  `yaml:"name"`
	Pass    Pass    `yaml:"pass"`
	Profile Profile `yaml:"profile"`
}

type Pass struct {
	TLS string `yaml:"tls"`
	SS  string `yaml:"ss"`
}

type Profile struct {
	Link    string `yaml:"link"`
	WebRoot string `yaml:"web_root"`
	Sample  string `yaml:"sample"`
}

func (p *Pools) Save(poolID int) error {
	poolFile := fmt.Sprintf("%s/%d.yaml", PoolPath, poolID)
	Log.Info("poolFile", "=", poolFile)
	return SaveYaml(poolFile, p)
}

var pools *Pools

func (p *Pools) Load(poolID int) (*Pools, error) {
	if poolID == 0 {
		return nil, fmt.Errorf("Load(poolID=%d) poolID cannot be zero", poolID)
	}

	poolFile := fmt.Sprintf("%s/%d.yaml", PoolPath, poolID)
	if pools != nil {
		Log.Warn("Load pool file ignored (already loaded)", "=", poolFile)
		return pools, nil
	}

	err := LoadYaml(poolFile, p)
	if err != nil {
		return nil, err
	}

	pools = p
	Log.Info("Load pool file", "=", poolFile)
	return pools, nil
}

func (sr *Server) Addr(path string) string {
	switch path {
	case "stats":
		return fmt.Sprintf("http://%s:%d/%s", sr.APIAddr, sr.APIPort, SsmApiPathStats)
	case "users":
		return fmt.Sprintf("http://%s:%d/%s", sr.APIAddr, sr.APIPort, SsmApiPathUsers)
	}
	return fmt.Sprintf("http://%s:%d/%s", sr.APIAddr, sr.APIPort, SsmApiPath)
}
