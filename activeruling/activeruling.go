package activeruling

import (
	"fmt"
	"sync"
	"time"

	lnksworks "github.com/efjoubert/lnkworks"
)

func RegisterSchedule(schdlname string, monduration time.Duration, actions ...func(string, time.Time)) (schdl *Schedule) {
	schdlslck.RLock()
	defer schdlslck.RUnlock()

	if _, schdlok := schedules[schdlname]; schdlok {
		schdl = schedules[schdlname]
	} else {
		schdl = &Schedule{actions: actions[:], monduration: monduration, schdlname: schdlname, cmdDone: make(chan bool, 1), nextcmds: make(chan scheduleCommand)}
		schedules[schdlname] = schdl
	}

	return schdl
}

func FindSchedule(schdlname string) (schdl *Schedule) {
	schdlslck.RLock()
	defer schdlslck.RUnlock()
	if _, schdlok := schedules[schdlname]; schdlok {
		schdl = schedules[schdlname]
	}
	return schdl
}

type Schedule struct {
	actions     []func(string, time.Time)
	monduration time.Duration
	schdlname   string
	nextcmds    chan scheduleCommand
	lastcmd     scheduleCommand
	cmdDone     chan bool
	tkr         *time.Ticker
}

type scheduleCommand int

const (
	noAction        scheduleCommand = 0
	enableSchedule  scheduleCommand = 1
	disableSchedule scheduleCommand = 2
)

func (schdl *Schedule) EnableSchedule() {

	go func() {
		scheduledQueue <- schdl
		schdl.nextcmds <- enableSchedule
	}()

	<-schdl.cmdDone

	fmt.Println(schdl.schdlname, ":", "enabled")

}

func (schdl *Schedule) DisableSchedule() {
	go func() {
		schdl.nextcmds <- disableSchedule
	}()
	<-schdl.cmdDone
	fmt.Println(schdl.schdlname, ":", "disabled")
}

var scheduledQueue chan *Schedule

var schedules map[string]*Schedule
var schdlslck *sync.RWMutex

func processQueuedChannels() {

	for {
		select {
		case schdl := <-scheduledQueue:
			go func() {
				executeScheduleCommands(schdl)
			}()
		}
	}
}

func executeScheduleCommands(schdl *Schedule) {
	for {
		select {
		case schdlcmd := <-schdl.nextcmds:
			schdl.lastcmd = schdlcmd
			break
		}
		if schdl.lastcmd == enableSchedule {
			if schdl.tkr != nil {
				schdl.tkr.Stop()
				schdl.tkr = nil
			}

			schdl.tkr = time.NewTicker(schdl.monduration)

			go func() {
				tkr := schdl.tkr
				schdl.cmdDone <- true
				alreadyTicking := false
				for {
					select {
					case c := <-tkr.C:
						if !alreadyTicking {
							alreadyTicking = true
							go func() {
								executeScheduleActions(schdl, c)
								alreadyTicking = false
							}()
						}
					}
				}
			}()
			schdl.lastcmd = noAction
		} else if schdl.lastcmd == disableSchedule {
			if schdl.tkr != nil {
				schdl.tkr.Stop()
				schdl.tkr = nil
			}
			schdl.lastcmd = noAction
			schdl.cmdDone <- true
			break
		}
	}
}

func executeScheduleActions(schdl *Schedule, tikStamp time.Time) {
	if actions := schdl.actions[:]; len(actions) > 0 {
		schdlname := schdl.schdlname
		wg := &sync.WaitGroup{}
		wg.Add(len(actions))
		for _, a := range actions {
			go func() {
				defer wg.Done()
				a(schdlname, tikStamp)
			}()
		}
		wg.Wait()
	}
}

type Action struct {
	name     string
	prev     string
	next     string
	stp      *Step
	flw      *Flow
	acnHndlr ActionHandler
	params   *lnksworks.Parameters
}

func (acn *Action) DBManager() *lnksworks.DbManager {
	return acn.flw.DBManager()
}

type ActionHandler = func(action *Action, step *Step, flow *Flow) (nxtaction *Action)

type Step struct {
	flw           *Flow
	name          string
	actions       map[string]*Action
	prev          string
	next          string
	done          chan bool
	queuedActions chan *Action
	stpHndlr      StepHandler
	params        *lnksworks.Parameters
}

func (stp *Step) DBManager() *lnksworks.DbManager {
	return stp.flw.DBManager()
}

type StepHandler = func(step *Step, flow *Flow) (nxtstep *Step)

type Flow struct {
	dbmnr       *lnksworks.DbManager
	name        string
	prev        *Flow
	next        *Flow
	steps       map[string]*Step
	curStep     *Step
	interval    *time.Duration
	queuedSteps chan *Step
	flwHndlr    FlowHandler
	params      *lnksworks.Parameters
}

