package actor

import (
	"errors"
	"time"

	"github.com/Ontology/common/log"
	"github.com/emirpasic/gods/stacks/linkedliststack"
	"fmt"
)

type localCcntmext struct {
	message            interface{}
	parent             *PID
	self               *PID
	actor              Actor
	supervisor         SupervisorStrategy
	producer           Producer
	inboundMiddleware  ActorFunc
	outboundMiddleware SenderFunc
	behavior           behaviorStack
	receive            ActorFunc
	children           PIDSet
	watchers           PIDSet
	watching           PIDSet
	stash              *linkedliststack.Stack
	stopping           bool
	restarting         bool
	receiveTimeout     time.Duration
	t                  *time.Timer
	restartStats       *RestartStatistics
}

func newLocalCcntmext(producer Producer, supervisor SupervisorStrategy, inboundMiddleware []InboundMiddleware, outboundMiddleware []OutboundMiddleware, parent *PID) *localCcntmext {
	this := &localCcntmext{
		parent:     parent,
		producer:   producer,
		supervisor: supervisor,
	}

	// Construct the inbound middleware chain with the final receiver at the end
	if inboundMiddleware != nil {
		this.inboundMiddleware = makeInboundMiddlewareChain(inboundMiddleware, func(ctx Ccntmext) {
			if _, ok := this.message.(*PoisonPill); ok {
				this.self.Stop()
			} else {
				this.receive(ctx)
			}
		})
	}

	// Construct the outbound middleware chain with the final sender at the end
	this.outboundMiddleware = makeOutboundMiddlewareChain(outboundMiddleware, func(_ Ccntmext, target *PID, envelope *MessageEnvelope) {
		target.ref().SendUserMessage(target, envelope)
	})

	this.incarnateActor()
	return this
}

func (ctx *localCcntmext) Actor() Actor {
	return ctx.actor
}

func (ctx *localCcntmext) Message() interface{} {
	envelope, ok := ctx.message.(*MessageEnvelope)
	if ok {
		return envelope.Message
	}
	return ctx.message
}

func (ctx *localCcntmext) Sender() *PID {
	envelope, ok := ctx.message.(*MessageEnvelope)
	if ok {
		return envelope.Sender
	}
	return nil
}

func (ctx *localCcntmext) MessageHeader() ReadonlyMessageHeader {
	envelope, ok := ctx.message.(*MessageEnvelope)
	if ok {
		if envelope.Header != nil {
			return envelope.Header
		}
	}
	return emptyMessageHeader
}

func (ctx *localCcntmext) Tell(pid *PID, message interface{}) {
	ctx.sendUserMessage(pid, message)
}

func (ctx *localCcntmext) sendUserMessage(pid *PID, message interface{}) {
	if ctx.outboundMiddleware != nil {
		if env, ok := message.(*MessageEnvelope); ok {
			ctx.outboundMiddleware(ctx, pid, env)
		} else {
			ctx.outboundMiddleware(ctx, pid, &MessageEnvelope{
				Header:  nil,
				Message: message,
				Sender:  nil,
			})
		}
	} else {
		pid.ref().SendUserMessage(pid, message)
	}
}

func (ctx *localCcntmext) Request(pid *PID, message interface{}) {
	env := &MessageEnvelope{
		Header:  nil,
		Message: message,
		Sender:  ctx.Self(),
	}

	ctx.sendUserMessage(pid, env)
}

func (ctx *localCcntmext) RequestFuture(pid *PID, message interface{}, timeout time.Duration) *Future {
	future := NewFuture(timeout)
	env := &MessageEnvelope{
		Header:  nil,
		Message: message,
		Sender:  future.PID(),
	}
	ctx.sendUserMessage(pid, env)

	return future
}

func (ctx *localCcntmext) Stash() {
	if ctx.stash == nil {
		ctx.stash = linkedliststack.New()
	}

	ctx.stash.Push(ctx.message)
}

func (ctx *localCcntmext) cancelTimer() {
	if ctx.t != nil {
		ctx.t.Stop()
		ctx.t = nil
		ctx.receiveTimeout = 0
	}
}

func (ctx *localCcntmext) receiveTimeoutHandler() {
	if ctx.t != nil {
		ctx.cancelTimer()
		ctx.self.Tell(receiveTimeoutMessage)
	}
}

func (ctx *localCcntmext) SetReceiveTimeout(d time.Duration) {
	if d == ctx.receiveTimeout {
		return
	}
	if ctx.t != nil {
		ctx.t.Stop()
	}

	if d < time.Millisecond {
		// anything less than than 1 millisecond is set to zero
		d = 0
	}

	ctx.receiveTimeout = d
	if d > 0 {
		if ctx.t == nil {
			ctx.t = time.AfterFunc(d, ctx.receiveTimeoutHandler)
		} else {
			ctx.t.Reset(d)
		}
	}
}

