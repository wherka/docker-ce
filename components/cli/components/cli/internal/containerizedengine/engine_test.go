package containerizedengine

import (
	"context"
	"fmt"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/errdefs"
	"github.com/containerd/containerd/oci"
	"github.com/docker/docker/api/types"
	"github.com/opencontainers/runtime-spec/specs-go"
	"gotest.tools/assert"
)

func healthfnHappy(ctx context.Context) error {
	return nil
}
func healthfnError(ctx context.Context) error {
	return fmt.Errorf("ping failure")
}

func TestInitGetEngineFail(t *testing.T) {
	ctx := context.Background()
	opts := EngineInitOptions{
		EngineVersion:  "engineversiongoeshere",
		RegistryPrefix: "registryprefixgoeshere",
		ConfigFile:     "/tmp/configfilegoeshere",
		EngineImage:    CommunityEngineImage,
	}
	container := &fakeContainer{}
	client := baseClient{
		cclient: &fakeContainerdClient{
			containersFunc: func(ctx context.Context, filters ...string) ([]containerd.Container, error) {
				return []containerd.Container{container}, nil
			},
		},
	}

	err := client.InitEngine(ctx, opts, &testOutStream{}, &types.AuthConfig{}, healthfnHappy)
	assert.Assert(t, err == ErrEngineAlreadyPresent)
}

func TestInitCheckImageFail(t *testing.T) {
	ctx := context.Background()
	opts := EngineInitOptions{
		EngineVersion:  "engineversiongoeshere",
		RegistryPrefix: "registryprefixgoeshere",
		ConfigFile:     "/tmp/configfilegoeshere",
		EngineImage:    CommunityEngineImage,
	}
	client := baseClient{
		cclient: &fakeContainerdClient{
			containersFunc: func(ctx context.Context, filters ...string) ([]containerd.Container, error) {
				return []containerd.Container{}, nil
			},
			getImageFunc: func(ctx context.Context, ref string) (containerd.Image, error) {
				return nil, fmt.Errorf("something went wrong")

			},
		},
	}

	err := client.InitEngine(ctx, opts, &testOutStream{}, &types.AuthConfig{}, healthfnHappy)
	assert.ErrorContains(t, err, "unable to check for image")
	assert.ErrorContains(t, err, "something went wrong")
}

func TestInitPullFail(t *testing.T) {
	ctx := context.Background()
	opts := EngineInitOptions{
		EngineVersion:  "engineversiongoeshere",
		RegistryPrefix: "registryprefixgoeshere",
		ConfigFile:     "/tmp/configfilegoeshere",
		EngineImage:    CommunityEngineImage,
	}
	client := baseClient{
		cclient: &fakeContainerdClient{
			containersFunc: func(ctx context.Context, filters ...string) ([]containerd.Container, error) {
				return []containerd.Container{}, nil
			},
			getImageFunc: func(ctx context.Context, ref string) (containerd.Image, error) {
				return nil, errdefs.ErrNotFound

			},
			pullFunc: func(ctx context.Context, ref string, opts ...containerd.RemoteOpt) (containerd.Image, error) {
				return nil, fmt.Errorf("pull failure")
			},
		},
	}

	err := client.InitEngine(ctx, opts, &testOutStream{}, &types.AuthConfig{}, healthfnHappy)
	assert.ErrorContains(t, err, "unable to pull image")
	assert.ErrorContains(t, err, "pull failure")
}

func TestInitStartFail(t *testing.T) {
	ctx := context.Background()
	opts := EngineInitOptions{
		EngineVersion:  "engineversiongoeshere",
		RegistryPrefix: "registryprefixgoeshere",
		ConfigFile:     "/tmp/configfilegoeshere",
		EngineImage:    CommunityEngineImage,
	}
	client := baseClient{
		cclient: &fakeContainerdClient{
			containersFunc: func(ctx context.Context, filters ...string) ([]containerd.Container, error) {
				return []containerd.Container{}, nil
			},
			getImageFunc: func(ctx context.Context, ref string) (containerd.Image, error) {
				return nil, errdefs.ErrNotFound

			},
			pullFunc: func(ctx context.Context, ref string, opts ...containerd.RemoteOpt) (containerd.Image, error) {
				return nil, nil
			},
		},
	}

	err := client.InitEngine(ctx, opts, &testOutStream{}, &types.AuthConfig{}, healthfnHappy)
	assert.ErrorContains(t, err, "failed to create docker daemon")
}

