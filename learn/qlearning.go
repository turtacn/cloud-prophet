package learn

import (
	//	"bufio"
	//	"bytes"
	"encoding/json"
	"fmt"
	//"io"
	"io/ioutil"
	"math"
	"math/rand"
	"os"
	"time"
)

type State struct {
	CPUH     bool `json:"cpu"`
	MemH     bool `json:"mem"`
	RpsH     int  `json:"rps"`
	RtimeH   bool `json:"rtime"`
	R5xxH    bool `json:"r5xx"`
	Replicas int  `json:"replicas"`
}

type Action struct {
	Plus  int64 `json:"plus"`
	Stay  int64 `json:"stay"`
	Minus int64 `json:"minus"`
}

type QLearn struct {
	Gamma   float64 `json:"gamma"` //0.4
	Epsilon float64 `json:"epsilon"`
	rnd     *rand.Rand
	States  map[string]Action `json:"states"`
	//	Action int -1,0,+1
	//StateList
	CurrentState State
}

// If prob < eps then Explore
// Else then choose maxRewardAction
func (q QLearn) ChooseAction() int {
	var nextAction int
	prob := q.rnd.Float64()
	fmt.Println("Prob ", prob, "? Epsilon", q.Epsilon)
	if prob < q.Epsilon {
		fmt.Println("Go Explore")
		nextAction = q.rnd.Intn(3) - 1 // -1,0,1
	} else {
		fmt.Println("Go Best")
		action := q.States[toKey(q.CurrentState)]
		maxR := int64(math.Max(float64(action.Plus), math.Max(float64(action.Stay), float64(action.Minus))))
		if action.Stay == maxR {
			nextAction = 0
		} else if action.Plus == maxR {
			nextAction = +1
		} else { // minus
			nextAction = -1
		}
	}
	fmt.Println("Action :", nextAction)
	return nextAction
}

/*
func (q QLearn) ValidAction(action int) bool {
	// TODO: May be 0 is good when no one using it
	if q.CurrentState.Replicas+action <= 0 {
		return false
	}
	return true
}
// TODO: This fn may not need
func (q *QLearn) GoNextState(action int) {
	q.CurrentState.Replicas += action
	// TODO: What about cpu,mem,etc..
}
*/
func (q QLearn) MaximumOp(state State) int64 {
	// Find Best action then return it Q-Matrix
	action := q.States[toKey(state)]
	max := math.Max(float64(action.Plus), math.Max(float64(action.Stay), float64(action.Minus)))
	fmt.Println("Max ", action)
	fmt.Println("Max=", max)
	return int64(max)
}

