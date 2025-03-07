// Copyright 2020 zhengyansheng
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package flag

import (
	"os"
)

var defaultFlags = []Flag{
	// HelpFlag prints usage of application.
	&BoolFlag{
		Name:  "help",
		Usage: "--help, show help information",
		Action: func(name string, fs *FlagSet) {
			fs.PrintDefaults()
			os.Exit(0)
		},
	},
}
