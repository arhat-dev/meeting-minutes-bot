/*
Copyright 2020 The arhat.dev Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"arhat.dev/meeting-minutes-bot/pkg/cmd"
	"arhat.dev/meeting-minutes-bot/pkg/version"

	// storage drivers
	_ "arhat.dev/meeting-minutes-bot/pkg/storage/s3"

	// post generation drivers
	_ "arhat.dev/meeting-minutes-bot/pkg/generator/gotemplate"

	// post publishing drivers
	_ "arhat.dev/meeting-minutes-bot/pkg/publisher/telegraph"

	// web archive drivers
	_ "arhat.dev/meeting-minutes-bot/pkg/webarchiver/cdp"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	rootCmd := cmd.NewRootCmd()
	rootCmd.AddCommand(version.NewVersionCmd())

	err := rootCmd.Execute()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to run meeting-minutes-bot %v: %v\n", os.Args, err)
		os.Exit(1)
	}
}
