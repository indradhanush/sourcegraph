package background

import (
	"context"

	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/lib/log"
)

func NewBackgroundJobs(logger log.Logger, db edb.EnterpriseDB) []goroutine.BackgroundRoutine {
	codeMonitorsStore := db.CodeMonitors()

	triggerMetrics := newMetricsForTriggerQueries()
	actionMetrics := newActionMetrics()

	// Create a new context. Each background routine will wrap this with
	// a cancellable context that is canceled when Stop() is called.
	ctx := context.Background()
	return []goroutine.BackgroundRoutine{
		newTriggerQueryEnqueuer(ctx, codeMonitorsStore),
		newTriggerJobsLogDeleter(ctx, codeMonitorsStore),
		newTriggerQueryRunner(ctx, logger, db, triggerMetrics),
		newTriggerQueryResetter(ctx, codeMonitorsStore, triggerMetrics),
		newActionRunner(ctx, logger, codeMonitorsStore, actionMetrics),
		newActionJobResetter(ctx, codeMonitorsStore, actionMetrics),
	}
}
