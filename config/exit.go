package config

// Exit codes signal why the application exited.
const (
	ExitSuccess           = 0 // This is actually never used.
	ExitErrorConfigAccess = 1 // Error accessing the config file.
	ExitErrorConfigRead   = 2 // Error reading the config file.
	ExitErrorConfigParse  = 3 // Error parsing the config file.
	ExitPrivateKeyCorrupt = 4 // Private key is corrupt.
	ExitPrivateKeyCreate  = 5 // Cannot create a new private key.
	ExitBlockchainCorrupt = 6 // Blockchain is corrupt.
)
