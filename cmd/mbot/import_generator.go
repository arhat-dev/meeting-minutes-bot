package main

// data generation drivers
import (
	_ "arhat.dev/mbot/pkg/generator/archiver"
	_ "arhat.dev/mbot/pkg/generator/chain"
	_ "arhat.dev/mbot/pkg/generator/cron"
	_ "arhat.dev/mbot/pkg/generator/exec"
	_ "arhat.dev/mbot/pkg/generator/filter"
	_ "arhat.dev/mbot/pkg/generator/gotemplate"
	_ "arhat.dev/mbot/pkg/generator/multigen"
)