func TestGetEngineFail(t *testing.T) {
	ctx := context.Background()
	client := baseClient{
		cclient: &fakeContainerdClient{
			containersFunc: func(ctx context.Context, filters ...string) ([]containerd.Container, error) {
				return nil, fmt.Errorf("container failure")
			},
		},
	}

	_, err := client.GetEngine(ctx)
	assert.ErrorContains(t, err, "failure")
}

func TestGetEngineNotPresent(t *testing.T) {
	ctx := context.Background()
	client := baseClient{
		cclient: &fakeContainerdClient{
			containersFunc: func(ctx context.Context, filters ...string) ([]containerd.Container, error) {
				return []containerd.Container{}, nil
			},
		},
	}

	_, err := client.GetEngine(ctx)
	assert.Assert(t, err == ErrEngineNotPresent)
}

func TestGetEngineFound(t *testing.T) {
	ctx := context.Background()
	container := &fakeContainer{}
	client := baseClient{
		cclient: &fakeContainerdClient{
			containersFunc: func(ctx context.Context, filters ...string) ([]containerd.Container, error) {
				return []containerd.Container{container}, nil
			},
		},
	}

	c, err := client.GetEngine(ctx)
	assert.NilError(t, err)
	assert.Equal(t, c, container)
}

func TestGetEngineImageFail(t *testing.T) {
	client := baseClient{}
	container := &fakeContainer{
		imageFunc: func(context.Context) (containerd.Image, error) {
			return nil, fmt.Errorf("failure")
		},
	}

	_, err := client.getEngineImage(container)
	assert.ErrorContains(t, err, "failure")
}

func TestGetEngineImagePass(t *testing.T) {
	client := baseClient{}
	image := &fakeImage{
		nameFunc: func() string {
			return "imagenamehere"
		},
	}
	container := &fakeContainer{
		imageFunc: func(context.Context) (containerd.Image, error) {
			return image, nil
		},
	}

	name, err := client.getEngineImage(container)
	assert.NilError(t, err)
	assert.Equal(t, name, "imagenamehere")
}

func TestWaitForEngineNeverShowsUp(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	engineWaitInterval = 1 * time.Millisecond
	client := baseClient{
		cclient: &fakeContainerdClient{
			containersFunc: func(ctx context.Context, filters ...string) ([]containerd.Container, error) {
				return []containerd.Container{}, nil
			},
		},
	}

	err := client.waitForEngine(ctx, &testOutStream{}, healthfnError)
	assert.ErrorContains(t, err, "timeout waiting")
}

func TestWaitForEnginePingFail(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	engineWaitInterval = 1 * time.Millisecond
	container := &fakeContainer{}
	client := baseClient{
		cclient: &fakeContainerdClient{
			containersFunc: func(ctx context.Context, filters ...string) ([]containerd.Container, error) {
				return []containerd.Container{container}, nil
			},
		},
	}

	err := client.waitForEngine(ctx, &testOutStream{}, healthfnError)
	assert.ErrorContains(t, err, "ping fail")
}

func TestWaitForEngineHealthy(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	engineWaitInterval = 1 * time.Millisecond
	container := &fakeContainer{}
	client := baseClient{
		cclient: &fakeContainerdClient{
			containersFunc: func(ctx context.Context, filters ...string) ([]containerd.Container, error) {
				return []containerd.Container{container}, nil
			},
		},
	}

	err := client.waitForEngine(ctx, &testOutStream{}, healthfnHappy)
	assert.NilError(t, err)
}

func TestRemoveEngineBadTaskBadDelete(t *testing.T) {
	ctx := context.Background()
	client := baseClient{}
	container := &fakeContainer{
		deleteFunc: func(context.Context, ...containerd.DeleteOpts) error {
			return fmt.Errorf("delete failure")
		},
		taskFunc: func(context.Context, cio.Attach) (containerd.Task, error) {
			return nil, errdefs.ErrNotFound
		},
	}

	err := client.RemoveEngine(ctx, container)
	assert.ErrorContains(t, err, "failed to remove existing engine")
	assert.ErrorContains(t, err, "delete failure")
}

func TestRemoveEngineTaskNoStatus(t *testing.T) {
	ctx := context.Background()
	client := baseClient{}
	task := &fakeTask{
		statusFunc: func(context.Context) (containerd.Status, error) {
			return containerd.Status{}, fmt.Errorf("task status failure")
		},
	}
	container := &fakeContainer{
		taskFunc: func(context.Context, cio.Attach) (containerd.Task, error) {
			return task, nil
		},
	}

	err := client.RemoveEngine(ctx, container)
	assert.ErrorContains(t, err, "task status failure")
}

