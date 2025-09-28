package deployer


type MiniTestDeployer interface {
	// Up should provision the environment for testing
	Up() error
	// Down should tear down the environment if any
	Down() error
	// IsUp should return true if a test cluster is successfully provisioned
	IsUp() ( bool,  error)
	// Execute execute a command in the deployed environment
	Execute(args ...string)error
	// Sync copy files from src on host to dst on deployed environment 
	Sync(src string, dst string) error
}

