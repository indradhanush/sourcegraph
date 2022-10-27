package sources

import (
	"context"
	"fmt"
	"strconv"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

type BitbucketServerSource struct {
	client *bitbucketserver.Client
	au     auth.Authenticator
}

var _ ForkableChangesetSource = BitbucketServerSource{}

// NewBitbucketServerSource returns a new BitbucketServerSource from the given external service.
func NewBitbucketServerSource(ctx context.Context, svc *types.ExternalService, cf *httpcli.Factory) (*BitbucketServerSource, error) {
	rawConfig, err := svc.Config.Decrypt(ctx)
	if err != nil {
		return nil, errors.Errorf("external service id=%d config error: %s", svc.ID, err)
	}
	var c schema.BitbucketServerConnection
	if err := jsonc.Unmarshal(rawConfig, &c); err != nil {
		return nil, errors.Errorf("external service id=%d config error: %s", svc.ID, err)
	}

	if cf == nil {
		cf = httpcli.ExternalClientFactory
	}

	opts := httpClientCertificateOptions(nil, c.Certificate)

	cli, err := cf.Doer(opts...)
	if err != nil {
		return nil, err
	}

	client, err := bitbucketserver.NewClient(svc.URN(), &c, cli)
	if err != nil {
		return nil, err
	}

	return &BitbucketServerSource{
		au:     client.Auth,
		client: client,
	}, nil
}

func (s BitbucketServerSource) GitserverPushConfig(repo *types.Repo) (*protocol.PushConfig, error) {
	return GitserverPushConfig(repo, s.au)
}

func (s BitbucketServerSource) WithAuthenticator(a auth.Authenticator) (ChangesetSource, error) {
	switch a.(type) {
	case *auth.OAuthBearerToken,
		*auth.OAuthBearerTokenWithSSH,
		*auth.BasicAuth,
		*auth.BasicAuthWithSSH,
		*bitbucketserver.SudoableOAuthClient:
		break

	default:
		return nil, newUnsupportedAuthenticatorError("BitbucketServerSource", a)
	}

	return &BitbucketServerSource{
		client: s.client.WithAuthenticator(a),
		au:     a,
	}, nil
}

// AuthenticatedUsername uses the underlying bitbucketserver.Client to get the
// username belonging to the credentials associated with the
// BitbucketServerSource.
func (s BitbucketServerSource) AuthenticatedUsername(ctx context.Context) (string, error) {
	return s.client.AuthenticatedUsername(ctx)
}

func (s BitbucketServerSource) ValidateAuthenticator(ctx context.Context) error {
	_, err := s.client.AuthenticatedUsername(ctx)
	return err
}

// CreateChangeset creates the given *Changeset in the code host.
func (s BitbucketServerSource) CreateChangeset(ctx context.Context, c *Changeset) (bool, error) {
	var exists bool

	remoteRepo := c.RemoteRepo.Metadata.(*bitbucketserver.Repo)
	targetRepo := c.TargetRepo.Metadata.(*bitbucketserver.Repo)

	pr := &bitbucketserver.PullRequest{Title: c.Title, Description: c.Body}

	pr.ToRef.Repository.Slug = targetRepo.Slug
	pr.ToRef.Repository.ID = targetRepo.ID
	pr.ToRef.Repository.Project.Key = targetRepo.Project.Key
	pr.ToRef.ID = gitdomain.EnsureRefPrefix(c.BaseRef)

	pr.FromRef.Repository.Slug = remoteRepo.Slug
	pr.FromRef.Repository.ID = remoteRepo.ID
	pr.FromRef.Repository.Project.Key = remoteRepo.Project.Key
	pr.FromRef.ID = gitdomain.EnsureRefPrefix(c.HeadRef)

	err := s.client.CreatePullRequest(ctx, pr)
	if err != nil {
		var e *bitbucketserver.ErrAlreadyExists
		if errors.As(err, &e) {
			if e.Existing == nil {
				return exists, errors.Errorf("existing PR is nil")
			}
			log15.Info("Existing PR extracted", "ID", e.Existing.ID)
			pr = e.Existing
			exists = true
		} else {
			return exists, err
		}
	}

	if err := s.loadPullRequestData(ctx, pr); err != nil {
		return false, errors.Wrap(err, "loading extra metadata")
	}
	if err = c.SetMetadata(pr); err != nil {
		return false, errors.Wrap(err, "setting changeset metadata")
	}

	return exists, nil
}

// CloseChangeset closes the given *Changeset on the code host and updates the
// Metadata column in the *batches.Changeset to the newly closed pull request.
func (s BitbucketServerSource) CloseChangeset(ctx context.Context, c *Changeset) error {
	declined, err := s.callAndRetryIfOutdated(ctx, c, s.client.DeclinePullRequest)
	if err != nil {
		return err
	}

	return c.Changeset.SetMetadata(declined)
}

// LoadChangeset loads the latest state of the given Changeset from the codehost.
func (s BitbucketServerSource) LoadChangeset(ctx context.Context, cs *Changeset) error {
	repo := cs.TargetRepo.Metadata.(*bitbucketserver.Repo)
	number, err := strconv.Atoi(cs.ExternalID)
	if err != nil {
		return err
	}

	pr := &bitbucketserver.PullRequest{ID: number}
	pr.ToRef.Repository.Slug = repo.Slug
	pr.ToRef.Repository.Project.Key = repo.Project.Key

	err = s.client.LoadPullRequest(ctx, pr)
	if err != nil {
		if err == bitbucketserver.ErrPullRequestNotFound {
			return ChangesetNotFoundError{Changeset: cs}
		}

		return err
	}

	err = s.loadPullRequestData(ctx, pr)
	if err != nil {
		return errors.Wrap(err, "loading pull request data")
	}
	if err = cs.SetMetadata(pr); err != nil {
		return errors.Wrap(err, "setting changeset metadata")
	}

	return nil
}

func (s BitbucketServerSource) loadPullRequestData(ctx context.Context, pr *bitbucketserver.PullRequest) error {
	if err := s.client.LoadPullRequestActivities(ctx, pr); err != nil {
		return errors.Wrap(err, "loading pr activities")
	}

	if err := s.client.LoadPullRequestCommits(ctx, pr); err != nil {
		return errors.Wrap(err, "loading pr commits")
	}

	if err := s.client.LoadPullRequestBuildStatuses(ctx, pr); err != nil {
		return errors.Wrap(err, "loading pr build status")
	}

	return nil
}

func (s BitbucketServerSource) UpdateChangeset(ctx context.Context, c *Changeset) error {
	pr, ok := c.Changeset.Metadata.(*bitbucketserver.PullRequest)
	if !ok {
		return errors.New("Changeset is not a Bitbucket Server pull request")
	}

	update := &bitbucketserver.UpdatePullRequestInput{
		PullRequestID: strconv.Itoa(pr.ID),
		Title:         c.Title,
		Description:   c.Body,
		Version:       pr.Version,
	}
	update.ToRef.ID = c.BaseRef
	update.ToRef.Repository.Slug = pr.ToRef.Repository.Slug
	update.ToRef.Repository.Project.Key = pr.ToRef.Repository.Project.Key

	updated, err := s.client.UpdatePullRequest(ctx, update)
	if err != nil {
		if !bitbucketserver.IsPullRequestOutOfDate(err) {
			return err
		}

		// If we have an outdated version of the pull request we extract the
		// pull request that was returned with the error...
		newestPR, err2 := bitbucketserver.ExtractPullRequest(err)
		if err2 != nil {
			return errors.Wrap(err, "failed to extract pull request after receiving error")
		}

		log15.Info("Updating Bitbucket Server PR failed because it's outdated. Retrying with newer version", "ID", pr.ID, "oldVersion", pr.Version, "newestVerssion", newestPR.Version)

		// ... and try again, but this time with the newest version
		update.Version = newestPR.Version
		updated, err = s.client.UpdatePullRequest(ctx, update)
		if err != nil {
			// If that didn't work, we bail out
			return err
		}
	}

	return c.Changeset.SetMetadata(updated)
}

// ReopenChangeset reopens the *Changeset on the code host and updates the
// Metadata column in the *batches.Changeset.
func (s BitbucketServerSource) ReopenChangeset(ctx context.Context, c *Changeset) error {
	reopened, err := s.callAndRetryIfOutdated(ctx, c, s.client.ReopenPullRequest)
	if err != nil {
		return err

	}

	return c.Changeset.SetMetadata(reopened)
}

// CreateComment posts a comment on the Changeset.
func (s BitbucketServerSource) CreateComment(ctx context.Context, c *Changeset, text string) error {
	// Bitbucket Server seems to ignore version conflicts when commenting, but
	// we use this here anyway.
	_, err := s.callAndRetryIfOutdated(ctx, c, func(ctx context.Context, pr *bitbucketserver.PullRequest) error {
		return s.client.CreatePullRequestComment(ctx, pr, text)
	})
	return err
}

// MergeChangeset merges a Changeset on the code host, if in a mergeable state.
// The squash parameter is ignored, as Bitbucket Server does not support
// squash merges.
func (s BitbucketServerSource) MergeChangeset(ctx context.Context, c *Changeset, squash bool) error {
	merged, err := s.callAndRetryIfOutdated(ctx, c, s.client.MergePullRequest)
	if err != nil {
		if bitbucketserver.IsMergePreconditionFailedException(err) {
			return &ChangesetNotMergeableError{ErrorMsg: err.Error()}
		}
		return err
	}

	return c.Changeset.SetMetadata(merged)
}

type bitbucketClientFunc func(context.Context, *bitbucketserver.PullRequest) error

func (s BitbucketServerSource) callAndRetryIfOutdated(ctx context.Context, c *Changeset, fn bitbucketClientFunc) (*bitbucketserver.PullRequest, error) {
	pr, ok := c.Changeset.Metadata.(*bitbucketserver.PullRequest)
	if !ok {
		return nil, errors.New("Changeset is not a Bitbucket Server pull request")
	}

	err := fn(ctx, pr)
	if err == nil {
		return pr, nil
	}

	if !bitbucketserver.IsPullRequestOutOfDate(err) {
		return nil, err
	}

	// If we have an outdated version of the pull request we extract the
	// pull request that was returned with the error...
	newestPR, err2 := bitbucketserver.ExtractPullRequest(err)
	if err2 != nil {
		return nil, errors.Wrap(err, "failed to extract pull request after receiving error")
	}

	log15.Info("Retrying Bitbucket Server operation because local PR is outdated. Retrying with newer version", "ID", pr.ID, "oldVersion", pr.Version, "newestVerssion", newestPR.Version)

	// ... and try again, but this time with the newest version
	err = fn(ctx, newestPR)
	if err != nil {
		return nil, err
	}

	return newestPR, nil
}

// GetUserFork returns a repo pointing to a fork of the given repo in the
// currently authenticated user's namespace.
func (s BitbucketServerSource) GetUserFork(ctx context.Context, targetRepo *types.Repo) (*types.Repo, error) {
	parent := targetRepo.Metadata.(*bitbucketserver.Repo)

	// Ascertain the user name for the token we're using.
	user, err := s.AuthenticatedUsername(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "getting username")
	}

	// We have to prepend a tilde to the user name to make this a "user-centric URL" in
	// Bitbucket Server parlance.
	forkNamespace := "~" + user

	// See if we already have a fork.
	fork, err := s.getFork(ctx, parent, forkNamespace)
	if err != nil && !bitbucketserver.IsNotFound(err) {
		return nil, errors.Wrapf(err, "getting user fork for %q", user)
	}

	// If not, then we need to create a fork.
	if fork == nil {
		// fork, err = s.client.Fork(ctx, parent.Project.Key, parent.Slug, bitbucketserver.CreateForkInput{})

		fmt.Printf("GOING TO CREATE FORK! %v-%v ID: %v", parent.Project.Key, parent.Slug, parent.ID)
		createdFork := parent.Project.Key + "-" + parent.Slug
		fork, err = s.client.Fork(ctx, parent.Project.Key, parent.Slug, bitbucketserver.CreateForkInput{Name: &createdFork})

		fmt.Printf("FORK CREATED SUCCESSFULLY! Origin ID: %v", fork.Origin.ID)
		if err != nil {
			return nil, errors.Wrapf(err, "creating user fork for %q", user)
		}
		fmt.Printf("GetUserFork - Created FORK Name %v-%v ID: %v", parent.Project.Key, parent.Slug, parent.ID)

	} else {
		fmt.Printf("FORK WAS FOUND! %v/%v", fork.Project.Key, fork.Name)
	}

	return s.copyRepoAsFork(targetRepo, fork, forkNamespace)
}

