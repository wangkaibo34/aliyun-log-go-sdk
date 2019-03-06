package consumerLibrary

import (
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"time"
)

type ConsumerHeatBeat struct {
	client       *ConsumerClient
	shutDownFlag bool
	heldShards   []int
	heartShards  []int
	logger       log.Logger
}

func initConsumerHeatBeat(consumerClient *ConsumerClient, logger log.Logger) *ConsumerHeatBeat {
	consumerHeatBeat := &ConsumerHeatBeat{
		client:       consumerClient,
		shutDownFlag: false,
		heldShards:   []int{},
		heartShards:  []int{},
		logger:       logger,
	}
	return consumerHeatBeat
}

func (consumerHeatBeat *ConsumerHeatBeat) getHeldShards() []int {
	m.RLock()
	defer m.RUnlock()
	return consumerHeatBeat.heldShards
}

func (consumerHeatBeat *ConsumerHeatBeat) setHeldShards(heldShards []int) {
	m.Lock()
	defer m.Unlock()
	consumerHeatBeat.heldShards = heldShards
}

func (consumerHeatBeat *ConsumerHeatBeat) setHeartShards(heartShards []int) {
	m.Lock()
	defer m.Unlock()
	consumerHeatBeat.heartShards = heartShards
}

func (consumerHeatBeat *ConsumerHeatBeat) shutDownHeart() {
	level.Info(consumerHeatBeat.logger).Log("msg", "try to stop heart beat")
	consumerHeatBeat.shutDownFlag = true
}

func (consumerHeatBeat *ConsumerHeatBeat) heartBeatRun() {
	var lastHeartBeatTime int64

	for !consumerHeatBeat.shutDownFlag {
		lastHeartBeatTime = time.Now().Unix()
		responseShards, err := consumerHeatBeat.client.heartBeat(consumerHeatBeat.heartShards)
		if err != nil {
			level.Warn(consumerHeatBeat.logger).Log("msg", "send heartbeat error", "error", err)
		} else {
			level.Info(consumerHeatBeat.logger).Log("heart beat result", fmt.Sprintf("%v", consumerHeatBeat.heartShards), "get", fmt.Sprintf("%v", responseShards))
			consumerHeatBeat.setHeldShards(responseShards)
			if !IntSliceReflectEqual(consumerHeatBeat.heartShards, consumerHeatBeat.heldShards) {
				currentSet := Set(consumerHeatBeat.heartShards)
				responseSet := Set(consumerHeatBeat.heldShards)
				add := Subtract(currentSet, responseSet)
				remove := Subtract(responseSet, currentSet)
				level.Info(consumerHeatBeat.logger).Log("shard reorganize, adding:", fmt.Sprintf("%v", add), "removing:", fmt.Sprintf("%v", remove))
			}

			consumerHeatBeat.setHeartShards(consumerHeatBeat.getHeldShards()[:])
		}
		TimeToSleep(int64(consumerHeatBeat.client.option.HeartbeatIntervalInSecond), lastHeartBeatTime, consumerHeatBeat.shutDownFlag)
	}
	level.Info(consumerHeatBeat.logger).Log("msg", "heart beat exit")
}
