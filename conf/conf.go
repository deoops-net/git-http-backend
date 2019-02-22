package conf

type Configuration struct {
	Home string
}

var Default Configuration = Configuration{
	Home: "/tmp/repo",
}
