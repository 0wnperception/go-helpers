package commandManager

import (
	"context"
	"encoding/json"
	"go-helpers/mqttConnector"
	"sync"

	"github.com/google/uuid"
)

type MqttAPIConfig struct {
	MqttAPICommandRequestTopic  string
	MqttAPICommandResponseTopic string
}

type MqttAPIResponse struct {
	RequestId string `json:"request_id"`
	Status    int    `json:"status"`
}

type MqttAPIRequest struct {
	RequestId string      `json:"request_id"`
	DeviceId  int         `json:"device_id"`
	RuleId    int         `json:"rule_id"`
	Args      interface{} `json:"args"`
}

type CommandStatus int

const (
	commandStatusInitial CommandStatus = 0
	commandStatusBusy    CommandStatus = 1
	commandStatusRunning CommandStatus = 2
	commandStatusDone    CommandStatus = 3
	commandStatusError   CommandStatus = 4
)

type mqttCommand struct {
	status CommandStatus
	done   chan struct{}
}

type MqttCommandManager struct {
	sync.Locker
	cfg       *MqttAPIConfig
	commands  map[string]*mqttCommand
	connector *mqttConnector.MqttConnector
}

func NewMqttCommandManager(cfg *MqttAPIConfig, connector *mqttConnector.MqttConnector) *MqttCommandManager {
	man := &MqttCommandManager{
		Locker:    &sync.Mutex{},
		commands:  make(map[string]*mqttCommand),
		connector: connector,
		cfg:       cfg,
	}
	connector.Subscripe(cfg.MqttAPICommandResponseTopic, man.processCommand)
	return man
}

func (man *MqttCommandManager) RunCommand(ctx context.Context, deviceId int, ruleId int, args interface{}) CommandStatus {
	uniqID := uuid.New().String()
	man.addCommand(uniqID)
	binCom, _ := json.Marshal(MqttAPIRequest{
		RequestId: uniqID,
		DeviceId:  deviceId,
		RuleId:    ruleId,
		Args:      args,
	})
	man.connector.Publish(man.cfg.MqttAPICommandRequestTopic, binCom)
	status := man.waitCommand(ctx, uniqID)
	man.removeCommand(uniqID)
	return status
}
func (man *MqttCommandManager) processCommand(bindata []byte) {
	response := &MqttAPIResponse{}
	if err := json.Unmarshal(bindata, response); err == nil {
		man.setCommandStatus(response.RequestId, response.Status)
	}
}

func (man *MqttCommandManager) addCommand(uniqId string) {
	man.Lock()
	man.commands[uniqId] = &mqttCommand{
		status: commandStatusInitial,
		done:   make(chan struct{}),
	}
	man.Unlock()
}

func (man *MqttCommandManager) removeCommand(uniqId string) {
	man.Lock()
	delete(man.commands, uniqId)
	man.Unlock()
}

func (man *MqttCommandManager) waitCommand(ctx context.Context, uniqId string) CommandStatus {
	com := man.commands[uniqId]
	select {
	case <-com.done:
		break
	case <-ctx.Done():
		break
	}
	return com.status
}

func (man *MqttCommandManager) setCommandStatus(uniqId string, status int) {
	if c, ok := man.commands[uniqId]; ok {
		c.status = CommandStatus(status)
		if c.status == commandStatusDone ||
			c.status == commandStatusError ||
			c.status == commandStatusBusy {
			close(c.done)
		}
	}
}
