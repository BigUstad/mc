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
	"errors"
	"os"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/minio/cli"
	"github.com/minio/minio/pkg/console"
)

var tagSetCmd = cli.Command{
	Name:   "set",
	Usage:  "set/configure tags for objects",
	Action: mainSetTag,
	Before: setGlobalsFromContext,
	Flags:  append(tagSetFlags, globalFlags...),
	CustomHelpTemplate: `Name:
	{{.HelpName}} - {{.Usage}}

USAGE:
  {{.HelpName}} [COMMAND FLAGS] TARGET

FLAGS:
  {{range .VisibleFlags}}{{.}}
  {{end}}
DESCRIPTION:
   Show tags assigned to an object.

EXAMPLES:
  1. Assign the tag values to testobject in the bucket testbucket on alias s3.
     {{.Prompt}} {{.HelpName}} s3/testbucket/testobject --tags "key1:value1" --tags "key2:value2" --tags "key3:value3"

`,
}

var tagSetFlags = []cli.Flag{
	cli.StringSliceFlag{
		Name:  "tags",
		Usage: "format '<key>:<value>'; multiple values allowed for multiple key/value pairs",
	},
}

// Color scheme for set tag results
func setTagSetColorScheme() {
	console.SetColor(fieldThemeSuccess, color.New(color.Bold, color.FgGreen))
	console.SetColor(fieldThemeFailure, color.New(color.Bold, color.FgRed))
}

func checkSetTagSyntax(ctx *cli.Context) {
	if len(ctx.Args()) != 1 {
		cli.ShowCommandHelp(ctx, "set")
		os.Exit(globalErrorExitStatus)
	}
}

func getTagMap(ctx *cli.Context) (map[string]string, error) {
	ilmTagKVMap := make(map[string]string)
	tagValues := ctx.StringSlice(strings.ToLower("tags"))
	for tagIdx, tag := range tagValues {
		key := splitStr(tag, ":", 2)[0]
		val := splitStr(tag, ":", 2)[1]
		if key != "" && val != "" {
			ilmTagKVMap[key] = val
		} else {
			return nil, errors.New("error extracting tag argument(#" + strconv.Itoa(tagIdx+1) + ") " + tag)
		}
	}
	return ilmTagKVMap, nil
}

func mainSetTag(ctx *cli.Context) error {
	checkSetTagSyntax(ctx)
	setTagSetColorScheme()

	objectURL := ctx.Args().Get(0)
	var err error
	var objTagMap map[string]string
	if objTagMap, err = getTagMap(ctx); err != nil {
		console.Errorln(err.Error() + ". Unable to get tags to set.")
		return err
	}
	alias, _ := url2Alias(objectURL)
	if alias == "" {
		fatalIf(errInvalidAliasedURL(objectURL), "Unable to set tags to target "+objectURL)
	}
	s3c := getS3ClientForTagOps(objectURL)
	bucket, object := s3c.url2BucketAndObject()
	if err = s3c.api.PutObjectTagging(bucket, object, objTagMap); err != nil {
		console.Errorln(err.Error() + ". Unable to set tags.")
		return err
	}
	console.Infoln("Tag set for `" + bucket + slashSeperator + object + "`.")
	return nil
}
