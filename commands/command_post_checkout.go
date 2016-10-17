package commands

import (
	"os"

	"github.com/github/git-lfs/git"

	"github.com/github/git-lfs/locking"

	"github.com/spf13/cobra"
)

// postCheckoutCommand is run through Git's post-checkout hook. The hook passes
// up to 3 arguments on the command line:
//
//   1. SHA of previous commit before the checkout
//   2. SHA of commit just checked out
//   3. Flag ("0" or "1") - 1 if a branch/tag/SHA was checked out, 0 if a file was
//      In the case of a file being checked out, the pre/post SHA are the same
//
// This hook checks that files which are lockable and not locked are made read-only,
// optimising that as best it can based on the available information.
func postCheckoutCommand(cmd *cobra.Command, args []string) {
	if len(args) != 3 {
		Print("This should be run through Git's post-commit hook.  Run `git lfs update` to install it.")
		os.Exit(1)
	}

	requireGitVersion()

	// Skip this hook if no lockable patterns have been configured
	if len(locking.GetLockablePatterns()) == 0 {
		os.Exit(0)
	}

	if args[2] == "1" {
		postCheckoutRevChange(args[0], args[1])
	} else {
		postCheckoutFileChange()
	}

}

func postCheckoutRevChange(pre, post string) {
	// We can speed things up by looking at the difference between previous HEAD
	// and current HEAD, and only checking lockable files that are different
	files, err := git.GetFilesChanged(pre, post)

	if err != nil {
		LoggedError(err, "Warning: post-checkout rev diff %v:%v failed: %v\nFalling back on full scan.", pre, post, err)
		postCheckoutFileChange()
	}
	err = locking.FixLockableFileWriteFlags(files)
	if err != nil {
		LoggedError(err, "Warning: post-checkout locked file check failed: %v", err)
	}
}

func postCheckoutFileChange() {
	// Sadly we don't get any information about what files were checked out,
	// so we have to check the entire repo
	err := locking.FixAllLockableFileWriteFlags()
	if err != nil {
		LoggedError(err, "Warning: post-checkout locked file check failed: %v", err)
	}
}

func init() {
	RegisterCommand("post-checkout", postCheckoutCommand, func(cmd *cobra.Command) {
	})
}