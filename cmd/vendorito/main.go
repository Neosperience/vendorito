package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
	"vendorito"

	"github.com/containers/common/pkg/retry"
	"github.com/containers/image/v5/copy"
	"github.com/containers/image/v5/signature"
	"github.com/containers/image/v5/types"
	"github.com/urfave/cli/v2"
)

func app(c *cli.Context) error {
	ctx := context.Background()
	timeout := c.Int64("timeout")
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
		defer cancel()
	}

	// Parse credentials, if specified
	credentials := c.String("credentials")
	authStore, err := vendorito.ParseCredentials(strings.Fields(credentials))
	if err != nil {
		return fmt.Errorf("error parsing credentials: %w", err)
	}

	//TODO: switch to signature.DefaultPolicy(nil) maybe?
	policy := &signature.Policy{Default: []signature.PolicyRequirement{signature.NewPRInsecureAcceptAnything()}}

	policyContext, err := signature.NewPolicyContext(policy)
	if err != nil {
		return fmt.Errorf("could not use policy: %w", err)
	}
	defer policyContext.Destroy()

	srcRef, srcInfo, err := vendorito.ParseDockerURL(c.String("source"))
	if err != nil {
		return fmt.Errorf("could not parse source image url: %w", err)
	}
	destRef, destInfo, err := vendorito.ParseDockerURL(c.String("target"))
	if err != nil {
		return fmt.Errorf("could not parse target image url: %w", err)
	}

	// Get auth file if specified
	authFile := c.String("auth-file")

	// If source or target have login info, set them for the context
	srcContext := &types.SystemContext{
		AuthFilePath: authFile,
	}
	err = vendorito.AddAuthToContext(srcContext, authStore, srcInfo, false)
	if err != nil {
		return fmt.Errorf("could not set auth info for source: %w", err)
	}

	destContext := &types.SystemContext{
		AuthFilePath: authFile,
	}
	err = vendorito.AddAuthToContext(destContext, authStore, destInfo, true)
	if err != nil {
		return fmt.Errorf("could not set auth info for target: %w", err)
	}

	runOp := func() error {
		_, err := copy.Image(ctx, policyContext, destRef, srcRef, &copy.Options{
			ReportWriter:    os.Stdout,
			PreserveDigests: true,
			SourceCtx:       srcContext,
			DestinationCtx:  destContext,
		})
		return err
	}

	retryNum := c.Int("retry-max")
	if retryNum > 0 {
		return retry.RetryIfNecessary(ctx, runOp, &retry.RetryOptions{
			MaxRetry: retryNum,
			Delay:    time.Duration(c.Int64("retry-delay")) * time.Second,
		})
	} else {
		return runOp()
	}
}

func main() {
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "source",
				Aliases:  []string{"i"},
				Usage:    "Source image path, including tag (if tag is omitted, 'latest' tag will be used)",
				EnvVars:  []string{"VENDORITO_SOURCE"},
				Required: true,
			},
			&cli.StringFlag{
				Name:     "target",
				Aliases:  []string{"o"},
				Usage:    "Target image path, including tag (if tag is omitted, will match the source tag)",
				EnvVars:  []string{"VENDORITO_TARGET"},
				Required: true,
			},
			&cli.StringFlag{
				Name:     "auth-file",
				Aliases:  []string{"f"},
				Usage:    "Auth file path",
				EnvVars:  []string{"VENDORITO_AUTH_FILE"},
				Required: false,
			},
			&cli.StringFlag{
				Name:     "credentials",
				Aliases:  []string{"k"},
				Usage:    "Credentials in the form of 'domain.tld:username:password', separated by spaces",
				EnvVars:  []string{"VENDORITO_CREDENTIALS"},
				Required: false,
			},
			&cli.Int64Flag{
				Name:     "timeout",
				Usage:    "Maximum time in seconds for the operation, if 0 or not set, no timeout is set",
				Value:    0,
				EnvVars:  []string{"VENDORITO_TIMEOUT"},
				Required: false,
			},
			&cli.IntFlag{
				Name:     "retry-max",
				Usage:    "In case of error, retry the operation this many times, if 0 or not set, no retry is set",
				Value:    0,
				EnvVars:  []string{"VENDORITO_RETRY_MAX"},
				Required: false,
			},
			&cli.Int64Flag{
				Name:     "retry-delay",
				Usage:    "When retrying, wait this many seconds between each attempt",
				Value:    1,
				EnvVars:  []string{"VENDORITO_RETRY_DELAY"},
				Required: false,
			},
		},
		Action: app,
	}

	check(app.Run(os.Args), "Fatal error")
}

func check(err error, format string, args ...interface{}) {
	if err != nil {
		args = append(args, err.Error())
		log.Fatalf(format+": %s", args...)
	}
}

func assert(cond bool, format string, args ...interface{}) {
	if !cond {
		log.Fatalf(format, args...)
	}
}
