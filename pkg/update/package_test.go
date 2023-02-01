package update

import (
	"testing"
	"time"

	"github.com/hashicorp/go-version"

	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-billy/v5/util"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/cache"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/filesystem"
	"github.com/stretchr/testify/assert"
)

func TestPackageUpdate_getFixesCVEList(t *testing.T) {
	dir := t.TempDir()
	r := setupTestRepo(t, dir)

	// create a tag 1.2.3
	createTestTag(t, r, "1.2.3")
	// create a commit with CVE which is in a release before our test range and so should not appear in the list of CVEs that we assert
	createTestCommit(t, r, "a1", "fixes: CVE123abc and some other text")

	// create a tag 1.2.4, this is the previous release, commits after this should appear in the list of CVEs we assert
	createTestTag(t, r, "1.2.4")
	createTestCommit(t, r, "a2", "fixes: CVE456abc and some other text")
	createTestCommit(t, r, "a3", "FIXES: cve78910qwerty and some other text")

	// create a new release tag 1.2.5
	createTestTag(t, r, "1.2.5")

	// run secfixes, current version 1.2.5, previous 1.2.4 and assert only two CVEs returned in list
	o := PackageOptions{
		Secfixes: true,
	}

	previousVersion, err := version.NewVersion("1.2.4")
	assert.NoError(t, err)

	cves, err := o.getFixesCVEList(dir, previousVersion)
	assert.NoError(t, err)

	assert.Equal(t, 2, len(cves))
	assert.Equal(t, "CVE78910QWERTY", cves[0])
	assert.Equal(t, "CVE456ABC", cves[1])
}

func setupTestRepo(t *testing.T, dir string) *git.Repository {
	fs := osfs.New(dir)

	storage := filesystem.NewStorage(fs, cache.NewObjectLRUDefault())
	wt, err := fs.Chroot("test")
	assert.NoError(t, err)

	r, err := git.Init(storage, wt)
	assert.NoError(t, err)

	w, err := r.Worktree()
	assert.NoError(t, err)

	err = util.WriteFile(w.Filesystem, "foo.yaml", []byte("ok"), 0o644)
	assert.NoError(t, err)

	_, err = w.Add("foo.yaml")
	assert.NoError(t, err)

	_, err = w.Commit("initial test checkin", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "John Doe",
			Email: "john@doe.org",
			When:  time.Now(),
		},
	})
	assert.NoError(t, err)

	return r
}

func createTestTag(t *testing.T, r *git.Repository, tag string) {
	h, err := r.Head()
	assert.NoError(t, err)

	tagOptions := &git.CreateTagOptions{
		Message: tag,
	}

	tagOptions.Tagger = &object.Signature{
		Name:  "test",
		Email: "test@tester.com",
		When:  time.Now(),
	}

	_, err = r.CreateTag(tag, h.Hash(), tagOptions)
	assert.NoError(t, err)
}

func createTestCommit(t *testing.T, r *git.Repository, testData, commitMessage string) {
	w, err := r.Worktree()
	assert.NoError(t, err)

	err = util.WriteFile(w.Filesystem, "cheese.yaml", []byte(testData), 0o644)
	assert.NoError(t, err)

	_, err = w.Add("cheese.yaml")
	assert.NoError(t, err)

	_, err = w.Commit(commitMessage, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "John Doe",
			Email: "john@doe.org",
			When:  time.Now(),
		},
	})
	assert.NoError(t, err)
}
