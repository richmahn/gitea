package file_handling

import (
	"code.gitea.io/git"
	"code.gitea.io/gitea/models"
	"code.gitea.io/sdk/gitea"
	"net/url"
	"strings"
	"time"
)

func GetFileResponseFromCommit(repo *models.Repository, commit *git.Commit, branch, treeName string) (*gitea.FileResponse, error) {
	fileContents, _ := GetFileContents(repo, treeName, branch)   // ok if fails, then will be nil
	fileCommitResponse, _ := GetFileCommitResponse(repo, commit) // ok if fails, then will be nil
	verification := GetPayloadCommitVerification(commit)
	fileResponse := &gitea.FileResponse{
		Content:      fileContents,
		Commit:       fileCommitResponse,
		Verification: verification,
	}
	return fileResponse, nil
}

func GetFileCommitResponse(repo *models.Repository, commit *git.Commit) (*gitea.FileCommitResponse, error) {
	commitURL, _ := url.Parse(repo.APIURL() + "/git/commits/" + commit.ID.String())
	commitTreeURL, _ := url.Parse(repo.APIURL() + "/git/trees/" + commit.Tree.ID.String())
	parents := make([]gitea.CommitMeta, commit.ParentCount())
	for i := 0; i <= commit.ParentCount(); i++ {
		if parent, err := commit.Parent(i); err == nil && parent != nil {
			parentCommitURL, _ := url.Parse(repo.APIURL() + "/git/commits/" + parent.ID.String())
			parents[i] = gitea.CommitMeta{
				SHA: parent.ID.String(),
				URL: parentCommitURL.String(),
			}
		}
	}
	commitHtmlURL, _ := url.Parse(repo.HTMLURL() + "/commit/" + commit.ID.String())
	fileCommit := &gitea.FileCommitResponse{
		CommitMeta: &gitea.CommitMeta{
			SHA: commit.ID.String(),
			URL: commitURL.String(),
		},
		HTMLURL: commitHtmlURL.String(),
		Author: &gitea.CommitUser{
			Date:  commit.Author.When.UTC().Format(time.RFC3339),
			Name:  commit.Author.Name,
			Email: commit.Author.Email,
		},
		Committer: &gitea.CommitUser{
			Date:  commit.Committer.When.UTC().Format(time.RFC3339),
			Name:  commit.Committer.Name,
			Email: commit.Committer.Email,
		},
		Message: commit.Message(),
		Tree: &gitea.CommitMeta{
			URL: commitTreeURL.String(),
			SHA: commit.Tree.ID.String(),
		},
		Parents: &parents,
	}
	return fileCommit, nil
}

// Gets the author and committer user objects from the IdentityOptions
func GetAuthorAndCommitterUsers(author, committer *IdentityOptions, doer *models.User) (committerUser, authorUser *models.User) {
	// Committer and author are optional. If they are not the doer (not same email address)
	// then we use bogus User objects for them to store their FullName and Email.
	// If only one of the two are provided, we set both of them to it.
	// If neither are provided, both are the doer.
	if committer != nil && committer.Email != "" {
		if doer != nil && strings.ToLower(doer.Email) == strings.ToLower(committer.Email) {
			committerUser = doer // the committer is the doer, so will use their user object
			if committer.Name != "" {
				committerUser.FullName = committer.Name
			}
		} else {
			committerUser = &models.User{
				FullName: committer.Name,
				Email: committer.Email,
			}
		}
	}
	if author != nil && author.Email != "" {
		if doer != nil && strings.ToLower(doer.Email) == strings.ToLower(author.Email) {
			authorUser = doer // the author is the doer, so will use their user object
			if authorUser.Name != "" {
				authorUser.FullName = author.Name
			}
		} else {
			authorUser = &models.User{
				FullName: author.Name,
				Email: author.Email,
			}
		}
	}
	if authorUser == nil {
		if committerUser != nil {
			authorUser = committerUser // No valid author was given so use the committer
		} else if doer != nil {
			authorUser = doer // No valid author was given and no valid committer so use the doer
		}
	}
	if committerUser == nil {
		committerUser = authorUser // No valid committer so use the author as the committer (was set to a valid user above)
	}
	return authorUser, committerUser
}