func TestRemoveEngineTaskNotRunningDeleteFail(t *testing.T) {
	ctx := context.Background()
	client := baseClient{}
	task := &fakeTask{
		statusFunc: func(context.Context) (containerd.Status, error) {
			return containerd.Status{Status: containerd.Unknown}, nil
		},
		deleteFunc: func(context.Context, ...containerd.ProcessDeleteOpts) (*containerd.ExitStatus, error) {
			return nil, fmt.Errorf("task delete failure")
		},
	}
	container := &fakeContainer{
		taskFunc: func(context.Context, cio.Attach) (containerd.Task, error) {
			return task, nil
		},
	}

	err := client.RemoveEngine(ctx, container)
	assert.ErrorContains(t, err, "task delete failure")
}

func TestRemoveEngineTaskRunningKillFail(t *testing.T) {
	ctx := context.Background()
	client := baseClient{}
	task := &fakeTask{
		statusFunc: func(context.Context) (containerd.Status, error) {
			return containerd.Status{Status: containerd.Running}, nil
		},
		killFunc: func(context.Context, syscall.Signal, ...containerd.KillOpts) error {
			return fmt.Errorf("task kill failure")
		},
	}
	container := &fakeContainer{
		taskFunc: func(context.Context, cio.Attach) (containerd.Task, error) {
			return task, nil
		},
	}

	err := client.RemoveEngine(ctx, container)
	assert.ErrorContains(t, err, "task kill failure")
}

func TestRemoveEngineTaskRunningWaitFail(t *testing.T) {
	ctx := context.Background()
	client := baseClient{}
	task := &fakeTask{
		statusFunc: func(context.Context) (containerd.Status, error) {
			return containerd.Status{Status: containerd.Running}, nil
		},
		waitFunc: func(context.Context) (<-chan containerd.ExitStatus, error) {
			return nil, fmt.Errorf("task wait failure")
		},
	}
	container := &fakeContainer{
		taskFunc: func(context.Context, cio.Attach) (containerd.Task, error) {
			return task, nil
		},
	}

	err := client.RemoveEngine(ctx, container)
	assert.ErrorContains(t, err, "task wait failure")
}

func TestRemoveEngineTaskRunningHappyPath(t *testing.T) {
	ctx := context.Background()
	client := baseClient{}
	ch := make(chan containerd.ExitStatus, 1)
	task := &fakeTask{
		statusFunc: func(context.Context) (containerd.Status, error) {
			return containerd.Status{Status: containerd.Running}, nil
		},
		waitFunc: func(context.Context) (<-chan containerd.ExitStatus, error) {
			ch <- containerd.ExitStatus{}
			return ch, nil
		},
	}
	container := &fakeContainer{
		taskFunc: func(context.Context, cio.Attach) (containerd.Task, error) {
			return task, nil
		},
	}

	err := client.RemoveEngine(ctx, container)
	assert.NilError(t, err)
}

func TestRemoveEngineTaskKillTimeout(t *testing.T) {
	ctx := context.Background()
	ch := make(chan containerd.ExitStatus, 1)
	client := baseClient{}
	engineWaitTimeout = 10 * time.Millisecond
	task := &fakeTask{
		statusFunc: func(context.Context) (containerd.Status, error) {
			return containerd.Status{Status: containerd.Running}, nil
		},
		waitFunc: func(context.Context) (<-chan containerd.ExitStatus, error) {
			//ch <- containerd.ExitStatus{} // let it timeout
			return ch, nil
		},
	}
	container := &fakeContainer{
		taskFunc: func(context.Context, cio.Attach) (containerd.Task, error) {
			return task, nil
		},
	}

	err := client.RemoveEngine(ctx, container)
	assert.Assert(t, err == ErrEngineShutdownTimeout)
}

func TestStartEngineOnContainerdImageErr(t *testing.T) {
	ctx := context.Background()
	imageName := "testnamegoeshere"
	configFile := "/tmp/configfilegoeshere"
	client := baseClient{
		cclient: &fakeContainerdClient{
			getImageFunc: func(ctx context.Context, ref string) (containerd.Image, error) {
				return nil, fmt.Errorf("some image lookup failure")

			},
		},
	}
	err := client.startEngineOnContainerd(ctx, imageName, configFile)
	assert.ErrorContains(t, err, "some image lookup failure")
}

func TestStartEngineOnContainerdImageNotFound(t *testing.T) {
	ctx := context.Background()
	imageName := "testnamegoeshere"
	configFile := "/tmp/configfilegoeshere"
	client := baseClient{
		cclient: &fakeContainerdClient{
			getImageFunc: func(ctx context.Context, ref string) (containerd.Image, error) {
				return nil, errdefs.ErrNotFound

			},
		},
	}
	err := client.startEngineOnContainerd(ctx, imageName, configFile)
	assert.ErrorContains(t, err, "engine image missing")
}

