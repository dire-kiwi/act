package runner

import (
	"context"
	"fmt"
	"strings"

	"github.com/kballard/go-shellquote"
)

// Docker create options a workflow may give its job and service containers
// under Config.RestrictContainers. They only change the container itself:
// none of them reaches the host's files, devices, namespaces or privileges,
// or reads files on the runner's machine (as --env-file and --label-file do).
var safeContainerFlags = map[string]bool{
	"add-host": true, "cpu-period": true, "cpu-quota": true, "cpu-shares": true, "cpus": true,
	"dns": true, "dns-option": true, "dns-search": true, "domainname": true, "entrypoint": true,
	"env": true, "expose": true, "health-cmd": true, "health-interval": true, "health-retries": true,
	"health-start-interval": true, "health-start-period": true, "health-timeout": true, "hostname": true,
	"init": true, "interactive": true, "label": true, "memory": true, "memory-reservation": true,
	"memory-swap": true, "memory-swappiness": true, "network-alias": true, "no-healthcheck": true,
	"pids-limit": true, "read-only": true, "rm": true, "shm-size": true, "stop-signal": true,
	"stop-timeout": true, "tmpfs": true, "tty": true, "ulimit": true, "user": true, "workdir": true,
}

// Their one-letter forms. The ones that take a value take the rest of the
// argument as it (-eFOO=bar); -i and -t may be combined.
const (
	safeShortValueFlags = "ehlmuw"
	safeShortBoolFlags  = "it"
)

// checkContainerOptions refuses options outside safeContainerFlags. Every
// argument that looks like a flag is checked, values included, so a value
// can't hide a flag from the check; such values need --flag=value.
func checkContainerOptions(options string) error {
	args, err := shellquote.Split(options)
	if err != nil {
		return fmt.Errorf("cannot split container options %q: %w", options, err)
	}
	for _, arg := range args {
		if !strings.HasPrefix(arg, "-") {
			continue
		}
		if name, ok := strings.CutPrefix(arg, "--"); ok {
			name, _, _ = strings.Cut(name, "=")
			if !safeContainerFlags[name] {
				return fmt.Errorf("container option %q is not allowed on this runner", arg)
			}
			continue
		}
		short := arg[1:]
		if short == "" {
			return fmt.Errorf("container option %q is not allowed on this runner", arg)
		}
		if strings.ContainsRune(safeShortValueFlags, rune(short[0])) {
			continue
		}
		for _, c := range short {
			if !strings.ContainsRune(safeShortBoolFlags, c) {
				return fmt.Errorf("container option %q is not allowed on this runner", arg)
			}
		}
	}
	return nil
}

// checkContainerVolumes allows only anonymous volumes ("/data"): a host path
// reaches outside the container and a named volume outlives the job.
func checkContainerVolumes(volumes []string) error {
	for _, v := range volumes {
		if strings.Contains(v, ":") {
			return fmt.Errorf("container volume %q is not allowed on this runner: only anonymous volumes are", v)
		}
	}
	return nil
}

// checkContainers applies Config.RestrictContainers to the job's container
// and services, as startJobContainer will interpolate them.
func (rc *RunContext) checkContainers(ctx context.Context) error {
	job := rc.Run.Job()
	if c := job.Container(); c != nil {
		if err := checkContainerOptions(rc.ExprEval.Interpolate(ctx, c.Options)); err != nil {
			return err
		}
		if err := checkContainerVolumes(c.Volumes); err != nil {
			return err
		}
	}
	for id, spec := range job.Services {
		if err := checkContainerOptions(rc.ExprEval.Interpolate(ctx, spec.Options)); err != nil {
			return fmt.Errorf("service %s: %w", id, err)
		}
		volumes := make([]string, 0, len(spec.Volumes))
		for _, v := range spec.Volumes {
			volumes = append(volumes, rc.ExprEval.Interpolate(ctx, v))
		}
		if err := checkContainerVolumes(volumes); err != nil {
			return fmt.Errorf("service %s: %w", id, err)
		}
	}
	return nil
}