func (ctx *localCcntmext) ReceiveTimeout() time.Duration {
	return ctx.receiveTimeout
}

func (ctx *localCcntmext) Children() []*PID {
	r := make([]*PID, ctx.children.Len())
	ctx.children.ForEach(func(i int, p PID) {
		r[i] = &p
	})
	return r
}

func (ctx *localCcntmext) Self() *PID {
	return ctx.self
}

func (ctx *localCcntmext) Parent() *PID {
	return ctx.parent
}

func (ctx *localCcntmext) Receive(message interface{}) {
	ctx.processMessage(message)
}

func (ctx *localCcntmext) RestartStats() *RestartStatistics {
	//lazy initialize the child restart stats if this is the first time
	//further mutations are handled within "restart"
	if ctx.restartStats == nil {
		ctx.restartStats = NewRestartStatistics()
	}
	return ctx.restartStats
}

func (ctx *localCcntmext) EscalateFailure(reason interface{}, message interface{}) {
	failure := &Failure{Reason: reason, Who: ctx.self, RestartStats: ctx.RestartStats()}
	ctx.self.sendSystemMessage(suspendMailboxMessage)
	if ctx.parent == nil {
		handleRootFailure(failure)
	} else {
		//TODO: Akka recursively suspends all children also on failure
		//Not sure if I think this is the right way to go, why do children need to wait for their parents failed state to recover?
		ctx.parent.sendSystemMessage(failure)
	}
}

func (ctx *localCcntmext) InvokeUserMessage(md interface{}) {
	influenceTimeout := true
	if ctx.receiveTimeout > 0 {
		_, influenceTimeout = md.(NotInfluenceReceiveTimeout)
		influenceTimeout = !influenceTimeout
		if influenceTimeout {
			ctx.t.Stop()
		}
	}

	ctx.processMessage(md)

	if ctx.receiveTimeout > 0 && influenceTimeout {
		ctx.t.Reset(ctx.receiveTimeout)
	}
}

func (ctx *localCcntmext) processMessage(m interface{}) {
	ctx.message = m

	if ctx.inboundMiddleware != nil {
		ctx.inboundMiddleware(ctx)
	} else {
		if _, ok := m.(*PoisonPill); ok {
			ctx.self.Stop()
		} else {
			ctx.receive(ctx)
		}
	}

	ctx.message = nil
}

func (ctx *localCcntmext) incarnateActor() {
	pid := ctx.producer()
	ctx.restarting = false
	ctx.stopping = false
	ctx.actor = pid
	ctx.receive = pid.Receive
}

func (ctx *localCcntmext) InvokeSystemMessage(message interface{}) {
	switch msg := message.(type) {
	case *ccntminuation:
		ctx.message = msg.message // apply the message that was present when we started the await
		msg.f()                   // invoke the ccntminuation in the current actor ccntmext
		ctx.message = nil         // release the message
	case *Started:
		ctx.InvokeUserMessage(msg) // forward
	case *Watch:
		if ctx.stopping {
			msg.Watcher.sendSystemMessage(&Terminated{Who: ctx.self})
		} else {
			ctx.watchers.Add(msg.Watcher)
		}
	case *Unwatch:
		ctx.watchers.Remove(msg.Watcher)
	case *Stop:
		ctx.handleStop(msg)
	case *Terminated:
		ctx.handleTerminated(msg)
	case *Failure:
		ctx.handleFailure(msg)
	case *Restart:
		ctx.handleRestart(msg)
	default:
		log.Error("unknown system message", fmt.Sprintf("%v",msg))
	}
}

func (ctx *localCcntmext) handleRestart(msg *Restart) {
	ctx.stopping = false
	ctx.restarting = true
	ctx.InvokeUserMessage(restartingMessage)
	ctx.children.ForEach(func(_ int, pid PID) {
		pid.Stop()
	})
	ctx.tryRestartOrTerminate()
}

//I am stopping
func (ctx *localCcntmext) handleStop(msg *Stop) {
	ctx.stopping = true
	ctx.restarting = false

	ctx.InvokeUserMessage(stoppingMessage)
	ctx.children.ForEach(func(_ int, pid PID) {
		pid.Stop()
	})
	ctx.tryRestartOrTerminate()
}

//child stopped, check if we can stop or restart (if needed)
func (ctx *localCcntmext) handleTerminated(msg *Terminated) {
	ctx.children.Remove(msg.Who)
	ctx.watching.Remove(msg.Who)

	ctx.InvokeUserMessage(msg)
	ctx.tryRestartOrTerminate()
}

