package runner

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckContainerOptions(t *testing.T) {
	for _, ok := range []string{
		"",
		"--health-cmd pg_isready --health-interval 10s --health-timeout 5s --health-retries 5",
		`--health-cmd "redis-cli ping" --user root -e FOO=bar -eBAR=baz --cpus=2 -it`,
		"--memory 2g --shm-size=1g --tmpfs /tmp:rw --entrypoint /bin/sh -w /src",
	} {
		assert.NoError(t, checkContainerOptions(ok), ok)
	}
	for _, bad := range []string{
		"--privileged",
		"--user root --privileged",
		"-v /:/host",
		"--volume=/var/run/docker.sock:/var/run/docker.sock",
		"--mount type=bind,src=/,dst=/host",
		"--network host",
		"--pid=host",
		"--cap-add SYS_ADMIN",
		"--device /dev/kvm",
		"--security-opt seccomp=unconfined",
		"--env-file /state/.runner",
		"--label-file /etc/shadow",
		"--volumes-from other",
		"--name other",
		"-p 80:80",
		"-itv /:/host",
		"--health-cmd --privileged",
		"-",
		"--",
		`"unterminated`,
	} {
		assert.Error(t, checkContainerOptions(bad), bad)
	}
}

func TestCheckContainerVolumes(t *testing.T) {
	assert.NoError(t, checkContainerVolumes([]string{"/data", "/cache"}))
	assert.Error(t, checkContainerVolumes([]string{"/:/host"}))
	assert.Error(t, checkContainerVolumes([]string{"shared:/cache"}))
	assert.Error(t, checkContainerVolumes([]string{"/var/run/docker.sock:/var/run/docker.sock:ro"}))
}

func TestJoinOptions(t *testing.T) {
	assert.Equal(t, "--user root --memory 1g", joinOptions("--user root", "--memory 1g"))
	assert.Equal(t, "--memory 1g", joinOptions("", "--memory 1g"))
	assert.Equal(t, "--user root", joinOptions("--user root", ""))
}