func (flw *Flow) DBManager() *lnksworks.DbManager {
	if flw.dbmnr == nil {
		flw.dbmnr = lnksworks.DatabaseManager()
	}
	return flw.dbmnr
}

type FlowHandler = func(flow *Flow) (nxtflow *Flow)

func NewFlow(name string, interval time.Duration) (flw *Flow) {
	flw = &Flow{}

	return flw
}

var queuedFlows []chan *Flow
var queuedFlowsLck *sync.RWMutex
var queuedFlowsI int

var queuedSteps []chan *Step
var queuedStepsLck *sync.RWMutex
var queuedStepsI int

var queuedActions []chan *Action
var queuedActionsLck *sync.RWMutex
var queuedActionsI int

func init() {

	if schedules == nil {
		schedules = map[string]*Schedule{}
	}
	if schdlslck == nil {
		schdlslck = &sync.RWMutex{}
	}
	if scheduledQueue == nil {
		scheduledQueue = make(chan *Schedule)

		go processQueuedChannels()
	}

	if queuedFlowsLck == nil {
		queuedFlowsLck = &sync.RWMutex{}
	}
	if queuedFlows == nil {
		qflwL := 15
		qflwi := 0
		queuedFlows = make([]chan *Flow, qflwL)

		for qflwi < qflwL {
			queuedFlows[qflwi] = make(chan *Flow)
			go processQueuedFlows(qflwi)
			qflwi++
		}
	}

	if queuedStepsLck == nil {
		queuedStepsLck = &sync.RWMutex{}
	}
	if queuedSteps == nil {
		qstpL := 15
		qstpi := 0
		queuedSteps = make([]chan *Step, qstpL)

		for qstpi < qstpL {
			queuedSteps[qstpi] = make(chan *Step)
			go processQueuedSteps(qstpi)
			qstpi++
		}
	}

	if queuedActionsLck == nil {
		queuedActionsLck = &sync.RWMutex{}
	}
	if queuedActions == nil {
		qacnL := 15
		qacni := 0
		queuedActions = make([]chan *Action, qacnL)

		for qacni < qacnL {
			queuedActions[qacni] = make(chan *Action)
			go processQueuedActions(qacni)
			qacni++
		}
	}
}

//Fow queueing

func enQueueFlow(flw *Flow) {
	queuedFlowsLck.RLock()
	defer queuedFlowsLck.RUnlock()
	if queuedFlowsI >= len(queuedFlows) {
		queuedFlowsI = 0
	}
	queuedFlows[queuedFlowsI] <- flw
	queuedFlowsI++
}

func processQueuedFlows(qflwi int) {
	for {
		select {
		case curFlw := <-queuedFlows[qflwi]:
			go enQueueFlow(executeFlow(curFlw))
		}
	}
}

func executeFlow(flw *Flow) (prvNxtFlw *Flow) {
	if flw.flwHndlr != nil {
		prvNxtFlw = flw.flwHndlr(flw)
	}
	return prvNxtFlw
}

//Step queueing

func enQueueStep(stp *Step) {
	queuedStepsLck.RLock()
	defer queuedStepsLck.RUnlock()
	if queuedStepsI >= len(queuedSteps) {
		queuedStepsI = 0
	}
	queuedSteps[queuedStepsI] <- stp
	queuedStepsI++
}

func processQueuedSteps(qstpi int) {
	for {
		select {
		case curStp := <-queuedSteps[qstpi]:
			go enQueueStep(executeStep(curStp))
		}
	}
}

func executeStep(stp *Step) (prvNxtStp *Step) {
	if stp.stpHndlr != nil {
		prvNxtStp = stp.stpHndlr(stp, stp.flw)
	}
	return prvNxtStp
}

//Action queueing

func enQueueAction(acn *Action) {
	queuedActionsLck.RLock()
	defer queuedActionsLck.RUnlock()
	if queuedActionsI >= len(queuedActions) {
		queuedActionsI = 0
	}
	queuedActions[queuedActionsI] <- acn
	queuedActionsI++
}

func processQueuedActions(qacni int) {
	for {
		select {
		case curAcn := <-queuedActions[qacni]:
			go enQueueAction(executeAction(curAcn))
		}
	}
}

func executeAction(acn *Action) (prvNxtAcn *Action) {
	if acn.acnHndlr != nil {
		prvNxtAcn = acn.acnHndlr(acn, acn.stp, acn.flw)
	}
	return prvNxtAcn
}

//Support io.Reader and string
func ReadAndExecuteFowInstructions(a ...interface{}) {

}
