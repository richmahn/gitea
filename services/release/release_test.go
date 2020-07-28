// Copyright 2018 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package release

import (
	"path/filepath"
	"testing"
	"time"

	"code.gitea.io/gitea/models"
	"code.gitea.io/gitea/modules/git"

	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	models.MainTest(m, filepath.Join("..", ".."))
}

func TestRelease_Create(t *testing.T) {
	assert.NoError(t, models.PrepareTestDatabase())

	user := models.AssertExistsAndLoadBean(t, &models.User{ID: 2}).(*models.User)
	repo := models.AssertExistsAndLoadBean(t, &models.Repository{ID: 1}).(*models.Repository)
	repoPath := models.RepoPath(user.Name, repo.Name)

	gitRepo, err := git.OpenRepository(repoPath)
	assert.NoError(t, err)
	defer gitRepo.Close()

	assert.NoError(t, CreateRelease(gitRepo, &models.Release{
		RepoID:       repo.ID,
		PublisherID:  user.ID,
		TagName:      "v0.1",
		Target:       "master",
		Title:        "v0.1 is released",
		Note:         "v0.1 is released",
		IsDraft:      false,
		IsPrerelease: false,
		IsTag:        false,
	}, nil))

	assert.NoError(t, CreateRelease(gitRepo, &models.Release{
		RepoID:       repo.ID,
		PublisherID:  user.ID,
		TagName:      "v0.1.1",
		Target:       "65f1bf27bc3bf70f64657658635e66094edbcb4d",
		Title:        "v0.1.1 is released",
		Note:         "v0.1.1 is released",
		IsDraft:      false,
		IsPrerelease: false,
		IsTag:        false,
	}, nil))

	assert.NoError(t, CreateRelease(gitRepo, &models.Release{
		RepoID:       repo.ID,
		PublisherID:  user.ID,
		TagName:      "v0.1.2",
		Target:       "65f1bf2",
		Title:        "v0.1.2 is released",
		Note:         "v0.1.2 is released",
		IsDraft:      false,
		IsPrerelease: false,
		IsTag:        false,
	}, nil))

	assert.NoError(t, CreateRelease(gitRepo, &models.Release{
		RepoID:       repo.ID,
		PublisherID:  user.ID,
		TagName:      "v0.1.3",
		Target:       "65f1bf2",
		Title:        "v0.1.3 is released",
		Note:         "v0.1.3 is released",
		IsDraft:      true,
		IsPrerelease: false,
		IsTag:        false,
	}, nil))

	assert.NoError(t, CreateRelease(gitRepo, &models.Release{
		RepoID:       repo.ID,
		PublisherID:  user.ID,
		TagName:      "v0.1.4",
		Target:       "65f1bf2",
		Title:        "v0.1.4 is released",
		Note:         "v0.1.4 is released",
		IsDraft:      false,
		IsPrerelease: true,
		IsTag:        false,
	}, nil))

	assert.NoError(t, CreateRelease(gitRepo, &models.Release{
		RepoID:       repo.ID,
		PublisherID:  user.ID,
		TagName:      "v0.1.5",
		Target:       "65f1bf2",
		Title:        "v0.1.5 is released",
		Note:         "v0.1.5 is released",
		IsDraft:      false,
		IsPrerelease: false,
		IsTag:        true,
	}, nil))
}

func TestRelease_Update(t *testing.T) {
	assert.NoError(t, models.PrepareTestDatabase())

	user := models.AssertExistsAndLoadBean(t, &models.User{ID: 2}).(*models.User)
	repo := models.AssertExistsAndLoadBean(t, &models.Repository{ID: 1}).(*models.Repository)
	repoPath := models.RepoPath(user.Name, repo.Name)

	gitRepo, err := git.OpenRepository(repoPath)
	assert.NoError(t, err)
	defer gitRepo.Close()

	// Create all releases used in the update calls below
	assert.NoError(t, CreateRelease(gitRepo, &models.Release{
		RepoID:       repo.ID,
		PublisherID:  user.ID,
		TagName:      "v1.1.1",
		Target:       "master",
		Title:        "v1.1.1 is released",
		Note:         "v1.1.1 is released",
		IsDraft:      false,
		IsPrerelease: false,
		IsTag:        false,
	}, nil))
	assert.NoError(t, CreateRelease(gitRepo, &models.Release{
		RepoID:       repo.ID,
		PublisherID:  user.ID,
		TagName:      "v1.2.1",
		Target:       "master",
		Title:        "v1.2.1 is draft",
		Note:         "v1.2.1 is draft",
		IsDraft:      true,
		IsPrerelease: false,
		IsTag:        false,
	}, nil))
	assert.NoError(t, CreateRelease(gitRepo, &models.Release{
		RepoID:       repo.ID,
		PublisherID:  user.ID,
		TagName:      "v1.3.1",
		Target:       "master",
		Title:        "v1.3.1 is pre-released",
		Note:         "v1.3.1 is pre-released",
		IsDraft:      false,
		IsPrerelease: true,
		IsTag:        false,
	}, nil))
	assert.NoError(t, CreateRelease(gitRepo, &models.Release{
		RepoID:       repo.ID,
		PublisherID:  user.ID,
		TagName:      "v1.4.1",
		Target:       "master",
		Title:        "v1.4.1 is another draft",
		Note:         "v1.4.1 is another draft",
		IsDraft:      true,
		IsPrerelease: false,
		IsTag:        false,
	}, nil))

	time.Sleep(time.Second) // sleep 1 second to ensure a different timestamp
	
	release, err := models.GetRelease(repo.ID, "v1.1.1")
	assert.NoError(t, err)
	releaseCreatedUnix := release.CreatedUnix
	release.Note = "Changed note"
	assert.NoError(t, UpdateRelease(user, gitRepo, release, nil))
	release, err = models.GetReleaseByID(release.ID)
	assert.NoError(t, err)
	assert.Equal(t, int64(releaseCreatedUnix), int64(release.CreatedUnix))

	// Test change of a release that doesn't have a tag (is a draft)
	release, err = models.GetRelease(repo.ID, "v1.2.1")
	assert.NoError(t, err)
	releaseCreatedUnix = release.CreatedUnix
	release.Title = "Changed title"
	assert.NoError(t, UpdateRelease(user, gitRepo, release, nil))
	release, err = models.GetReleaseByID(release.ID)
	assert.NoError(t, err)
	assert.Less(t, int64(releaseCreatedUnix), int64(release.CreatedUnix))

	// Test a changed pre-release
	release, err = models.GetRelease(repo.ID, "v1.3.1")
	assert.NoError(t, err)
	releaseCreatedUnix = release.CreatedUnix
	release.Title = "Changed title"
	release.Note = "Changed note"
	assert.NoError(t, UpdateRelease(user, gitRepo, release, nil))
	release, err = models.GetReleaseByID(release.ID)
	assert.NoError(t, err)
	assert.Equal(t, int64(releaseCreatedUnix), int64(release.CreatedUnix))

	// Test a change from draft to tagged release
	release, err = models.GetRelease(repo.ID, "v1.4.1")
	assert.NoError(t, err)
	releaseCreatedUnix = release.CreatedUnix
	release.IsDraft = false
	release.TagName = "newTagName"
	assert.NoError(t, UpdateRelease(user, gitRepo, release, nil))
	release, err = models.GetReleaseByID(release.ID)
	assert.NoError(t, err)
	assert.Less(t, int64(releaseCreatedUnix), int64(release.CreatedUnix))
	assert.Equal(t, "newtagname", release.LowerTagName)
}