func TestStartEngineOnContainerdHappy(t *testing.T) {
	ctx := context.Background()
	imageName := "testnamegoeshere"
	configFile := "/tmp/configfilegoeshere"
	ch := make(chan containerd.ExitStatus, 1)
	streams := cio.Streams{}
	task := &fakeTask{
		statusFunc: func(context.Context) (containerd.Status, error) {
			return containerd.Status{Status: containerd.Running}, nil
		},
		waitFunc: func(context.Context) (<-chan containerd.ExitStatus, error) {
			ch <- containerd.ExitStatus{}
			return ch, nil
		},
	}
	container := &fakeContainer{
		newTaskFunc: func(ctx context.Context, creator cio.Creator, opts ...containerd.NewTaskOpts) (containerd.Task, error) {
			if streams.Stdout != nil {
				streams.Stdout.Write([]byte("{}"))
			}
			return task, nil
		},
	}
	client := baseClient{
		cclient: &fakeContainerdClient{
			getImageFunc: func(ctx context.Context, ref string) (containerd.Image, error) {
				return nil, nil

			},
			newContainerFunc: func(ctx context.Context, id string, opts ...containerd.NewContainerOpts) (containerd.Container, error) {
				return container, nil
			},
		},
	}
	err := client.startEngineOnContainerd(ctx, imageName, configFile)
	assert.NilError(t, err)
}

func TestGetEngineConfigFilePathBadSpec(t *testing.T) {
	ctx := context.Background()
	client := baseClient{}
	container := &fakeContainer{
		specFunc: func(context.Context) (*oci.Spec, error) {
			return nil, fmt.Errorf("spec error")
		},
	}
	_, err := client.getEngineConfigFilePath(ctx, container)
	assert.ErrorContains(t, err, "spec error")
}

func TestGetEngineConfigFilePathDistinct(t *testing.T) {
	ctx := context.Background()
	client := baseClient{}
	container := &fakeContainer{
		specFunc: func(context.Context) (*oci.Spec, error) {
			return &oci.Spec{
				Process: &specs.Process{
					Args: []string{
						"--another-flag",
						"foo",
						"--config-file",
						"configpath",
					},
				},
			}, nil
		},
	}
	configFile, err := client.getEngineConfigFilePath(ctx, container)
	assert.NilError(t, err)
	assert.Assert(t, err, configFile == "configpath")
}

func TestGetEngineConfigFilePathEquals(t *testing.T) {
	ctx := context.Background()
	client := baseClient{}
	container := &fakeContainer{
		specFunc: func(context.Context) (*oci.Spec, error) {
			return &oci.Spec{
				Process: &specs.Process{
					Args: []string{
						"--another-flag=foo",
						"--config-file=configpath",
					},
				},
			}, nil
		},
	}
	configFile, err := client.getEngineConfigFilePath(ctx, container)
	assert.NilError(t, err)
	assert.Assert(t, err, configFile == "configpath")
}

func TestGetEngineConfigFilePathMalformed1(t *testing.T) {
	ctx := context.Background()
	client := baseClient{}
	container := &fakeContainer{
		specFunc: func(context.Context) (*oci.Spec, error) {
			return &oci.Spec{
				Process: &specs.Process{
					Args: []string{
						"--another-flag",
						"--config-file",
					},
				},
			}, nil
		},
	}
	_, err := client.getEngineConfigFilePath(ctx, container)
	assert.Assert(t, err == ErrMalformedConfigFileParam)
}

// getEngineConfigFilePath will extract the config file location from the engine flags
func (c baseClient) getEngineConfigFilePath(ctx context.Context, engine containerd.Container) (string, error) {
	spec, err := engine.Spec(ctx)
	configFile := ""
	if err != nil {
		return configFile, err
	}
	for i := 0; i < len(spec.Process.Args); i++ {
		arg := spec.Process.Args[i]
		if strings.HasPrefix(arg, "--config-file") {
			if strings.Contains(arg, "=") {
				split := strings.SplitN(arg, "=", 2)
				configFile = split[1]
			} else {
				if i+1 >= len(spec.Process.Args) {
					return configFile, ErrMalformedConfigFileParam
				}
				configFile = spec.Process.Args[i+1]
			}
		}
	}

	if configFile == "" {
		// TODO - any more diagnostics to offer?
		return configFile, ErrEngineConfigLookupFailure
	}
	return configFile, nil
}
