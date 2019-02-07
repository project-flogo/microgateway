package zapsyslogcore

import (
	"log"
	"log/syslog"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/project-flogo/core/support/log/zapconfig"
	"github.com/project-flogo/core/support/log/zapcores"
	"go.uber.org/zap/zapcore"
)

func init() {
	trcLogr, logr := zapSysLogCore()
	zapcores.RegisterLogCore("zapsyslogcore", logr)
	zapcores.RegisterTraceLogCore("zapsyslogtracecore", trcLogr)
}

// this code is taken from github.com/tchap/zapext. Modified it accoriding to the needs.

// SysLogCore root struct for zap core user friendly syslog.
type SysLogCore struct {
	zapcore.LevelEnabler
	encoder zapcore.Encoder
	writer  *syslog.Writer
}

func newSyslogCore(enab zapcore.LevelEnabler, encoder zapcore.Encoder, writer *syslog.Writer) *SysLogCore {
	return &SysLogCore{
		LevelEnabler: enab,
		encoder:      encoder,
		writer:       writer,
	}
}

// With one of the zap core interface methods implimented here
func (core *SysLogCore) With(fields []zapcore.Field) zapcore.Core {
	clone := core.clone()
	for _, field := range fields {
		field.AddTo(clone.encoder)
	}
	return clone
}

// Check one of the zap core interface methods implimented here
func (core *SysLogCore) Check(entry zapcore.Entry, checked *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if core.Enabled(entry.Level) {
		return checked.AddCore(entry, core)
	}
	return checked
}

// Write one of the core interface methods implimented here
func (core *SysLogCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	// Generate the message.
	buffer, err := core.encoder.EncodeEntry(entry, fields)
	if err != nil {
		return errors.Wrap(err, "failed to encode log entry")
	}

	//message := fmt.Sprintf("[metadata={process='microgateway',function='microgateway',TMG_CLUSTER_NAME='%s',TMG_ZONE_NAME='%s',POD_IP='%s'}] ", os.Getenv("TMG_CLUSTER_NAME"), os.Getenv("TMG_ZONE_NAME"), os.Getenv("POD_IP")) + buffer.String()
	message := buffer.String()

	// Write the message.
	switch entry.Level {
	case zapcore.DebugLevel:
		return core.writer.Debug(message)

	case zapcore.InfoLevel:
		return core.writer.Info(message)

	case zapcore.WarnLevel:
		return core.writer.Warning(message)

	case zapcore.ErrorLevel:
		return core.writer.Err(message)

	case zapcore.DPanicLevel:
		return core.writer.Crit(message)

	case zapcore.PanicLevel:
		return core.writer.Crit(message)

	case zapcore.FatalLevel:
		return core.writer.Crit(message)

	default:
		return errors.Errorf("unknown log level: %v", entry.Level)
	}
}

// Sync one of the core interface methods implimented here
func (core *SysLogCore) Sync() error {
	return nil
}

func (core *SysLogCore) clone() *SysLogCore {
	return &SysLogCore{
		LevelEnabler: core.LevelEnabler,
		encoder:      core.encoder.Clone(),
		writer:       core.writer,
	}
}

// zapSysLogCore returns zapcore.core impl for syslog
func zapSysLogCore() (*SysLogCore, *SysLogCore) {

	var enc, traceEnc zapcore.Encoder

	envLogFormat := strings.ToUpper(os.Getenv("FLOGO_LOG_FORMAT"))
	if strings.Compare(envLogFormat, "JSON") != 0 {
		enc = zapcore.NewConsoleEncoder(zapconfig.DefaultCfg().LogCfg().EncoderConfig)
		traceEnc = zapcore.NewConsoleEncoder(zapconfig.DefaultCfg().TraceLogCfg().EncoderConfig)
	} else {
		enc = zapcore.NewJSONEncoder(zapconfig.DefaultCfg().LogCfg().EncoderConfig)
		traceEnc = zapcore.NewJSONEncoder(zapconfig.DefaultCfg().TraceLogCfg().EncoderConfig)
	}

	envSysLogTag := strings.ToUpper(os.Getenv("MICROGATEWAY_SYSLOG_TAG"))
	if len(envSysLogTag) == 0 {
		envSysLogTag = "zapsyslogtag"
	}

	// Initialize syslog writer.
	writer, err := syslog.New(syslog.LOG_INFO, envSysLogTag)
	if err != nil {
		log.Fatal("failed to set up syslog")
	}

	return newSyslogCore(zapconfig.DefaultCfg().TraceLogLvl(), traceEnc, writer), newSyslogCore(zapconfig.DefaultCfg().LogLvl(), enc, writer)
}
