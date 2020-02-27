/*
 * MinIO Client (C) 2020 MinIO, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cmd

import (
	"github.com/minio/cli"
)

var tagCmd = cli.Command{
	Name:   "tag",
	Usage:  "configure tags for objects",
	Action: mainTag,
	Before: setGlobalsFromContext,
	Flags:  globalFlags,
	Subcommands: []cli.Command{
		//tagShowCmd,
		tagSetCmd,
	},
}

const (
	fieldMainHeader   string = "Main-Heading"
	fieldThemeHeader  string = "Row-Header"
	fieldThemeRow     string = "Row-Normal"
	fieldThemeSuccess string = "Result-Success"
	fieldThemeFailure string = "Result-Failure"
)

func checkMainTagSyntax(ctx *cli.Context) {
	cli.ShowCommandHelp(ctx, "")
}

// getS3Client - To make api calls.
func getS3ClientForTagOps(urlStr string) s3Client {
	client, err := newClient(urlStr)
	if err != nil {
		fatalIf(err.Trace(urlStr), "Cannot parse the provided url.")
	}

	s3Client, ok := client.(*s3Client)
	if !ok {
		fatalIf(err.Trace(urlStr), "The provided url doesn't point to a S3 server.")
	}
	return *s3Client
}

func mainTag(ctx *cli.Context) error {
	checkMainTagSyntax(ctx)

	return nil
}
