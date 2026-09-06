package powerpacks

// Marker flags every line tk owns in a file it shares with the user (the root
// Taskfile.yaml includes and .envrc). Lines without it are never touched.
const Marker = "#!tk"

const (
	// legacyEnvrcExport is the variable tk used to export from .envrc to turn on the Task
	// remote taskfiles experiment. Task released the experiment and now warns about the
	// variable on every single invocation, so tk removes the line instead of writing it.
	legacyEnvrcExport = "TASK_X_REMOTE_TASKFILES"

	taskfileVersion = "version: '3'"
	taskfileHeader  = "includes:"
	taskfileFooter  = `dotenv:
  - .env
  - PROJECT
  - .env.default`
)
