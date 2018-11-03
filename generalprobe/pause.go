package generalprobe

import "time"

type pauseScene struct {
	interval uint
	baseScene
}

func Pause(interval uint) *pauseScene {
	scene := pauseScene{
		interval: interval,
	}
	return &scene
}

func (x *pauseScene) play() error {
	time.Sleep(time.Second * time.Duration(x.interval))
	return nil
}
