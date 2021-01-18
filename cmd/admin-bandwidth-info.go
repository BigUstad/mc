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
	"context"
	"errors"
	"fmt"

	"github.com/fatih/color"
	"github.com/minio/cli"
	json "github.com/minio/mc/pkg/colorjson"
	"github.com/minio/mc/pkg/probe"
	"github.com/minio/minio/pkg/bandwidth"
	"github.com/minio/minio/pkg/console"
)

var adminBwInfoCmd = cli.Command{
	Name:   "bandwidth",
	Usage:  "Show bandwidth info for buckets on the MinIO server",
	Action: mainAdminBwInfo,
	Before: setGlobalsFromContext,
	Flags:  append(adminBwFlags, globalFlags...),
	CustomHelpTemplate: `NAME:
  {{.HelpName}} - {{.Usage}}

USAGE:
  {{.HelpName}} TARGET

FLAGS:
  {{range .VisibleFlags}}{{.}}
  {{end}}
EXAMPLES:
  1. Show the bandwidth usage for the MinIO server
     {{.Prompt}} {{.HelpName}} --bucket bucket1 --bucket bucket2 play/
  2. Show the bandwidth usage for a distributed MinIO servers setup
     {{.Prompt}} {{.HelpName}} --bucket bucket1 --bucket bucket2 --server minio01/ --server minio02/ --server minio03/ --server minio04/
`,
}

var adminBwFlags = []cli.Flag{
	cli.StringSliceFlag{
		Name:  "bucket",
		Usage: "format '--bucket <bucket>', multiple values allowed for multiple buckets",
	},
	cli.StringSliceFlag{
		Name:  "servers",
		Usage: "format '--server <server>', multiple values allowed for multiple servers for a distributed setup",
	},
}

// Wrap single server "Info" message together with fields "Status" and "Error"
type bandwidthSingleStruct struct {
	Status string           `json:"status"`
	Error  string           `json:"error,omitempty"`
	Server string           `json:"server,omitempty"`
	Result bandwidth.Report `json:"result,omitempty"`
}

func (b bandwidthSingleStruct) String() (msg string) {
	if b.Status == "error" {
		fatal(probe.NewError(errors.New(b.Error)), "Unable to get service status")
	}
	// Color palette initialization
	console.SetColor("Info", color.New(color.FgGreen, color.Bold))
	console.SetColor("InfoHeader", color.New(color.FgRed, color.Bold))
	console.SetColor("InfoFail", color.New(color.FgRed, color.Bold))
	console.SetColor("BlinkLoad", color.New(color.BlinkSlow, color.FgCyan))
	msg += fmt.Sprintf("%s  %s\n", console.Colorize("Info", dot), console.Colorize("PrintB", b.Server))
	for bucket, details := range b.Result.BucketStats {
		avgAg := fmt.Sprintf("%.4f", details.CurrentBandwidthInBytesPerSecond)
		avgMv := fmt.Sprintf("%d", details.LimitInBytesPerSecond)
		msg += fmt.Sprintf("   Bucket: %s\n", console.Colorize("Info", bucket))
		msg += fmt.Sprintf("      Current Bandwidth  : %s\n", console.Colorize("Info", avgAg))
		msg += fmt.Sprintf("      Limit Bandwidth     : %s\n", console.Colorize("Info", avgMv))
	}
	return msg
}

func (b bandwidthSingleStruct) JSON() string {
	statusJSONBytes, e := json.MarshalIndent(b, "", "    ")
	fatalIf(probe.NewError(e), "Unable to marshal into JSON.")

	return string(statusJSONBytes)
}

// Returns a collection fetched for all the buckets
// from each of the server

func fetchInitialBandwidthData(server string, buckets []string) map[string]bandwidth.Details {
	ctx, _ := context.WithCancel(globalContext)
	bwSample := make(map[string]bandwidth.Details)
	client, err := newAdminClient(server)
	bwCh := client.GetBucketBandwidth(ctx, buckets...)
	fatalIf(err, "Unable to initialize admin connection with "+server)
	for i := 0; i < 10; i++ {
		select {
		case sample, ok := <-bwCh:
			if !ok {
				return bwSample
			}
			if i == 0 {
				for bucket, details := range sample.Report.BucketStats {
					bwSample[bucket] = details
				}
				continue
			}
			for bucket, details := range sample.Report.BucketStats {
				if val, ok := bwSample[bucket]; ok {
					cur := (val.CurrentBandwidthInBytesPerSecond + details.CurrentBandwidthInBytesPerSecond) / 2.0
					limit := (val.LimitInBytesPerSecond + details.LimitInBytesPerSecond) / 2
					bwSample[bucket] = bandwidth.Details{
						limit,
						cur,
					}
				}
			}
		}
	}
	return bwSample
}

func printTable(bwDistributedSample map[string]bandwidth.Details) {
	dspOrder := []col{colRed, colGrey}
	printColors := []*color.Color{}
	for _, c := range dspOrder {
		printColors = append(printColors, getPrintCol(c))
	}
	for bucket, values := range bwDistributedSample {
		t := console.NewTable(printColors, []bool{false, false, false}, 1)
		cellText := make([][]string, 2)
		cellText[0] = []string{
			fmt.Sprintf("Bucket"),
			fmt.Sprintf("Avg Current Bandwidth"),
			fmt.Sprintf("Avg Limit Bandwidth"),
		}
		cellText[1] = []string{
			fmt.Sprintf(console.Colorize("InfoHeader", bucket)),
			fmt.Sprintf("%.4f", values.CurrentBandwidthInBytesPerSecond),
			fmt.Sprintf("%d", values.LimitInBytesPerSecond),
		}
		t.DisplayTable(cellText)
	}
}

func checkAdminBwInfoSyntax(ctx *cli.Context) {
	if len(ctx.Args()) > 1 {
		cli.ShowCommandHelpAndExit(ctx, "bandwidth", globalErrorExitStatus)
	}
}

func mainAdminBwInfo(ctx *cli.Context) error {
	checkAdminBwInfoSyntax(ctx)
	// Set color preference of command outputs
	console.SetColor("ConfigHeading", color.New(color.Bold, color.FgHiRed))
	console.SetColor("ConfigFG", color.New(color.FgHiWhite))

	var buckets []string
	args := ctx.Args()
	console.PrintC(console.Colorize("BlinkLoad", "COLLECTING SAMPLES & CALCULATING SUM & AVERAGE. These are the initial values..\n"))
	urlStr := args.Get(0)
	rewindLines := (4 * len(buckets))
	for {
		bwSample := fetchInitialBandwidthData(urlStr, buckets)
		printTable(bwSample)
		console.RewindLines(rewindLines)
	}
}
