package deploy



type MiniTestDeployer interface {
	// Up should provision the environment for testing
	Up() error
	// Down should tear down the environment if any
	Down() error
	// IsUp should return true if a test cluster is successfully provisioned
	IsUp() ( bool,  error)
	
}


