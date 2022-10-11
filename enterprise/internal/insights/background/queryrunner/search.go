package queryrunner

import (
	"context"
	"time"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/discovery"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/query/streaming"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func GetSearchHandlers() map[types.GenerationMethod]InsightsHandler {

	searchStream := func(ctx context.Context, query string) (*streaming.TabulationResult, error) {
		decoder, streamResults := streaming.TabulationDecoder()
		err := streaming.Search(ctx, query, nil, decoder)
		if err != nil {
			return nil, errors.Wrap(err, "streaming.Search")
		}
		return streamResults, nil
	}

	computeSearchStream := func(ctx context.Context, query string) (*streaming.ComputeTabulationResult, error) {
		decoder, streamResults := streaming.MatchContextComputeDecoder()
		err := streaming.ComputeMatchContextStream(ctx, query, decoder)
		if err != nil {
			return nil, errors.Wrap(err, "streaming.Compute")
		}
		return streamResults, nil
	}

	computeTextExtraSearch := func(ctx context.Context, query string) (*streaming.ComputeTabulationResult, error) {
		decoder, streamResults := streaming.ComputeTextDecoder()
		err := streaming.ComputeTextExtraStream(ctx, query, decoder)
		if err != nil {
			return nil, errors.Wrap(err, "streaming.ComputeText")
		}
		return streamResults, nil
	}

	return map[types.GenerationMethod]InsightsHandler{
		types.MappingCompute: makeMappingComputeHandler(computeTextExtraSearch),
		types.SearchCompute:  makeComputeHandler(computeSearchStream),
		types.Search:         makeSearchHandler(searchStream),
	}

}

// checkSubRepoPermissions returns true if the repo has sub-repo permissions or any error occurred while checking it
// Returns false only if the repo doesn't have sub-repo permissions or these are disabled in settings.
// Note that repo ID is received untyped and being cast to api.RepoID
// err is an upstream error to which any new occurring error is appended
func checkSubRepoPermissions(ctx context.Context, checker authz.SubRepoPermissionChecker, untypedRepoID any, err error) (bool, error) {
	if !authz.SubRepoEnabled(checker) {
		return false, err
	}

	// casting repoID
	var repoID api.RepoID
	switch untypedRepoID := untypedRepoID.(type) {
	case api.RepoID:
		repoID = untypedRepoID
	case string:
		var idErr error
		repoID, idErr = graphqlbackend.UnmarshalRepositoryID(graphql.ID(untypedRepoID))
		if idErr != nil {
			err = errors.Append(err, errors.Wrap(idErr, "Checking sub-repo permissions: UnmarshalRepositoryID"))
			return true, err
		}
	default:
		return true, errors.Append(err, errors.Newf("Checking sub-repo permissions for repoID=%v: Unsupported untypedRepoID type=%T",
			untypedRepoID, untypedRepoID))
	}

	// performing the check itself
	enabled, checkErr := authz.SubRepoEnabledForRepoID(ctx, checker, repoID)
	if checkErr != nil {
		err = errors.Append(err, errors.Wrap(checkErr, "Checking sub-repo permissions"))
		return true, err
	}
	return enabled, err
}

func toRecording(record *Job, value float64, recordTime time.Time, repoName string, repoID api.RepoID, capture *string) []store.RecordSeriesPointArgs {
	args := make([]store.RecordSeriesPointArgs, 0, len(record.DependentFrames)+1)
	base := store.RecordSeriesPointArgs{
		SeriesID: record.SeriesID,
		Point: store.SeriesPoint{
			SeriesID: record.SeriesID,
			Time:     recordTime,
			Value:    value,
			Capture:  capture,
		},
		RepoName:    &repoName,
		RepoID:      &repoID,
		PersistMode: store.PersistMode(record.PersistMode),
	}
	args = append(args, base)
	for _, dependent := range record.DependentFrames {
		arg := base
		arg.Point.Time = dependent
		args = append(args, arg)
	}
	return args
}

type streamComputeProvider func(context.Context, string) (*streaming.ComputeTabulationResult, error)
type streamSearchProvider func(context.Context, string) (*streaming.TabulationResult, error)

func generateComputeRecordingsStream(ctx context.Context, job *Job, recordTime time.Time, provider streamComputeProvider) (_ []store.RecordSeriesPointArgs, err error) {
	streamResults, err := provider(ctx, job.SearchQuery)
	if err != nil {
		return nil, err
	}
	if len(streamResults.Errors) > 0 {
		return nil, classifiedError(streamResults.Errors, types.SearchCompute)
	}
	if len(streamResults.Alerts) > 0 {
		return nil, errors.Errorf("compute streaming search: alerts: %v", streamResults.Alerts)
	}

	checker := authz.DefaultSubRepoPermsChecker
	var recordings []store.RecordSeriesPointArgs

	for _, match := range streamResults.RepoCounts {
		var subRepoEnabled bool
		subRepoEnabled, err = checkSubRepoPermissions(ctx, checker, match.RepositoryID, err)
		if subRepoEnabled {
			continue
		}

		for capturedValue, count := range match.ValueCounts {
			capture := capturedValue
			if len(capture) == 0 {
				// there seems to be some behavior where empty string values get returned from the compute API. We will just skip them. If there are future changes
				// to fix this, we will automatically pick up any new results without changes here.
				continue
			}
			recordings = append(recordings, toRecording(job, float64(count), recordTime, match.RepositoryName, api.RepoID(match.RepositoryID), &capture)...)
		}
	}
	return recordings, nil
}

func generateSearchRecordingsStream(ctx context.Context, job *Job, recordTime time.Time, provider streamSearchProvider) ([]store.RecordSeriesPointArgs, error) {
	tabulationResult, err := provider(ctx, job.SearchQuery)
	if err != nil {
		return nil, err
	}

	tr := *tabulationResult

	if len(tr.Errors) > 0 {
		return nil, classifiedError(tr.Errors, types.Search)
	}
	if len(tr.Alerts) > 0 {
		return nil, errors.Errorf("streaming search: alerts: %v", tr.Alerts)
	}

	checker := authz.DefaultSubRepoPermsChecker
	var recordings []store.RecordSeriesPointArgs

	for _, match := range tr.RepoCounts {
		// sub-repo permissions filtering. If the repo supports it, then it should be excluded from search results
		var subRepoEnabled bool
		repoID := api.RepoID(match.RepositoryID)
		subRepoEnabled, err = checkSubRepoPermissions(ctx, checker, repoID, err)
		if subRepoEnabled {
			continue
		}
		recordings = append(recordings, toRecording(job, float64(match.MatchCount), recordTime, match.RepositoryName, repoID, nil)...)
	}
	return recordings, nil
}

func makeSearchHandler(provider streamSearchProvider) InsightsHandler {
	return func(ctx context.Context, job *Job, series *types.InsightSeries, recordTime time.Time) ([]store.RecordSeriesPointArgs, error) {
		recordings, err := generateSearchRecordingsStream(ctx, job, recordTime, provider)
		if err != nil {
			return nil, errors.Wrapf(err, "searchHandler")
		}
		return recordings, nil
	}
}

func makeComputeHandler(provider streamComputeProvider) InsightsHandler {
	return func(ctx context.Context, job *Job, series *types.InsightSeries, recordTime time.Time) ([]store.RecordSeriesPointArgs, error) {
		computeDelegate := func(ctx context.Context, job *Job, recordTime time.Time) (_ []store.RecordSeriesPointArgs, err error) {
			return generateComputeRecordingsStream(ctx, job, recordTime, provider)
		}
		recordings, err := computeDelegate(ctx, job, recordTime)
		if err != nil {
			return nil, errors.Wrapf(err, "computeHandler")
		}
		return recordings, nil
	}
}

func makeMappingComputeHandler(provider streamComputeProvider) InsightsHandler {
	return func(ctx context.Context, job *Job, series *types.InsightSeries, recordTime time.Time) ([]store.RecordSeriesPointArgs, error) {
		recordings, err := generateComputeRecordingsStream(ctx, job, recordTime, provider)
		if err != nil {
			return nil, errors.Wrapf(err, "mappingComputeHandler")
		}
		return recordings, err
	}
}

func (r *workHandler) persistRecordings(ctx context.Context, job *Job, series *types.InsightSeries, recordings []store.RecordSeriesPointArgs, recordTime time.Time) (err error) {
	tx, err := r.insightsStore.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	if store.PersistMode(job.PersistMode) == store.SnapshotMode {
		// The purpose of the snapshot is for low fidelity but recently updated data points.
		// We store one snapshot of an insight at any time, so we prune the table whenever adding a new series.
		if err := tx.DeleteSnapshots(ctx, series); err != nil {
			return errors.Wrap(err, "DeleteSnapshots")
		}
		if err := tx.SetInsightSeriesRecordingTimes(ctx, []types.InsightSeriesRecordingTimes{{series.SeriesID, []time.Time{recordTime}}}); err != nil {
			return errors.Wrap(err, "SetInsightSeriesRecordingTimes")
		}
	} else {
		// This will fetch insight series recording times even for jobs for a backfill when setting recording times
		// isn't necessary, as we don't have a way of determining whether this is a backfill record job.
		seriesRecordingTimes, err := tx.GetInsightSeriesRecordingTimes(ctx, series.SeriesID)
		if err != nil {
			return errors.Wrap(err, "GetInsightSeriesRecordingTimes")
		}
		newRecordingTimes := updateSeriesRecordingTimes(seriesRecordingTimes.RecordingTimes, recordTime)
		if len(newRecordingTimes) > 0 {
			if err := tx.SetInsightSeriesRecordingTimes(ctx, []types.InsightSeriesRecordingTimes{{series.SeriesID, newRecordingTimes}}); err != nil {
				return errors.Wrap(err, "SetInsightSeriesRecordingTimes")
			}
		}
	}

	// Newly queued queries should be scoped to correct repos however leaving filtering
	// in place to ensure any older queued jobs get filtered properly. It's a noop for global insights.
	filteredRecordings, err := filterRecordingsBySeriesRepos(ctx, r.repoStore, series, recordings)
	if err != nil {
		return errors.Wrap(err, "filterRecordingsBySeriesRepos")
	}

	if recordErr := tx.RecordSeriesPoints(ctx, filteredRecordings); recordErr != nil {
		err = errors.Append(err, errors.Wrap(recordErr, "RecordSeriesPointsCapture"))
	}
	return err
}

func updateSeriesRecordingTimes(recordingTimes []time.Time, newTime time.Time) []time.Time {
	var newRecordingTimes []time.Time
	aYearAgo := newTime.AddDate(-1, 0, 0)
	if len(recordingTimes) < 12 {
		newRecordingTimes = append(recordingTimes, newTime)
	} else if len(recordingTimes) >= 12 && recordingTimes[0].Before(aYearAgo) {
		// We replace the first recording time (shift left).
		newRecordingTimes = append(recordingTimes[1:], newTime)
	} else if len(recordingTimes) >= 12 && recordingTimes[0].After(aYearAgo) {
		// This is an insight over less than a year ago, so we can just append to the list.
		newRecordingTimes = append(recordingTimes, newTime)
	}
	return newRecordingTimes
}

func filterRecordingsBySeriesRepos(ctx context.Context, repoStore discovery.RepoStore, series *types.InsightSeries, recordings []store.RecordSeriesPointArgs) ([]store.RecordSeriesPointArgs, error) {
	// If this series isn't scoped to some repos return all
	if len(series.Repositories) == 0 {
		return recordings, nil
	}

	seriesRepos, err := repoStore.List(ctx, database.ReposListOptions{Names: series.Repositories})
	if err != nil {
		return nil, errors.Wrap(err, "repoStore.List")
	}
	repos := map[api.RepoID]bool{}
	for _, repo := range seriesRepos {
		repos[repo.ID] = true
	}

	filteredRecords := make([]store.RecordSeriesPointArgs, 0, len(series.Repositories))
	for _, record := range recordings {
		if record.RepoID == nil {
			continue
		}
		if included := repos[*record.RepoID]; included == true {
			filteredRecords = append(filteredRecords, record)
		}
	}
	return filteredRecords, nil

}