//offload the supervision completely to the supervisor strategy
func (ctx *localCcntmext) handleFailure(msg *Failure) {
	if strategy, ok := ctx.actor.(SupervisorStrategy); ok {
		strategy.HandleFailure(ctx, msg.Who, msg.RestartStats, msg.Reason, msg.Message)
		return
	}
	ctx.supervisor.HandleFailure(ctx, msg.Who, msg.RestartStats, msg.Reason, msg.Message)
}

func (ctx *localCcntmext) tryRestartOrTerminate() {
	if ctx.t != nil {
		ctx.t.Stop()
		ctx.t = nil
		ctx.receiveTimeout = 0
	}

	if !ctx.children.Empty() {
		return
	}

	if ctx.restarting {
		ctx.restart()
		return
	}

	if ctx.stopping {
		ctx.stopped()
	}
}

func (ctx *localCcntmext) restart() {
	ctx.incarnateActor()
	ctx.InvokeUserMessage(startedMessage)
	if ctx.stash != nil {
		for !ctx.stash.Empty() {
			msg, _ := ctx.stash.Pop()
			ctx.InvokeUserMessage(msg)
		}
	}
	ctx.self.sendSystemMessage(resumeMailboxMessage)
}

func (ctx *localCcntmext) stopped() {
	ProcessRegistry.Remove(ctx.self)
	ctx.InvokeUserMessage(stoppedMessage)
	otherStopped := &Terminated{Who: ctx.self}
	ctx.watchers.ForEach(func(i int, pid PID) {
		pid.sendSystemMessage(otherStopped)
	})
}

func (ctx *localCcntmext) SetBehavior(behavior ActorFunc) {
	ctx.behavior.Clear()
	ctx.receive = behavior
}

func (ctx *localCcntmext) PushBehavior(behavior ActorFunc) {
	ctx.behavior.Push(ctx.receive)
	ctx.receive = behavior
}

func (ctx *localCcntmext) PopBehavior() {
	if ctx.behavior.Len() == 0 {
		panic("Cannot unbecome actor base behavior")
	}
	ctx.receive, _ = ctx.behavior.Pop()
}

func (ctx *localCcntmext) Watch(who *PID) {
	who.sendSystemMessage(&Watch{
		Watcher: ctx.self,
	})
	ctx.watching.Add(who)
}

func (ctx *localCcntmext) Unwatch(who *PID) {
	who.sendSystemMessage(&Unwatch{
		Watcher: ctx.self,
	})
	ctx.watching.Remove(who)
}
func (ctx *localCcntmext) Respond(response interface{}) {
	// If the message is addressed to nil forward it to the dead letter channel
	if ctx.Sender() == nil {
		deadLetter.SendUserMessage(nil, response)
		return
	}

	ctx.Tell(ctx.Sender(), response)
}

func (ctx *localCcntmext) Spawn(props *Props) *PID {
	pid, _ := ctx.SpawnNamed(props, ProcessRegistry.NextId())
	return pid
}

func (ctx *localCcntmext) SpawnPrefix(props *Props, prefix string) *PID {
	pid, _ := ctx.SpawnNamed(props, prefix+ProcessRegistry.NextId())
	return pid
}

func (ctx *localCcntmext) SpawnNamed(props *Props, name string) (*PID, error) {
	if props.guardianStrategy != nil {
		panic(errors.New("Props used to spawn child cannot have GuardianStrategy"))
	}

	pid, err := props.spawn(ctx.self.Id+"/"+name, ctx.self)
	if err != nil {
		return pid, err
	}

	ctx.children.Add(pid)
	ctx.Watch(pid)

	return pid, nil
}

func (ctx *localCcntmext) GoString() string {
	return ctx.self.String()
}

func (ctx *localCcntmext) String() string {
	return ctx.self.String()
}

func (ctx *localCcntmext) AwaitFuture(f *Future, ccntm func(res interface{}, err error)) {
	wrapper := func() {
		ccntm(f.result, f.err)
	}

	message := ctx.message
	//invoke the callback when the future completes
	f.ccntminueWith(func(res interface{}, err error) {
		//send the wrapped callaback as a ccntminuation message to self
		ctx.self.sendSystemMessage(&ccntminuation{
			f:       wrapper,
			message: message,
		})
	})
}

func (*localCcntmext) RestartChildren(pids ...*PID) {
	for _, pid := range pids {
		pid.sendSystemMessage(restartMessage)
	}
}

func (*localCcntmext) StopChildren(pids ...*PID) {
	for _, pid := range pids {
		pid.sendSystemMessage(stopMessage)
	}
}

func (*localCcntmext) ResumeChildren(pids ...*PID) {
	for _, pid := range pids {
		pid.sendSystemMessage(resumeMailboxMessage)
	}
}
