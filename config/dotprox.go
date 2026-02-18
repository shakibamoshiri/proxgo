package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

var yamlGlobal = &YamlFiles{}

type YamlFiles struct {
	Pools
	Agents
	poolsLoaded  bool
	agentsLoaded bool
}

func NewYamlFile () *YamlFiles {
    return yamlGlobal
}

func GetYamlConfig() (*Agents, *Pools) {
    return agents, pools
}

var (
	// find the target pools id
	activePoolIndex = -1
	// find the target info id
	activeInfoIndex = -1
)

func SaveYaml(path string, content any) (err error) {
	Log.Info("SaveYaml path", "=", path)
	dir := filepath.Dir(path)

	tmp, err := os.CreateTemp(dir, ".tmp-*.yaml")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	Log.Info("os.CreateTemp", "tmpName", tmpName)

	defer func() {
		errClose := tmp.Close()
		if err != nil {
			err = errClose
			return
		}
		/* for learning
		   /* when os.Rename is called, the file is gone (moved to a new one)
		   /* we cannot not remove it, because it does not exists
		   errRemove := os.Remove(tmpName)
		   if errRemove != nil {
		       err = errRemove
		       return
		   }
		*/
	}()

	// do we need this ?
	// tmp.Sync()

	err = yaml.NewEncoder(tmp).Encode(content)
	if err != nil {
		return err
	}

	err = os.Rename(tmpName, path)
	return
}

func LoadYaml(path string, holder any) (err error) {
	Log.Info("LoadYaml path", "=", path)
	file, err := os.Open(path)
	if err != nil {
		Log.Info("os.Open", "err", err)
		return fmt.Errorf("config / dotprox / LoadYaml() %w", err)
	}
	defer func() {
		errClose := file.Close()
		if errClose != nil {
			err = errClose
			return
		}
	}()

	err = yaml.NewDecoder(file).Decode(holder)
	return err
}

func (yy *YamlFiles) ActivePoolIndex() int {
	Log.Info("Agent.PoolID", "=", yy.Agent.PoolID)
	for id, pool := range yy.Agents.Pools {
		if yy.Agent.PoolID == pool.ID {
			return id
		}
	}
	return -1
}

func (yy *YamlFiles) ActiveInfoIndex() int {
	for id, info := range yy.Pools.DB.Info {
		if yy.Agents.Agent.GroupName == info.Name {
			return id
		}
	}
	return -1
}

func (yy *YamlFiles) SaveAgent2() (err error) {

	agentFile := fmt.Sprintf("%s/%d.yaml", AgentPath, AgentID)
	dir := filepath.Dir(agentFile)

	tmp, err := os.CreateTemp(dir, ".tmp-*.yaml")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	Log.Info("os.CreateTemp", "tmpName", tmpName)

	defer func() {
		errClose := tmp.Close()
		if err != nil {
			err = errClose
			return
		}
		/* for learning
		   /* when os.Rename is called, the file is gone (moved to a new one)
		   /* we cannot not remove it, because it does not exists
		   errRemove := os.Remove(tmpName)
		   if errRemove != nil {
		       err = errRemove
		       return
		   }
		*/
	}()

	// do we need this ?
	// tmp.Sync()

	err = yaml.NewEncoder(tmp).Encode(yy.Agents)
	if err != nil {
		return err
	}

	err = os.Rename(tmpName, agentFile)
	return
}

func (yy *YamlFiles) LoadAgent2() error {
	if yy.agentsLoaded {
		return nil
	}

	agentFile := fmt.Sprintf("%s/%d.yaml", AgentPath, AgentID)
	file, err := os.Open(agentFile)
	if err != nil {
		return err
	}
	defer file.Close()

	err = yaml.NewDecoder(file).Decode(&yy.Agents)
	Log.Info("Agent.PoolID", "=", yy.Agent.PoolID)

	yy.agentsLoaded = true
	return err
}

func (yy *YamlFiles) LoadPool2() error {
	if yy.poolsLoaded {
		return nil
	}
	// pools
	Log.Info("Agent.PoolID", "=", yy.Agent.PoolID)
	for id, pool := range yy.Agents.Pools {
		if yy.Agent.PoolID == pool.ID {
			activePoolIndex = id
		}
	}

	if activePoolIndex == -1 {
		err := fmt.Errorf("no pool id matched: %d\n", yy.Agent.PoolID)
		return err
	}

	poolFile := fmt.Sprintf("%s/%d.yaml", PoolPath, yy.Agent.PoolID)
	Log.Info("poolFile", "=", poolFile)

	file, err := os.Open(poolFile)
	if err != nil {
		return err
	}

	err = yaml.NewDecoder(file).Decode(&yy.Pools)

	yy.poolsLoaded = true
	return err
}

func (yy *YamlFiles) LoadAgent() error {
	if yy.agentsLoaded {
		return nil
	}
	// agents
	agentFile := fmt.Sprintf("%s/%d.yaml", AgentPath, AgentID)
	data, err := os.ReadFile(agentFile)
	if err != nil {
		log.Fatal(err)
	}

	err = yaml.Unmarshal(data, &yy.Agents)
	if err != nil {
		log.Fatal(err)
	}
	Log.Info("Agent.PoolID", "=", yy.Agent.PoolID)

	yy.agentsLoaded = true
	return err
}

func (yy *YamlFiles) LoadPool() error {
	if yy.poolsLoaded {
		return nil
	}
	// pools
	Log.Info("Agent.PoolID", "=", yy.Agent.PoolID)
	for id, pool := range yy.Agents.Pools {
		if yy.Agent.PoolID == pool.ID {
			activePoolIndex = id
		}
	}

	if activePoolIndex == -1 {
		err := fmt.Errorf("no pool id matched: %d\n", yy.Agent.PoolID)
		return err
	}

	poolFile := fmt.Sprintf("%s/%d.yaml", PoolPath, yy.Agent.PoolID)
	Log.Info("poolFile", "=", poolFile)

	data, err := os.ReadFile(poolFile)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(data, &yy.Pools)
	if err != nil {
		return err
	}

	for id, info := range yy.Pools.DB.Info {
		if yy.Agents.Agent.GroupName == info.Name {
			activeInfoIndex = id
		}
	}

	if activeInfoIndex == -1 {
		err := fmt.Errorf("no name matched: %s\n", yy.Agents.Agent.GroupName)
		return err
	}

	yy.poolsLoaded = true
	return err
}
