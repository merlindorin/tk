package powerpacks

// coreOwner is the owner recorded for the migrations tk carries itself, as opposed to
// the ones a powerpack declares in its manifest.
const coreOwner = "tk"

// coreMigrations are the migrations tk owns: they undo what the engine itself used to
// write, which no powerpack can be held responsible for. They are expressed with the
// same instructions a powerpack manifest uses, so there is one mechanism, not two.
func coreMigrations() []migration {
	return []migration{
		{
			owner: coreOwner,
			Migration: Migration{
				ID:     "drop-remote-taskfiles-export",
				Reason: `Task released the REMOTE_TASKFILES experiment and warns about the variable on every run`,
				RemoveLines: &RemoveLines{
					Path:            EnvrcFilename,
					Matching:        `^\s*export\s+TASK_X_REMOTE_TASKFILES=`,
					DeleteWhenEmpty: true,
				},
				RemoveFile: "",
				RenameFile: nil,
			},
		},
	}
}
