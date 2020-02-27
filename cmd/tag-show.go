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
	"encoding/xml"
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/minio/cli"
	tagging "github.com/minio/minio/pkg/bucket/object/tagging"
	"github.com/minio/minio/pkg/console"
)

var tagShowCmd = cli.Command{
	Name:   "show",
	Usage:  "show tags for objects",
	Action: mainShowTag,
	Before: setGlobalsFromContext,
	Flags:  globalFlags,
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
  1. Show the tags assigned to the object named testobject in the bucket testbucket on alias s3.
     {{.Prompt}} {{.HelpName}} s3/testbucket/testobject

`,
}

func getHostCfgFromURL(urlStr string) *hostConfigV9 {
	clientURL := newClientURL(urlStr)
	alias := splitStr(clientURL.String(), "/", 3)[0]
	return mustGetHostConfig(alias)
}

func showObjectName(urlStr string) {
	s3c := getS3ClientForTagOps(urlStr)
	bucket, object := s3c.url2BucketAndObject()
	alias, _ := url2Alias(urlStr)
	hostCfg := getHostCfgFromURL(urlStr)
	hostHeader := fmt.Sprintf("%-10s: %s", "Host", hostCfg.URL)
	console.Println(console.Colorize(fieldThemeHeader, hostHeader))
	aliasHeader := fmt.Sprintf("%-10s: %s", "Alias", alias)
	console.Println(console.Colorize(fieldThemeHeader, aliasHeader))
	objectName := fmt.Sprintf("%-10s: %s", "Object", bucket+slashSeperator+object)
	console.Println(console.Colorize(fieldThemeHeader, objectName))
}

func getTagList(urlStr string) []tagging.Tag {
	var err error

	s3c := getS3ClientForTagOps(urlStr)
	bucketName, objectName := s3c.url2BucketAndObject()
	tagXML, err := s3c.api.GetObjectTagging(bucketName, objectName)
	if err != nil {
		console.Errorln(err.Error() + " Unable to get Object Tags.")
		return nil
	}
	var tagObj tagging.Tagging
	if err = xml.Unmarshal([]byte(tagXML), &tagObj); err != nil {
		console.Errorln(err.Error() + ", Unable to set Object Tags for display.")
	}

	return tagObj.TagSet.Tags
}

// Color scheme for tag display
func setTagShowColorScheme() {
	console.SetColor(fieldMainHeader, color.New(color.Bold, color.FgHiRed))
	console.SetColor(fieldThemeRow, color.New(color.FgWhite))
	console.SetColor(fieldThemeHeader, color.New(color.Bold, color.FgCyan))
}

func checkShowTagSyntax(ctx *cli.Context) {
	if len(ctx.Args()) != 1 {
		cli.ShowCommandHelp(ctx, "show")
		os.Exit(globalErrorExitStatus)
	}
}

func showInfoFieldOne(label string, field string) {
	displayField := fmt.Sprintf("   %-16s: %s ", label, field)
	console.Println(console.Colorize(fieldThemeRow, displayField))
}

func showInfoFieldMultiple(kvpairs []tagging.Tag) {
	for idx := 0; idx < len(kvpairs); idx++ {
		displayField := fmt.Sprintf("   %-16s:  %s ", kvpairs[idx].Key, kvpairs[idx].Value)
		console.Println(console.Colorize(fieldThemeRow, displayField))
	}
}

func mainShowTag(ctx *cli.Context) {
	checkShowTagSyntax(ctx)
	setTagShowColorScheme()
	args := ctx.Args()
	objectURL := args.Get(0)
	allTagStr := getTagList(objectURL)

	switch len(allTagStr) {
	case 0:
		console.Infoln("No tags set.")
	case 1:
		showObjectName(objectURL)
		showInfoFieldOne(allTagStr[0].Key, allTagStr[0].Value)
	default:
		showObjectName(objectURL)
		showInfoFieldMultiple(allTagStr)
	}
}