func TestRelease_createTag(t *testing.T) {
	assert.NoError(t, models.PrepareTestDatabase())

	user := models.AssertExistsAndLoadBean(t, &models.User{ID: 2}).(*models.User)
	repo := models.AssertExistsAndLoadBean(t, &models.Repository{ID: 1}).(*models.Repository)
	repoPath := models.RepoPath(user.Name, repo.Name)

	gitRepo, err := git.OpenRepository(repoPath)
	assert.NoError(t, err)
	defer gitRepo.Close()

	release := &models.Release{
		RepoID:       repo.ID,
		PublisherID:  user.ID,
		TagName:      "v2.1.1",
		Target:       "master",
		Title:        "v2.1.1 is released",
		Note:         "v2.1.1 is released",
		IsDraft:      false,
		IsPrerelease: false,
		IsTag:        false,
	}
	assert.NoError(t, createTag(gitRepo, release))
	assert.NotEmpty(t, release.CreatedUnix)

	releaseDraft := &models.Release{
		RepoID:       repo.ID,
		PublisherID:  user.ID,
		TagName:      "v2.2.1",
		Target:       "65f1bf2",
		Title:        "v2.2.1 is draft",
		Note:         "v2.2.1 is draft",
		IsDraft:      true,
		IsPrerelease: false,
		IsTag:        false,
	}
	assert.NoError(t, createTag(gitRepo, releaseDraft))
	assert.NotEmpty(t, releaseDraft.CreatedUnix)

	releasePre := &models.Release{
		RepoID:       repo.ID,
		PublisherID:  user.ID,
		TagName:      "v2.3.1",
		Target:       "65f1bf2",
		Title:        "v2.3.1 is pre-released",
		Note:         "v2.3.1 is pre-released",
		IsDraft:      false,
		IsPrerelease: true,
		IsTag:        false,
	}
	assert.NoError(t, createTag(gitRepo, releasePre))
	assert.NotEmpty(t, releasePre.CreatedUnix)

	releaseDraft2 := &models.Release{
		RepoID:       repo.ID,
		PublisherID:  user.ID,
		TagName:      "v2.3.1",
		Target:       "65f1bf2",
		Title:        "v2.4.1 is another draft",
		Note:         "v2.4.1 is another draft",
		IsDraft:      true,
		IsPrerelease: false,
		IsTag:        false,
	}
	assert.NoError(t, createTag(gitRepo, releaseDraft2))
	assert.NotEmpty(t, releaseDraft2.CreatedUnix)

	time.Sleep(time.Second) // sleep 1 second to ensure a different timestamp

	// Test a changed release
	releaseCreatedUnix := release.CreatedUnix
	release.Note = "Changed note"
	assert.NoError(t, createTag(gitRepo, release))
	assert.Equal(t, int64(releaseCreatedUnix), int64(release.CreatedUnix))

	// Test a changed draft
	releaseCreatedUnix = releaseDraft.CreatedUnix
	releaseDraft.Title = "Changed title"
	assert.NoError(t, createTag(gitRepo, releaseDraft))
	assert.Less(t, int64(releaseCreatedUnix), int64(releaseDraft.CreatedUnix))

	// Test a changed pre-release
	releaseCreatedUnix = releasePre.CreatedUnix
	release.Title = "Changed title"
	release.Note = "Changed note"
	assert.NoError(t, createTag(gitRepo, releasePre))
	assert.Equal(t, int64(releaseCreatedUnix), int64(releasePre.CreatedUnix))

	// Test a change from draft to tagged release
	releaseCreatedUnix = releaseDraft2.CreatedUnix
	releaseDraft2.IsDraft = false
	assert.NoError(t, createTag(gitRepo, releaseDraft2))
	assert.Less(t, int64(releaseCreatedUnix), int64(releaseDraft2.CreatedUnix))
}
