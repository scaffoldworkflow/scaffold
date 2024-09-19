package scron

import (
	// "scaffold/server/bulwark"

	"scaffold/server/constants"
	"scaffold/server/history"
	"scaffold/server/msg"
	"scaffold/server/rabbitmq"
	"scaffold/server/state"
	"scaffold/server/task"
	"scaffold/server/utils"
	"scaffold/server/workflow"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	logger "github.com/jfcarter2358/go-logger"

	"github.com/robfig/cron"
)

func Start() {
	c := cron.New()
	c.AddFunc("* * * * * *", checkTaskCrons)
	go c.Start()
}

func checkTaskCrons() {
	ts, err := task.GetAllTasks()
	if err != nil {
		logger.Errorf("", "Unable to get tasks: %s", err.Error())
	}
	currentTime := time.Now()

	for _, t := range ts {
		if t.Cron != "" && !t.Disabled {
			c, err := workflow.GetWorkflowByName(t.Workflow)
			if err != nil {
				logger.Errorf("", "Error getting workflow: %s", err.Error())
				continue
			}
			valid, err := task.VerifyDepends(t.Workflow, t.Name)
			if err != nil {
				logger.Errorf("", "Error verify tasks parent statuses: %s", err.Error())
				continue
			}
			if !valid {
				continue
			}
			checkCron(currentTime, t.Cron, t.Name, t.RunNumber, c)
		}
		// if t.Check.Cron != "" && !t.Disabled {
		// 	s, err := state.GetStateByNames(t.Workflow, t.Name)
		// 	if err != nil {
		// 		logger.Errorf("", "Error getting state: %s", err.Error())
		// 		continue
		// 	}
		// 	if s.Status != constants.STATE_STATUS_ERROR && s.Status != constants.STATE_STATUS_RUNNING {
		// 		continue
		// 	}
		// 	valid, err := task.VerifyDepends(t.Workflow, t.Name)
		// 	if err != nil {
		// 		logger.Errorf("", "Error verify tasks parent statuses: %s", err.Error())
		// 		continue
		// 	}
		// 	if !valid {
		// 		continue
		// 	}
		// 	c, err := workflow.GetWorkflowByName(t.Workflow)
		// 	if err != nil {
		// 		logger.Errorf("", "Error getting workflow: %s", err.Error())
		// 		continue
		// 	}
		// 	checkCron(currentTime, t.Cron, fmt.Sprintf("SCAFFOLD-CHECK_%s", t.Name), t.Check.RunNumber, c)
		// }
	}
}

func checkCron(currentTime time.Time, crontab, name string, runNumber int, c *workflow.Workflow) {
	second := currentTime.Second()
	month := currentTime.Month()
	day := currentTime.Day()
	hour := currentTime.Hour()
	minute := currentTime.Minute()
	dayOfWeek := currentTime.Weekday()

	parts := strings.Split(crontab, " ")
	isSecond := checkCronValue(int(second), 0, 59, parts[0])
	isMinute := checkCronValue(int(minute), 0, 59, parts[1])
	isHour := checkCronValue(int(hour), 0, 23, parts[2])
	isDay := checkCronValue(int(day), 0, 31, parts[3])
	isMonth := checkCronValue(int(month), 1, 12, parts[4])
	isDayOfWeek := checkCronValue(int(dayOfWeek), 0, 7, parts[5])

	if isSecond && isMinute && isHour && isDay && isMonth && isDayOfWeek {
		t, err := task.GetTaskByNames(c.Name, name)
		if err != nil {
			logger.Errorf("", "Error getting cron run task: %s", err.Error())
			return
		}
		if t.Disabled {
			return
		}

		for _, tt := range t.DependsOn.Success {
			s, err := state.GetStateByNames(c.Name, tt)
			if err != nil {
				logger.Errorf("", "Error getting cron run state: %s", err.Error())
				return
			}
			if s.Status != constants.STATE_STATUS_SUCCESS {
				logger.Tracef("", "Cron status of %s does not match %s", s.Status, constants.STATE_STATUS_SUCCESS)
				return
			}
		}
		for _, tt := range t.DependsOn.Error {
			s, err := state.GetStateByNames(c.Name, tt)
			if err != nil {
				logger.Errorf("", "Error getting cron run state: %s", err.Error())
				return
			}
			if s.Status != constants.STATE_STATUS_ERROR {
				logger.Tracef("", "Cron status of %s does not match %s", s.Status, constants.STATE_STATUS_ERROR)
				return
			}
		}
		for _, tt := range t.DependsOn.Always {
			s, err := state.GetStateByNames(c.Name, tt)
			if err != nil {
				logger.Errorf("", "Error getting cron run state: %s", err.Error())
				return
			}
			if s.Status != constants.STATE_STATUS_SUCCESS && s.Status != constants.STATE_STATUS_ERROR {
				logger.Tracef("", "Cron status of %s does not match %s or %s", s.Status, constants.STATE_STATUS_SUCCESS, constants.STATE_STATUS_ERROR)
				return
			}
		}
		s, err := state.GetStateByNames(c.Name, name)
		if err != nil {
			logger.Errorf("", "Error getting cron run state: %s", err.Error())
			return
		}

		runID := uuid.New().String()

		m := msg.TriggerMsg{
			Task:     name,
			Workflow: c.Name,
			Action:   constants.ACTION_TRIGGER,
			Groups:   c.Groups,
			Number:   runNumber + 1,
			Context:  s.Context,
		}
		h := history.History{
			RunID:    runID,
			States:   make([]state.State, 0),
			Workflow: c.Name,
		}

		if err := history.CreateHistory(&h); err != nil {
			logger.Errorf("", "Error triggering cron run: %s", err.Error())
			return
		}

		logger.Infof("", "Triggering run with message %v", m)
		if err := rabbitmq.ManagerPublish(m); err != nil {
			logger.Errorf("", "Error triggering cron run: %s", err.Error())
		}
	}
}

func checkCronValue(t, start, end int, x string) bool {
	step := 1
	vals := []int{}

	parts_slash := strings.Split(x, "/")
	if len(parts_slash) == 2 {
		step, _ = strconv.Atoi(parts_slash[1])
		x = parts_slash[0]
	}

	if strings.HasPrefix(x, "*") {
		for i := start; i < end+1; i += step {
			vals = append(vals, i)
		}
	} else {
		parts_dash := strings.Split(x, "-")
		if len(parts_dash) == 2 {
			start, _ = strconv.Atoi(parts_dash[0])
			end, _ = strconv.Atoi(parts_dash[1])
			for i := start; i < end+1; i += step {
				vals = append(vals, i)
			}
		} else {
			parts_comma := strings.Split(x, ",")
			if len(parts_comma) > 1 {
				for _, i := range parts_comma {
					ii, _ := strconv.Atoi(i)
					vals = append(vals, ii)
				}
			} else {
				ii, _ := strconv.Atoi(x)
				vals = append(vals, ii)
			}
		}
	}
	return utils.ContainsInt(vals, t)
}
