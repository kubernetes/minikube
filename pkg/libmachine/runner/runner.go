package runner

type Runner interface {
	Session()
	RunCommand(string) (string, error)
}
