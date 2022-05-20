package registry

import (
	"context"
	"fmt"
	"time"

	logrus_bugsnag "github.com/Shopify/logrus-bugsnag"
	logstash "github.com/bshuster-repo/logrus-logstash-hook"
	bugsnag "github.com/bugsnag/bugsnag-go"
	"github.com/distribution/distribution/v3/configuration"
	dcontext "github.com/distribution/distribution/v3/context"
	"github.com/sirupsen/logrus"
)

// https://github.com/distribution/distribution/blob/4363fb1ef4676df2b9d99e3630e1b568141597c4/registry/registry.go#L342-L390
// configureLogging prepares the context with a logger using the
// configuration.
func ConfigureLogging(ctx context.Context, config *configuration.Configuration) (context.Context, error) {
	logrus.SetLevel(logLevel(config.Log.Level))

	formatter := config.Log.Formatter
	if formatter == "" {
		formatter = "text" // default formatter
	}

	switch formatter {
	case "json":
		logrus.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat:   time.RFC3339Nano,
			DisableHTMLEscape: true,
		})
	case "text":
		logrus.SetFormatter(&logrus.TextFormatter{
			TimestampFormat: time.RFC3339Nano,
		})
	case "logstash":
		logrus.SetFormatter(&logstash.LogstashFormatter{
			Formatter: &logrus.JSONFormatter{TimestampFormat: time.RFC3339Nano},
		})
	default:
		// just let the library use default on empty string.
		if config.Log.Formatter != "" {
			return ctx, fmt.Errorf("unsupported logging formatter: %q", config.Log.Formatter)
		}
	}

	if config.Log.Formatter != "" {
		logrus.Debugf("using %q logging formatter", config.Log.Formatter)
	}

	if len(config.Log.Fields) > 0 {
		// build up the static fields, if present.
		var fields []interface{}
		for k := range config.Log.Fields {
			fields = append(fields, k)
		}

		ctx = dcontext.WithValues(ctx, config.Log.Fields)
		ctx = dcontext.WithLogger(ctx, dcontext.GetLogger(ctx, fields...))
	}

	dcontext.SetDefaultLogger(dcontext.GetLogger(ctx))
	return ctx, nil
}

// https://github.com/distribution/distribution/blob/4363fb1ef4676df2b9d99e3630e1b568141597c4/registry/registry.go#L392-L400
func logLevel(level configuration.Loglevel) logrus.Level {
	l, err := logrus.ParseLevel(string(level))
	if err != nil {
		l = logrus.InfoLevel
		logrus.Warnf("error parsing level %q: %v, using %q	", level, err, l)
	}

	return l
}

// https://github.com/distribution/distribution/blob/4363fb1ef4676df2b9d99e3630e1b568141597c4/registry/registry.go#L402-L426
// configureBugsnag configures bugsnag reporting, if enabled
func ConfigureBugsnag(config *configuration.Configuration) {
	if config.Reporting.Bugsnag.APIKey == "" {
		return
	}

	bugsnagConfig := bugsnag.Configuration{
		APIKey: config.Reporting.Bugsnag.APIKey,
	}
	if config.Reporting.Bugsnag.ReleaseStage != "" {
		bugsnagConfig.ReleaseStage = config.Reporting.Bugsnag.ReleaseStage
	}
	if config.Reporting.Bugsnag.Endpoint != "" {
		bugsnagConfig.Endpoint = config.Reporting.Bugsnag.Endpoint
	}
	bugsnag.Configure(bugsnagConfig)

	// configure logrus bugsnag hook
	hook, err := logrus_bugsnag.NewBugsnagHook()
	if err != nil {
		logrus.Fatalln(err)
	}

	logrus.AddHook(hook)
}