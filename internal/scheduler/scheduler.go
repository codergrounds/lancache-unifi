package scheduler

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/robfig/cron/v3"
)

// Scheduler wraps a cron scheduler to run a sync function on a schedule.
type Scheduler struct {
	cron     *cron.Cron
	schedule string
}

// New creates a Scheduler. If cronExpr is empty, a random daily schedule is generated.
func New(cronExpr string) *Scheduler {
	if cronExpr == "" {
		cronExpr = randomDailySchedule()
		log.Printf("[INFO] no CRON_SCHEDULE set, using random daily schedule: %s", cronExpr)
	} else {
		log.Printf("[INFO] using configured cron schedule: %s", cronExpr)
	}

	return &Scheduler{
		cron:     cron.New(),
		schedule: cronExpr,
	}
}

// Start adds the job to the cron scheduler and starts it.
// The provided function is called on each tick. Start blocks forever.
func (s *Scheduler) Start(fn func() error) error {
	_, err := s.cron.AddFunc(s.schedule, func() {
		log.Println("[INFO] cron trigger: starting scheduled sync")
		if err := fn(); err != nil {
			log.Printf("[ERROR] scheduled sync failed: %v", err)
		}
	})
	if err != nil {
		return fmt.Errorf("invalid cron expression %q: %w", s.schedule, err)
	}

	s.cron.Start()
	log.Printf("[INFO] scheduler started, next run: %s", s.nextRun())

	// Block forever
	select {}
}

// Schedule returns the cron expression being used.
func (s *Scheduler) Schedule() string {
	return s.schedule
}

func (s *Scheduler) nextRun() string {
	entries := s.cron.Entries()
	if len(entries) == 0 {
		return "unknown"
	}
	return entries[0].Next.Format(time.RFC3339)
}

// randomDailySchedule generates a random cron expression to run once per day.
// The minute and hour are randomised at startup to spread load when multiple
// instances exist.
func randomDailySchedule() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	minute := r.Intn(60)
	hour := r.Intn(24)
	return fmt.Sprintf("%d %d * * *", minute, hour)
}