func (q *QLearn) Reward(state State, action int, nowStatus map[string]int64) int64 {

	var reward int64 = 0
	/*
		var stateReward int64 = 0
		var stateNegReward int64 = 0
		if state.CPUH {
			stateReward += 5
		} else {
			stateNegReward += 5
		}

		if state.MemH {
			stateReward += 5
		} else {
			stateNegReward += 5
		}

		if state.RtimeH {
			stateReward += 5
		} else {
			stateNegReward += 5
		}

		// RpsH
		stateReward += 5 * (int64(state.RpsH))
		//stateNegReward += 5 * (int64(state.RpsH))

		// TODO : use in reward? R5xx

		// Replicas
		//stateReward += 5 * (int64(state.Replicas))
		stateNegReward += 5 * (int64(state.Replicas))

		// ACTION
		if action == 1 {
			reward -= 5
			reward += stateReward
		} else if action == -1 {
			reward += 5
			reward -= stateReward
			reward += stateNegReward
			// Replicas 1-1=0
			if nowStatus["replicas"] == 1 {
				reward -= 30
			}
		} else {
			//	reward += stateReward
			reward -= stateNegReward
		}
		// Replicas 1-1=0
		if action == -1 && nowStatus["replicas"] == 1 {
			reward -= 20
		} else {
			// Replicas - More replicas more penalty
			reward += 3 - (nowStatus["replicas"]-1)*int64(3)

			// RTime
			reward += (4 - nowStatus["rtime"]) * 10

			// 5XX
			reward -= nowStatus["r5xx"]
			// CPU -- ( replicas -1 before * bcuz 1 replicas is good)
			if nowStatus["cpu"] < 75 {
				reward += nowStatus["cpu"]/3 - (nowStatus["replicas"]-1)*5
			} else {
				reward -= nowStatus["cpu"] - 75
			}
			// Low Mem --
			if nowStatus["cpu"] < 75 {
				reward += nowStatus["memory"]/3 - (nowStatus["replicas"]-1)*5
			} else {
				reward -= nowStatus["memory"] - 75
			}
		}
	*/
	// Goal/High state reward
	fmt.Print(" rew00", reward)
	if q.CurrentState.RpsH > 0 && q.CurrentState.Replicas > 1 {
		goalReward := 0
		if !q.CurrentState.CPUH {
			goalReward += 30
		}

		if !q.CurrentState.MemH {
			goalReward += 30
		}
		reward += int64(goalReward)
		fmt.Print(" +goal", goalReward)
	}
	fmt.Print(" rew", reward)
	// Low state reward
	if q.CurrentState.CPUH {
		cpuCost := nowStatus["cpu"] / 10
		cpuCost = cpuCost * 3
		fmt.Print(" -cpu", cpuCost)
		reward -= int64(cpuCost)
	}

	fmt.Print(" rew", reward)
	if q.CurrentState.MemH {
		memCost := nowStatus["mem"] / 10
		memCost = memCost * 3
		fmt.Print(" -mem", memCost)
		reward -= int64(memCost)
	}

	fmt.Print(" rew", reward)
	rpsReward := int64(q.CurrentState.RpsH * 15)
	fmt.Print(" +rps", rpsReward)
	reward += rpsReward

	fmt.Print(" rew", reward)
	// Rtime reward
	if q.CurrentState.RtimeH {
		reward -= 5
	} else {
		reward += 5
	}

	// replicasCost
	replicasCost := int64((q.CurrentState.Replicas - 1) * 15)
	fmt.Print(" -replicas", replicasCost)
	reward -= replicasCost

	fmt.Print(" rew", reward)
	if action == 1 {
		reward -= 2
	} else if action == 0 {
		reward += 2
	} else if action == -1 {
		reward += 3
	}

	fmt.Print(" rew", reward)
	if action == -1 && state.Replicas == 1 {
		reward -= 100
	}

	// R(c,a) = R(current,action)+gamma*MaximumOp(NextState)
	reward += int64(q.Gamma * float64(q.MaximumOp(state)))
	fmt.Println("Reward ", reward)
	// Update Q-Matrix
	qAction := q.States[toKey(state)]
	if action == 1 {
		qAction.Plus = reward
		q.States[toKey(state)] = qAction
	} else if action == -1 {
		qAction.Minus = reward
		q.States[toKey(state)] = qAction
	} else {
		qAction.Stay = reward
		q.States[toKey(state)] = qAction
	}
	fmt.Println("State", state, " Plus:", qAction.Plus, " Stay:", qAction.Stay, " Minus:", qAction.Minus)
	return reward
}

func (q *QLearn) Init() {
	seed := rand.NewSource(time.Now().UnixNano())
	q.rnd = rand.New(seed)

	//	q.CurrentState.Replicas = 1
	q.States = make(map[string]Action)
	q.CurrentState = State{Replicas: 1}
}

// Change current state according to real cpu,mem,etc....
func (q *QLearn) SetCurrentState(cpu, mem, rps, rtime, r5xx int64, replicas int) {
	// TODO: Create Fn to set all this value
	if int(cpu) > 70 {
		q.CurrentState.CPUH = true
	} else {
		q.CurrentState.CPUH = false
	}

	if int(mem) > 70 {
		q.CurrentState.MemH = true
	} else {
		q.CurrentState.MemH = false
	}

	// TODO: Create more state to represent hom much work load is
	q.CurrentState.RpsH = int(rps) / 100

	if rtime > 5 {
		q.CurrentState.RtimeH = true
	} else {
		q.CurrentState.RtimeH = false
	}

	if r5xx > 10 {
		q.CurrentState.R5xxH = true
	} else {
		q.CurrentState.R5xxH = false
	}
	q.CurrentState.Replicas = replicas

	// Create state if it doesnt exist
	if _, have := q.States[toKey(q.CurrentState)]; have == false {
		action := Action{}
		//if q.CurrentState.RtimeH {
		q.States[toKey(q.CurrentState)] = action
	}
}

func (q QLearn) Save(path string) {
	jsonData, _ := json.Marshal(q)
	file, err := os.Create(path)
	if err != nil {
		//		return err
	}
	defer file.Close()
	file.WriteString(string(jsonData))
	file.Sync()
}

func (q *QLearn) Load(path string) error {

	seed := rand.NewSource(time.Now().UnixNano())
	q.rnd = rand.New(seed)

	dat, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	//fmt.Print(string(dat))
	err = json.Unmarshal(dat, q)
	if err != nil {
		return err
	}
	fmt.Println("Load success")
	return nil

}

func toKey(c State) string {
	return fmt.Sprint(c.CPUH, c.MemH, c.RpsH, c.RtimeH, c.R5xxH, c.Replicas)
}

/*
func main() {
	agent := QLearn{Gamma: 0.4, Epsilon: 0.6}
	agent.Init()
	agent.States[State{Replicas: 2}] = Action{Plus: 1, Stay: 2, Minus: -1}
	fmt.Println(agent)
	agent.CurrentState = State{Replicas: 2}
	fmt.Println("MAX", agent.MaximumOp())
}
*/