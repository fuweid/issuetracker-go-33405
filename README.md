## [golang/go#33405](https://github.com/golang/go/issues/33405): fd leaking

### background

I run containerd to manage linux containers. One container uses snapshotter
provided by containerd as rootfs and the files on the snapshotter will be
cleanup by containerd if the container has been deleted. The cleanup action
is called as **snapshotter gc**.

If the container inits during snapshotter gc, we found that some
containerd-shim process, which manages container's life cycle, will contain
the fd opened by parent - containerd process.

### caused by `os.RemoveAll` function

The `os.RemoveAll` will be called during the snapshotter gc and the
`os.RemoveAll` does `open` and `unlink` action to remove file and dir.

Basically, the golang runtime will add `O_CLOEXEC` flag when the user calls
any `Open*` functions provided golang builtin package, for example, `os.Open`.
`O_CLOEXEC` flag can prevent fd leaking to child processes.

Some platforms doesn't support `O_CLOEXEC` flag. golang runtime will call
fnctl to set `FD_CLOEXEC` to make sure there is no fd leaking. But it is not
atomic. More information is [here](https://docs.fedoraproject.org/en-US/Fedora_Security_Team/1/html/Defensive_Coding/sect-Defensive_Coding-Tasks-Descriptors-Child_Processes.html)

golang 1.12 refactors `os.RemoveAll` function and but forgot to add
`O_CLOEXEC` flag.

```
func openFdAt(dirfd int, name string) (*File, error) {
	var r int
	for {
		var e error
		r, e = unix.Openat(dirfd, name, O_RDONLY, 0)
		if e == nil {
			break
		}

		// See comment in openFileNolog.
		if runtime.GOOS == "darwin" && e == syscall.EINTR {
			continue
		}

		return nil, e
	}

	if !supportsCloseOnExec {
		syscall.CloseOnExec(r)
	}

	return newFile(uintptr(r), name, kindOpenFile), nil
}
```

The opened fd will be closed when the dir has been removed. If the child
process is created between `openFdAt` and `close`, the child process will
get the leaking fds.

### reproduce

In order to reproduce this issue easier, I will create dir containing >100
levels to let `os.RemoveAll` call maintain opened fd longer.

Using following script to reproduce the issue.

```
$ make clean
$ make build
$ make test
```

But please make sure that your golang version is <= 1.12.7 and > 1.11.

### status

golang community has fixed this issue! Please update to 1.12.8 if there is new
release.