// GetNamespaceFork returns a repo pointing to a fork of the given repo in
// the given namespace, ensuring that the fork exists and is a fork of the
// target repo.
func (s BitbucketServerSource) GetNamespaceFork(ctx context.Context, targetRepo *types.Repo, namespace string) (*types.Repo, error) {
	parent := targetRepo.Metadata.(*bitbucketserver.Repo)

	// See if we already have a fork.
	fork, err := s.getFork(ctx, parent, namespace)
	if err != nil {
		return nil, errors.Wrapf(err, "getting fork in %q", namespace)
	}

	// If not, then we need to create a fork.
	if fork == nil {
		// fork, err = s.client.Fork(ctx, parent.Project.Key, parent.Slug, bitbucketserver.CreateForkInput{
		createdFork := parent.Project.Key + "-" + parent.Slug
		fork, err = s.client.Fork(ctx, parent.Project.Key, parent.Slug, bitbucketserver.CreateForkInput{Name: &createdFork,
			Project: &bitbucketserver.CreateForkInputProject{Key: namespace},
		})
		if err != nil {
			return nil, errors.Wrapf(err, "creating fork in %q", namespace)

		}
		fmt.Printf("GetNamepaceFork - Created FORK Name %v", createdFork)
	}

	return s.copyRepoAsFork(targetRepo, fork, namespace)
}

