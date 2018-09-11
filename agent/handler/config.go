package handler

type Config struct {
	ScratchRoot string `env:"scratch_root" default:"/tmp"`
}
