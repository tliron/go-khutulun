package plugin

type Container struct {
	Name            string
	Reference       string
	CreateArguments []string
	Ports           []Port
}

type Port struct {
	External int64
	Internal int64
	Protocol string
}
