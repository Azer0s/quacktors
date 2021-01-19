package component

import (
	"github.com/Azer0s/quacktors"
	"math"
)

type pidWithUsage struct {
	pid   *quacktors.Pid
	usage uint64
}

type loadBalancerComponent struct {
	threshold     uint16
	actor         quacktors.Actor
	pids          []*pidWithUsage
	usageFunction func() uint16
}

func (l *loadBalancerComponent) Init(ctx *quacktors.Context) {
	l.spawnOrDestroy(ctx)

	//if the load balancer is killed forcefully, kill all the
	//spawned actors
	ctx.Defer(func() {
		for _, p := range l.pids {
			ctx.Kill(p.pid)
		}
	})
}

func (l *loadBalancerComponent) spawnOrDestroy(ctx *quacktors.Context) {
	usage := l.usageFunction()
	currentPids := len(l.pids)
	requiredPids := int(math.Max(1, math.Round((float64(usage)-(float64(l.threshold)/2))/float64(l.threshold))+1))

	if currentPids < requiredPids {
		delta := requiredPids - currentPids
		//spawn delta amount of pids

		for i := 0; i < delta; i++ {
			p := &pidWithUsage{
				pid:   quacktors.SpawnStateful(l.actor),
				usage: 0,
			}

			ctx.Monitor(p.pid)

			l.pids = append(l.pids, p)
		}
	} else if currentPids > requiredPids && requiredPids != 0 {
		delta := currentPids - requiredPids
		//kill delta amount of pids

		for i := 0; i < delta; i++ {
			p := l.pids[len(l.pids)-1]
			ctx.Send(p.pid, quacktors.PoisonPill{})
			l.pids = l.pids[:len(l.pids)-1]
		}
	} else {
		//currentPids < requiredPids is the optimum
	}
}

func (l *loadBalancerComponent) Run(ctx *quacktors.Context, message quacktors.Message) {
	if d, ok := message.(quacktors.DownMessage); ok {
		for i, p := range l.pids {
			if d.Who.Is(p.pid) {
				copy(l.pids[i:], l.pids[i+1:])
				l.pids = l.pids[:len(l.pids)-1]

				l.spawnOrDestroy(ctx)

				return
			}
		}
	}

	l.spawnOrDestroy(ctx)
	var sendTo = &pidWithUsage{
		pid:   nil,
		usage: math.MaxUint64,
	}

	for _, pid := range l.pids {
		if pid.usage < sendTo.usage {
			sendTo = pid
		}
	}

	ctx.Send(sendTo.pid, message)

	sendTo.usage++
}

//LoadBalancer creates a load balancer component from
//the provided parameters. The load balancer scales an
//actor according to usage determined by the usage function
//and the scaling threshold. If one actor in the pool goes
//down, it is automatically restarted if needed. At least
//one instance of the actor has to be always running. The
//scaling of the actor is calculated by a threshold function.
//If the load balancer is killed, it takes down all of the
//actors in its pool so to avoid actor leaks.
//
//Threshold function
//
//The threshold function is defined like so:
//
//f(u,t) = max(1, (\lfloor {\frac{u - (t/2)}_{t}} \rfloor) + 1)
//
//Where `u` is the usage and `t` is the threshold and the
//function domain is u >= 0 and t >= 1.
func LoadBalancer(threshold uint16, actor quacktors.Actor, usageFunction func() uint16) quacktors.Actor {
	return &loadBalancerComponent{
		threshold:     threshold,
		actor:         actor,
		usageFunction: usageFunction,
		pids:          make([]*pidWithUsage, 0),
	}
}
