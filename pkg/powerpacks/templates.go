package powerpacks

// Marker flags every line tk owns in a file it shares with the user (the root
// Taskfile.yaml includes and .envrc). Lines without it are never touched.
const Marker = "#!tk"

const (
	taskfileVersion = "version: '3'"
	taskfileHeader  = "includes:"
	taskfileFooter  = `dotenv:
  - .env
  - PROJECT
  - .env.default`
)
