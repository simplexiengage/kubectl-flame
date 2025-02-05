package profiler

import (
	"fmt"

	"github.com/simplexiengage/kubectl-flame/agent/details"
	"github.com/simplexiengage/kubectl-flame/api"
)

type FlameGraphProfiler interface {
	SetUp(job *details.ProfilingJob) error
	Invoke(job *details.ProfilingJob) error
}

var (
	jvm    = JvmProfiler{}
	bpf    = BpfProfiler{}
	python = PythonProfiler{}
	ruby   = RubyProfiler{}
	perf   = PerfProfiler{}
)

func ForLanguage(lang api.ProgrammingLanguage) (FlameGraphProfiler, error) {
	switch lang {
	case api.Java:
		return &jvm, nil
	case api.Go:
		return &bpf, nil
	case api.Python:
		return &python, nil
	case api.Ruby:
		return &ruby, nil
	case api.Node:
		return &perf, nil
	default:
		return nil, fmt.Errorf("could not find profiler for language %s", lang)
	}
}
