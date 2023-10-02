package main

import (
	"log"
	"os"
	"time"

	"github.com/vinegarhq/vinegar/internal/config"
	"github.com/vinegarhq/vinegar/roblox"
	"github.com/vinegarhq/vinegar/wine"
)

type Binary struct {
	name     string
	log      chan string
	progress chan float32

	dir string
	pfx *wine.Prefix

	cfg  *config.Config
	bcfg *config.Application

	btype roblox.BinaryType
	ver   roblox.Version
}

func NewBinary(bt roblox.BinaryType, cfg *config.Config, pfx *wine.Prefix) Binary {
	var bcfg config.Application

	switch bt {
	case roblox.Player:
		bcfg = cfg.Player
	case roblox.Studio:
		bcfg = cfg.Studio
	}

	return Binary{
		name:     bt.String(),
		log:      make(chan string),
		progress: make(chan float32),

		btype: bt,
		pfx:   pfx,
		cfg:   cfg,
		bcfg:  &bcfg,
	}
}

func (b *Binary) Run(args ...string) {
	exitChan := make(chan bool)

	go func() {
		b.Glass(exitChan)
	}()

	fatal := func(err error) {
		log.Println(err)
		b.log <- err.Error()
		os.Exit(1)
	}

	if err := b.Setup(); err != nil {
		fatal(err)
	}

	cmd, err := b.Command(args...)
	if err != nil {
		fatal(err)
	}

	log.Printf("Launching %s", b.name)
	b.log <- "Launching Roblox"

	time.Sleep(time.Second * 2)

	if err := cmd.Start(); err != nil {
		fatal(err)
	}

	log.Println("Sending exit")
	exitChan <- true
	cmd.Wait()

	if b.bcfg.AutoKillPrefix {
		b.pfx.Kill()
	}
}
