package component

import (
	"github.com/Azer0s/quacktors"
	"github.com/Azer0s/quacktors/register"
	"sync"
)

//TODO: logging

type Strategy int

const (
	ONE_FOR_ONE_STRATEGY Strategy = iota
	ALL_FOR_ONE_STRATEGY
	FAIL_ALL_STRATEGY
)

func Supervisor(strategy Strategy, actors map[string]quacktors.Actor) *SupervisorComponent {
	return &SupervisorComponent{
		strategy: strategy,
		actors:   actors,
		pids:     make(map[string]*quacktors.Pid),
		monitors: make(map[string]quacktors.Abortable),
	}
}

type SupervisorComponent struct {
	strategy Strategy
	actors   map[string]quacktors.Actor
	pids     map[string]*quacktors.Pid
	monitors map[string]quacktors.Abortable
}

func (s *SupervisorComponent) setupActor(ctx *quacktors.Context, id string, actor quacktors.Actor) {
	register.ModifyUnsafe(func(register *map[string]*quacktors.Pid, mu *sync.RWMutex) {
		p := quacktors.SpawnStateful(actor)
		s.pids[id] = p

		s.monitors[id] = ctx.Monitor(p)

		(*register)[id] = p
	})
}

func (s *SupervisorComponent) Init(ctx *quacktors.Context) {
	register.ModifyUnsafe(func(register *map[string]*quacktors.Pid, mu *sync.RWMutex) {
		mu.Lock()
		defer mu.Unlock()

		for id, actor := range s.actors {
			s.setupActor(ctx, id, actor)
		}
	})

	ctx.Defer(func() {
		for id := range s.pids {
			//first abort all monitors
			s.monitors[id].Abort()

			//then kill all other actors
			ctx.Kill(s.pids[id])
		}
	})
}

func (s *SupervisorComponent) Run(ctx *quacktors.Context, message quacktors.Message) {
	register.ModifyUnsafe(func(register *map[string]*quacktors.Pid, mu *sync.RWMutex) {
		//lock here so that the chance of someone getting a dead pid is minimized

		mu.Lock()
		defer mu.Unlock()

		if m, ok := message.(quacktors.DownMessage); ok {
			switch s.strategy {
			case ONE_FOR_ONE_STRATEGY:
				//just restart the actor that failed
				for id, pid := range s.pids {
					if pid.Is(m.Who) {
						s.setupActor(ctx, id, s.actors[id])
					}
				}

			case ALL_FOR_ONE_STRATEGY:
				for id, actor := range s.actors {
					pid := s.pids[id]

					if !pid.Is(m.Who) {
						//first abort all other monitors
						s.monitors[id].Abort()

						//then kill all other actors
						ctx.Kill(s.pids[id])
					}

					//then respawn every actor
					s.setupActor(ctx, id, actor)
				}

			case FAIL_ALL_STRATEGY:
				for id, pid := range s.pids {
					if !pid.Is(m.Who) {
						//first abort all other monitors
						s.monitors[id].Abort()

						//then kill all other actors
						ctx.Kill(s.pids[id])
					}
				}

				//set the pids map to empty so that it's not cleared in defer
				s.pids = make(map[string]*quacktors.Pid)

				//kill supervisor
				ctx.Send(ctx.Self(), quacktors.PoisonPill{})
			}
		}

		if _, ok := message.(quacktors.KillMessage); ok {
			//"gracefully" shut down supervisor

			ctx.Logger.Info("gracefully shutting down supervisor")

			//then kill supervisor
			ctx.Send(ctx.Self(), quacktors.PoisonPill{})
		}
	})
}