func (s BitbucketServerSource) copyRepoAsFork(targetRepo *types.Repo, fork *bitbucketserver.Repo, forkNamespace string) (*types.Repo, error) {
	targetMeta := targetRepo.Metadata.(*bitbucketserver.Repo)

	targetNameAndNamespace := targetMeta.Project.Key + "/" + targetMeta.Slug
	forkNameAndNamespace := forkNamespace + "/" + targetMeta.Slug

	// Now we make a copy of the target repo, but with its sources and metadata updated to
	// point to the fork
	forkRepo, err := CopyRepoAsFork(targetRepo, fork, targetNameAndNamespace, forkNameAndNamespace)
	if err != nil {
		return nil, errors.Wrap(err, "updating target repo sources")
	}

	return forkRepo, nil
}

var (
	errNotAFork            = errors.New("repo is not a fork")
	errNotForkedFromParent = errors.New("repo was not forked from the given parent")
)

func (s BitbucketServerSource) getFork(ctx context.Context, parent *bitbucketserver.Repo, namespace string) (*bitbucketserver.Repo, error) {
	repo, err := s.client.Repo(ctx, namespace, parent.Project.Key+"-"+parent.Slug)
	// repo, err := s.client.Repo(ctx, namespace, parent.Slug)

	if err != nil {
		if bitbucketserver.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	fmt.Printf("Origin ID: %v, parent ID: %v", repo.Origin.ID, parent.ID)

	// Sanity check: is the returned repo _actually_ a fork of the original?
	if repo.Origin == nil {
		fmt.Printf("DOES NOT HAVE ORIGIN")
		return nil, errNotAFork
	} else if repo.Origin.ID != parent.ID {
		fmt.Printf("WRONG ORIGIN")
		return nil, errNotForkedFromParent
	}

	return repo, nil
}
